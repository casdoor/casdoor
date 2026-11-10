// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package hsm signs with private keys kept in a hardware security module through PKCS #11. The vendor library is
// loaded at runtime without cgo.
package hsm

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/asn1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"sync"
	"unsafe"

	"github.com/casdoor/casdoor/conf"
)

const TypePkcs11 = "pkcs11"

const ptrSize = int(unsafe.Sizeof(uintptr(0)))

// Config is read from app.conf (or the environment), the PIN never goes to the database
type Config struct {
	Type       string
	Library    string
	TokenLabel string
	Slot       string
	Pin        string
}

func GetConfig() Config {
	return Config{
		Type:       strings.ToLower(strings.TrimSpace(conf.GetConfigString("hsmType"))),
		Library:    strings.TrimSpace(conf.GetConfigString("hsmLibrary")),
		TokenLabel: strings.TrimSpace(conf.GetConfigString("hsmTokenLabel")),
		Slot:       strings.TrimSpace(conf.GetConfigString("hsmSlot")),
		Pin:        conf.GetConfigString("hsmPin"),
	}
}

// IsConfigured tells whether app.conf sets an HSM
func IsConfigured() bool {
	return GetConfig().Type != ""
}

type module interface {
	// loadKey finds a private key by its reference and returns its public key
	loadKey(ref string) (crypto.PublicKey, error)
	// sign signs a digest prepared for the key type: the DigestInfo for RSA PKCS #1 v1.5, the hash for ECDSA
	sign(ref string, public crypto.PublicKey, digest []byte, opts crypto.SignerOpts) ([]byte, error)
	describe() string
	// forget drops what is cached about a key
	forget(ref string)
}

var (
	moduleMu      sync.Mutex
	currentModule module
	signers       = map[string]*Signer{}
)

func getModule() (module, error) {
	moduleMu.Lock()
	defer moduleMu.Unlock()
	if currentModule != nil {
		return currentModule, nil
	}

	config := GetConfig()
	if config.Library == "" {
		return nil, errors.New("hsmLibrary is not set in app.conf")
	}

	var res module
	var err error
	switch config.Type {
	case TypePkcs11:
		res, err = newPkcs11Module(config)
	case "":
		return nil, errors.New("no HSM is configured, set hsmType and hsmLibrary in app.conf")
	default:
		return nil, fmt.Errorf("unknown hsmType: %s, it should be pkcs11", config.Type)
	}
	if err != nil {
		return nil, err
	}
	currentModule = res
	return res, nil
}

// Signer is a crypto.Signer whose private key stays in the HSM
type Signer struct {
	ref    string
	public crypto.PublicKey
	module module
}

// GetSigner returns the signer of the key with the reference: a PKCS #11 CKA_LABEL (or "id:<hex CKA_ID>")
func GetSigner(ref string) (*Signer, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, errors.New("the HSM key reference is empty")
	}

	m, err := getModule()
	if err != nil {
		return nil, err
	}

	moduleMu.Lock()
	signer, ok := signers[ref]
	moduleMu.Unlock()
	if ok {
		return signer, nil
	}

	public, err := m.loadKey(ref)
	if err != nil {
		return nil, err
	}

	signer = &Signer{ref: ref, public: public, module: m}
	moduleMu.Lock()
	signers[ref] = signer
	moduleMu.Unlock()
	return signer, nil
}

// ReloadSigner looks the key up in the HSM again, e.g. after it was replaced under the same reference
func ReloadSigner(ref string) (*Signer, error) {
	ref = strings.TrimSpace(ref)
	moduleMu.Lock()
	delete(signers, ref)
	m := currentModule
	moduleMu.Unlock()
	if m != nil {
		m.forget(ref)
	}
	return GetSigner(ref)
}

// IsSigner tells whether a key is an HSM signer
func IsSigner(key interface{}) bool {
	_, ok := key.(*Signer)
	return ok
}

func (s *Signer) Public() crypto.PublicKey {
	return s.public
}

func (s *Signer) Reference() string {
	return s.ref
}

// Sign follows crypto.Signer: digest is hashed by opts.HashFunc(), with *rsa.PSSOptions for RSA-PSS. ECDSA
// signatures are returned in ASN.1 DER.
func (s *Signer) Sign(_ io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	if opts == nil {
		opts = crypto.Hash(0)
	}

	switch public := s.public.(type) {
	case *rsa.PublicKey:
		return s.module.sign(s.ref, public, digest, opts)
	case *ecdsa.PublicKey:
		if opts.HashFunc() != 0 && len(digest) != opts.HashFunc().Size() {
			return nil, errors.New("the digest length doesn't match the hash")
		}
		return s.module.sign(s.ref, public, digest, opts)
	}
	return nil, fmt.Errorf("unsupported HSM key type: %T", s.public)
}

// Status describes the configured HSM, connecting to it when asked
type Status struct {
	Type       string `json:"type"`
	Library    string `json:"library"`
	Token      string `json:"token"`
	Configured bool   `json:"configured"`
	Connected  bool   `json:"connected"`
	Device     string `json:"device"`
	Error      string `json:"error"`
}

func GetStatus(connect bool) *Status {
	config := GetConfig()
	res := &Status{Type: config.Type, Library: config.Library, Configured: config.Type != ""}
	if config.TokenLabel != "" {
		res.Token = config.TokenLabel
	} else if config.Slot != "" {
		res.Token = "slot " + config.Slot
	}
	if !res.Configured || !connect {
		return res
	}

	m, err := getModule()
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Connected = true
	res.Device = m.describe()
	return res
}

// cMem allocates memory that the C code may write to; it is returned to stay on the heap, where it doesn't move
//
//go:noinline
func cMem(n int) []byte {
	if n <= 0 {
		n = 1
	}
	return make([]byte, n)
}

func cBytes(data []byte) []byte {
	res := cMem(len(data) + 1)
	copy(res, data)
	return res
}

func addr(b []byte) uintptr {
	if len(b) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&b[0]))
}

// cStruct lays out a C struct of CK_ULONG and pointer fields, packed on Windows and naturally aligned elsewhere,
// where every field is 8 bytes on the supported 64-bit systems
type cStruct struct {
	buf []byte
	off int
}

func newCStruct(size int) *cStruct {
	return &cStruct{buf: cMem(size)}
}

func (s *cStruct) ulong(v uint64) *cStruct {
	putUlong(s.buf[s.off:], v)
	s.off += ulongSize
	return s
}

func (s *cStruct) ptr(p uintptr) *cStruct {
	if ptrSize == 8 {
		binary.NativeEndian.PutUint64(s.buf[s.off:], uint64(p))
	} else {
		binary.NativeEndian.PutUint32(s.buf[s.off:], uint32(p))
	}
	s.off += ptrSize
	return s
}

func putUlong(b []byte, v uint64) {
	if ulongSize == 8 {
		binary.NativeEndian.PutUint64(b, v)
	} else {
		binary.NativeEndian.PutUint32(b, uint32(v))
	}
}

func getUlong(b []byte) uint64 {
	if ulongSize == 8 {
		return binary.NativeEndian.Uint64(b)
	}
	return uint64(binary.NativeEndian.Uint32(b))
}

func getPtr(b []byte) uintptr {
	if ptrSize == 8 {
		return uintptr(binary.NativeEndian.Uint64(b))
	}
	return uintptr(binary.NativeEndian.Uint32(b))
}

func marshalEcSignature(r, s []byte) ([]byte, error) {
	return asn1.Marshal(struct {
		R, S *big.Int
	}{new(big.Int).SetBytes(r), new(big.Int).SetBytes(s)})
}

func parseEcSignature(der []byte) (*big.Int, *big.Int, error) {
	var sig struct {
		R, S *big.Int
	}
	rest, err := asn1.Unmarshal(der, &sig)
	if err != nil || len(rest) != 0 || sig.R == nil || sig.S == nil {
		return nil, nil, errors.New("invalid ECDSA signature")
	}
	return sig.R, sig.S, nil
}
