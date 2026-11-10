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

package hsm

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

// the indexes of the functions in CK_FUNCTION_LIST (PKCS #11 v2.x)
const (
	fnInitialize        = 0
	fnGetSlotList       = 4
	fnGetTokenInfo      = 6
	fnOpenSession       = 12
	fnCloseSession      = 13
	fnLogin             = 18
	fnGetAttributeValue = 24
	fnFindObjectsInit   = 26
	fnFindObjects       = 27
	fnFindObjectsFinal  = 28
	fnSignInit          = 42
	fnSign              = 43
	fnCount             = 44
)

const (
	ckrOk                       = 0x0
	ckrDeviceError              = 0x30
	ckrDeviceRemoved            = 0x32
	ckrSessionClosed            = 0xB0
	ckrSessionHandleInvalid     = 0xB3
	ckrTokenNotPresent          = 0xE0
	ckrUserAlreadyLoggedIn      = 0x100
	ckrUserNotLoggedIn          = 0x101
	ckrBufferTooSmall           = 0x150
	ckrCryptokiNotInitialized   = 0x190
	ckrCryptokiAlreadyInitiated = 0x191

	ckfOsLockingOk    = 0x2
	ckfSerialSession  = 0x4
	ckuUser           = 1
	ckoPublicKey      = 2
	ckoPrivateKey     = 3
	ckkRsa            = 0x0
	ckaClass          = 0x0
	ckaLabel          = 0x3
	ckaKeyType        = 0x100
	ckaId             = 0x102
	ckaModulus        = 0x120
	ckaPublicExponent = 0x122
	ckaEcParams       = 0x180
	ckaEcPoint        = 0x181

	ckmRsaPkcs    = 0x1
	ckmRsaPkcsPss = 0xD
	ckmEcdsa      = 0x1041
	ckmSha1       = 0x220
	ckmSha224     = 0x255
	ckmSha256     = 0x250
	ckmSha384     = 0x260
	ckmSha512     = 0x270
	ckgMgf1Sha1   = 0x1
	ckgMgf1Sha256 = 0x2
	ckgMgf1Sha384 = 0x3
	ckgMgf1Sha512 = 0x4
	ckgMgf1Sha224 = 0x5
)

var ckrNames = map[uint64]string{
	0x5: "CKR_GENERAL_ERROR", 0x6: "CKR_FUNCTION_FAILED", 0x7: "CKR_ARGUMENTS_BAD", 0x11: "CKR_ATTRIBUTE_SENSITIVE",
	0x12: "CKR_ATTRIBUTE_TYPE_INVALID", 0x30: "CKR_DEVICE_ERROR", 0x31: "CKR_DEVICE_MEMORY", 0x32: "CKR_DEVICE_REMOVED",
	0x60: "CKR_KEY_HANDLE_INVALID", 0x63: "CKR_KEY_TYPE_INCONSISTENT", 0x68: "CKR_KEY_FUNCTION_NOT_PERMITTED",
	0x70: "CKR_MECHANISM_INVALID", 0x71: "CKR_MECHANISM_PARAM_INVALID", 0x82: "CKR_OBJECT_HANDLE_INVALID",
	0xA0: "CKR_PIN_INCORRECT", 0xA4: "CKR_PIN_LOCKED", 0xB0: "CKR_SESSION_CLOSED", 0xB3: "CKR_SESSION_HANDLE_INVALID",
	0xE0: "CKR_TOKEN_NOT_PRESENT", 0x100: "CKR_USER_ALREADY_LOGGED_IN", 0x101: "CKR_USER_NOT_LOGGED_IN",
	0x150: "CKR_BUFFER_TOO_SMALL", 0x190: "CKR_CRYPTOKI_NOT_INITIALIZED",
}

type ckError struct {
	function string
	rv       uint64
}

func (e *ckError) Error() string {
	name := ckrNames[e.rv]
	if name == "" {
		name = fmt.Sprintf("0x%X", e.rv)
	}
	return fmt.Sprintf("PKCS #11 %s failed: %s", e.function, name)
}

// isSessionLost tells whether the session should be opened again, e.g. after the HSM restarted
func isSessionLost(err error) bool {
	var ckErr *ckError
	if !errors.As(err, &ckErr) {
		return false
	}
	switch ckErr.rv {
	case ckrSessionClosed, ckrSessionHandleInvalid, ckrUserNotLoggedIn, ckrDeviceRemoved, ckrDeviceError, ckrTokenNotPresent, ckrCryptokiNotInitialized:
		return true
	}
	return false
}

type pkcs11Key struct {
	handle uint64
	public crypto.PublicKey
}

type pkcs11Module struct {
	config Config
	lib    *library
	fns    [fnCount]uintptr
	slot   uint64

	mu         sync.Mutex
	session    uint64
	hasSession bool
	keys       map[string]*pkcs11Key
}

func newPkcs11Module(config Config) (*pkcs11Module, error) {
	m := &pkcs11Module{config: config, keys: map[string]*pkcs11Key{}}

	lib, err := openLibrary(config.Library)
	if err != nil {
		return nil, err
	}
	m.lib = lib

	getFunctionList, err := lib.proc("C_GetFunctionList")
	if err != nil {
		return nil, err
	}
	out := cMem(ptrSize)
	if rv := ckrv(call(getFunctionList, addr(out))); rv != ckrOk {
		return nil, &ckError{"C_GetFunctionList", rv}
	}
	list := *(*unsafe.Pointer)(unsafe.Pointer(&out[0]))
	if list == nil {
		return nil, errors.New("C_GetFunctionList returned no function list")
	}

	// CK_FUNCTION_LIST starts with a 2-byte CK_VERSION, the function pointers follow it packed or aligned
	offset := ptrSize
	if packedStructs {
		offset = 2
	}
	table := unsafe.Slice((*byte)(list), offset+fnCount*ptrSize)
	for i := range m.fns {
		m.fns[i] = getPtr(table[offset+i*ptrSize:])
		if m.fns[i] == 0 {
			return nil, fmt.Errorf("the PKCS #11 library %s doesn't have the function %d", config.Library, i)
		}
	}

	// CK_C_INITIALIZE_ARGS: four mutex callbacks, the flags and a reserved pointer
	args := newCStruct(4*ptrSize + ulongSize + ptrSize).ptr(0).ptr(0).ptr(0).ptr(0).ulong(ckfOsLockingOk).ptr(0)
	rv := m.ck(fnInitialize, addr(args.buf))
	runtime.KeepAlive(args)
	if rv != ckrOk && rv != ckrCryptokiAlreadyInitiated {
		return nil, &ckError{"C_Initialize", rv}
	}

	m.slot, err = m.findSlot()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ckrv reads a CK_RV, a C unsigned long
func ckrv(rv uintptr) uint64 {
	if ulongSize == 4 {
		return uint64(uint32(rv))
	}
	return uint64(rv)
}

func (m *pkcs11Module) ck(index int, args ...uintptr) uint64 {
	return ckrv(call(m.fns[index], args...))
}

func (m *pkcs11Module) findSlot() (uint64, error) {
	count := cMem(ulongSize)
	if rv := m.ck(fnGetSlotList, 1, 0, addr(count)); rv != ckrOk {
		return 0, &ckError{"C_GetSlotList", rv}
	}
	n := int(getUlong(count))
	if n == 0 {
		return 0, errors.New("the HSM has no token")
	}
	list := cMem(n * ulongSize)
	if rv := m.ck(fnGetSlotList, 1, addr(list), addr(count)); rv != ckrOk {
		return 0, &ckError{"C_GetSlotList", rv}
	}
	slots := make([]uint64, int(getUlong(count)))
	for i := range slots {
		slots[i] = getUlong(list[i*ulongSize:])
	}

	if m.config.Slot != "" {
		slot, err := strconv.ParseUint(m.config.Slot, 0, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid hsmSlot: %s", m.config.Slot)
		}
		return slot, nil
	}
	if m.config.TokenLabel == "" {
		return slots[0], nil
	}

	labels := []string{}
	for _, slot := range slots {
		label, err := m.getTokenLabel(slot)
		if err != nil {
			return 0, err
		}
		if label == m.config.TokenLabel {
			return slot, nil
		}
		labels = append(labels, label)
	}
	return 0, fmt.Errorf("the HSM has no token labeled %q, the tokens are: %s", m.config.TokenLabel, strings.Join(labels, ", "))
}

// getTokenLabel reads the label, the first 32 bytes of CK_TOKEN_INFO
func (m *pkcs11Module) getTokenLabel(slot uint64) (string, error) {
	info := cMem(1024)
	if rv := m.ck(fnGetTokenInfo, uintptr(slot), addr(info)); rv != ckrOk {
		return "", &ckError{"C_GetTokenInfo", rv}
	}
	return strings.TrimRight(string(info[:32]), " \x00"), nil
}

func (m *pkcs11Module) describe() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	label, err := m.getTokenLabel(m.slot)
	if err != nil {
		return fmt.Sprintf("PKCS #11 slot %d", m.slot)
	}
	return fmt.Sprintf("PKCS #11 slot %d, token %q", m.slot, label)
}

func (m *pkcs11Module) forget(ref string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.keys, ref)
}

// openSession needs m.mu
func (m *pkcs11Module) openSession() error {
	if m.hasSession {
		return nil
	}

	session := cMem(ulongSize)
	if rv := m.ck(fnOpenSession, uintptr(m.slot), ckfSerialSession, 0, 0, addr(session)); rv != ckrOk {
		return &ckError{"C_OpenSession", rv}
	}
	m.session = getUlong(session)
	m.hasSession = true

	if m.config.Pin != "" {
		pin := cBytes([]byte(m.config.Pin))
		rv := m.ck(fnLogin, uintptr(m.session), ckuUser, addr(pin), uintptr(len(m.config.Pin)))
		runtime.KeepAlive(pin)
		if rv != ckrOk && rv != ckrUserAlreadyLoggedIn {
			m.closeSession()
			return &ckError{"C_Login", rv}
		}
	}
	return nil
}

func (m *pkcs11Module) closeSession() {
	if m.hasSession {
		m.ck(fnCloseSession, uintptr(m.session))
	}
	m.hasSession = false
	m.keys = map[string]*pkcs11Key{}
}

// attributeSize is the size of CK_ATTRIBUTE and of CK_MECHANISM: a CK_ULONG, a pointer and a CK_ULONG
func attributeSize() int {
	return 2*ulongSize + ptrSize
}

// newTemplate lays out CK_ATTRIBUTE structs, the values stay referenced by the returned buffers
func newTemplate(attributes []uint64, values [][]byte) (*cStruct, [][]byte) {
	template := newCStruct(len(attributes) * attributeSize())
	buffers := make([][]byte, len(values))
	for i, attribute := range attributes {
		if values[i] == nil {
			template.ulong(attribute).ptr(0).ulong(0)
			continue
		}
		buffers[i] = cBytes(values[i])
		template.ulong(attribute).ptr(addr(buffers[i])).ulong(uint64(len(values[i])))
	}
	return template, buffers
}

func ulongBytes(v uint64) []byte {
	res := make([]byte, ulongSize)
	putUlong(res, v)
	return res
}

func (m *pkcs11Module) findObject(class uint64, attribute uint64, value []byte) (uint64, bool, error) {
	template, buffers := newTemplate([]uint64{ckaClass, attribute}, [][]byte{ulongBytes(class), value})
	rv := m.ck(fnFindObjectsInit, uintptr(m.session), addr(template.buf), 2)
	runtime.KeepAlive(template)
	runtime.KeepAlive(buffers)
	if rv != ckrOk {
		return 0, false, &ckError{"C_FindObjectsInit", rv}
	}

	handle := cMem(ulongSize)
	count := cMem(ulongSize)
	rv = m.ck(fnFindObjects, uintptr(m.session), addr(handle), 1, addr(count))
	m.ck(fnFindObjectsFinal, uintptr(m.session))
	if rv != ckrOk {
		return 0, false, &ckError{"C_FindObjects", rv}
	}
	if getUlong(count) == 0 {
		return 0, false, nil
	}
	return getUlong(handle), true, nil
}

func (m *pkcs11Module) getAttribute(handle uint64, attribute uint64) ([]byte, error) {
	template := newCStruct(attributeSize()).ulong(attribute).ptr(0).ulong(0)
	if rv := m.ck(fnGetAttributeValue, uintptr(m.session), uintptr(handle), addr(template.buf), 1); rv != ckrOk {
		return nil, &ckError{"C_GetAttributeValue", rv}
	}
	length := getUlong(template.buf[ulongSize+ptrSize:])
	if length == 0 || (ulongSize == 4 && length == 0xFFFFFFFF) || length == ^uint64(0) || length > 1<<20 {
		return nil, fmt.Errorf("the attribute 0x%X is not available", attribute)
	}

	value := cMem(int(length))
	template = newCStruct(attributeSize()).ulong(attribute).ptr(addr(value)).ulong(length)
	rv := m.ck(fnGetAttributeValue, uintptr(m.session), uintptr(handle), addr(template.buf), 1)
	runtime.KeepAlive(value)
	if rv != ckrOk {
		return nil, &ckError{"C_GetAttributeValue", rv}
	}
	length = getUlong(template.buf[ulongSize+ptrSize:])
	if length > uint64(len(value)) {
		return nil, fmt.Errorf("the attribute 0x%X is not available", attribute)
	}
	return value[:length], nil
}

// keyAttribute parses a key reference: "id:<hex>" is a CKA_ID, anything else a CKA_LABEL
func keyAttribute(ref string) (uint64, []byte, error) {
	if strings.HasPrefix(ref, "id:") {
		id, err := hex.DecodeString(strings.TrimPrefix(ref, "id:"))
		if err != nil || len(id) == 0 {
			return 0, nil, fmt.Errorf("invalid HSM key ID: %s", ref)
		}
		return ckaId, id, nil
	}
	return ckaLabel, []byte(ref), nil
}

// findKey needs m.mu and an open session
func (m *pkcs11Module) findKey(ref string) (*pkcs11Key, error) {
	if key, ok := m.keys[ref]; ok {
		return key, nil
	}

	attribute, value, err := keyAttribute(ref)
	if err != nil {
		return nil, err
	}
	handle, ok, err := m.findObject(ckoPrivateKey, attribute, value)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("the HSM token has no private key %q, or hsmPin is not set", ref)
	}

	public, err := m.readPublicKey(handle, attribute, value)
	if err != nil {
		return nil, fmt.Errorf("failed to read the public key of the HSM key %q: %v", ref, err)
	}
	key := &pkcs11Key{handle: handle, public: public}
	m.keys[ref] = key
	return key, nil
}

// publicKeyObject finds the public key object of a private key: by its CKA_ID, otherwise by the reference
func (m *pkcs11Module) publicKeyObject(private uint64, attribute uint64, value []byte) (uint64, error) {
	if id, err := m.getAttribute(private, ckaId); err == nil && len(id) > 0 {
		if handle, ok, err := m.findObject(ckoPublicKey, ckaId, id); err == nil && ok {
			return handle, nil
		}
	}
	handle, ok, err := m.findObject(ckoPublicKey, attribute, value)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, errors.New("the HSM token has no public key object for it")
	}
	return handle, nil
}

var (
	oidP256 = asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}
	oidP384 = asn1.ObjectIdentifier{1, 3, 132, 0, 34}
	oidP521 = asn1.ObjectIdentifier{1, 3, 132, 0, 35}
)

func (m *pkcs11Module) readPublicKey(private uint64, attribute uint64, value []byte) (crypto.PublicKey, error) {
	keyType, err := m.getAttribute(private, ckaKeyType)
	if err != nil {
		return nil, err
	}

	if len(keyType) >= ulongSize && getUlong(keyType) == ckkRsa {
		modulus, err := m.getAttribute(private, ckaModulus)
		var exponent []byte
		if err == nil {
			exponent, err = m.getAttribute(private, ckaPublicExponent)
		}
		if err != nil {
			public, err2 := m.publicKeyObject(private, attribute, value)
			if err2 != nil {
				return nil, err
			}
			if modulus, err = m.getAttribute(public, ckaModulus); err != nil {
				return nil, err
			}
			if exponent, err = m.getAttribute(public, ckaPublicExponent); err != nil {
				return nil, err
			}
		}
		e := new(big.Int).SetBytes(exponent)
		if !e.IsInt64() || e.Int64() > 1<<31-1 {
			return nil, errors.New("the RSA public exponent is too large")
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(modulus), E: int(e.Int64())}, nil
	}

	// an EC key
	public, err := m.publicKeyObject(private, attribute, value)
	if err != nil {
		return nil, err
	}
	params, err := m.getAttribute(private, ckaEcParams)
	if err != nil {
		if params, err = m.getAttribute(public, ckaEcParams); err != nil {
			return nil, err
		}
	}
	point, err := m.getAttribute(public, ckaEcPoint)
	if err != nil {
		return nil, err
	}
	return parseEcPublicKey(params, point)
}

func parseEcPublicKey(params []byte, point []byte) (*ecdsa.PublicKey, error) {
	var oid asn1.ObjectIdentifier
	if _, err := asn1.Unmarshal(params, &oid); err != nil {
		return nil, fmt.Errorf("unsupported CKA_EC_PARAMS: %x", params)
	}

	// CKA_EC_POINT is the DER OCTET STRING of the point, some modules give the point itself
	var raw []byte
	if _, err := asn1.Unmarshal(point, &raw); err != nil || len(raw) == 0 || raw[0] != 4 {
		raw = point
	}

	var curve elliptic.Curve
	switch {
	case oid.Equal(oidP256):
		curve = elliptic.P256()
	case oid.Equal(oidP384):
		curve = elliptic.P384()
	case oid.Equal(oidP521):
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", oid.String())
	}
	return ecdsa.ParseUncompressedPublicKey(curve, raw)
}

func (m *pkcs11Module) loadKey(ref string) (crypto.PublicKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var key *pkcs11Key
	err := m.withSession(func() error {
		var err error
		key, err = m.findKey(ref)
		return err
	})
	if err != nil {
		return nil, err
	}
	return key.public, nil
}

// withSession runs f in a session, opening it again once when the HSM lost it; it needs m.mu
func (m *pkcs11Module) withSession(f func() error) error {
	err := m.openSession()
	if err == nil {
		err = f()
	}
	if err == nil || !isSessionLost(err) {
		return err
	}

	m.closeSession()
	if err = m.openSession(); err != nil {
		return err
	}
	return f()
}

var digestInfoPrefixes = map[crypto.Hash][]byte{
	crypto.SHA1:   {0x30, 0x21, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x0e, 0x03, 0x02, 0x1a, 0x05, 0x00, 0x04, 0x14},
	crypto.SHA224: {0x30, 0x2d, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x04, 0x05, 0x00, 0x04, 0x1c},
	crypto.SHA256: {0x30, 0x31, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01, 0x05, 0x00, 0x04, 0x20},
	crypto.SHA384: {0x30, 0x41, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x02, 0x05, 0x00, 0x04, 0x30},
	crypto.SHA512: {0x30, 0x51, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x03, 0x05, 0x00, 0x04, 0x40},
}

var pssHashes = map[crypto.Hash][2]uint64{
	crypto.SHA1:   {ckmSha1, ckgMgf1Sha1},
	crypto.SHA224: {ckmSha224, ckgMgf1Sha224},
	crypto.SHA256: {ckmSha256, ckgMgf1Sha256},
	crypto.SHA384: {ckmSha384, ckgMgf1Sha384},
	crypto.SHA512: {ckmSha512, ckgMgf1Sha512},
}

func (m *pkcs11Module) sign(ref string, public crypto.PublicKey, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	var mechanism uint64
	var param []byte
	data := digest
	isEc := false

	switch public.(type) {
	case *rsa.PublicKey:
		hash := opts.HashFunc()
		if pss, ok := opts.(*rsa.PSSOptions); ok {
			ids, ok := pssHashes[hash]
			if !ok {
				return nil, fmt.Errorf("unsupported RSA-PSS hash: %v", hash)
			}
			saltLength := pss.SaltLength
			if saltLength == rsa.PSSSaltLengthEqualsHash || saltLength == rsa.PSSSaltLengthAuto {
				saltLength = hash.Size()
			}
			mechanism = ckmRsaPkcsPss
			// CK_RSA_PKCS_PSS_PARAMS: the hash mechanism, the MGF and the salt length
			param = newCStruct(3 * ulongSize).ulong(ids[0]).ulong(ids[1]).ulong(uint64(saltLength)).buf
		} else {
			mechanism = ckmRsaPkcs
			if hash != 0 {
				prefix, ok := digestInfoPrefixes[hash]
				if !ok {
					return nil, fmt.Errorf("unsupported RSA hash: %v", hash)
				}
				if len(digest) != hash.Size() {
					return nil, errors.New("the digest length doesn't match the hash")
				}
				data = append(append([]byte{}, prefix...), digest...)
			}
		}
	case *ecdsa.PublicKey:
		isEc = true
		mechanism = ckmEcdsa
	default:
		return nil, fmt.Errorf("unsupported HSM key type: %T", public)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var signature []byte
	err := m.withSession(func() error {
		key, err := m.findKey(ref)
		if err != nil {
			return err
		}
		signature, err = m.signWith(key.handle, mechanism, param, data)
		return err
	})
	if err != nil {
		return nil, err
	}

	if isEc {
		if len(signature) == 0 || len(signature)%2 != 0 {
			return nil, fmt.Errorf("the HSM returned an EC signature of %d bytes", len(signature))
		}
		half := len(signature) / 2
		return marshalEcSignature(signature[:half], signature[half:])
	}
	return signature, nil
}

// signWith needs m.mu and an open session
func (m *pkcs11Module) signWith(handle uint64, mechanism uint64, param []byte, data []byte) ([]byte, error) {
	var paramBuf []byte
	if param != nil {
		paramBuf = cBytes(param)
	}
	mech := newCStruct(attributeSize()).ulong(mechanism).ptr(addr(paramBuf)).ulong(uint64(len(param)))
	rv := m.ck(fnSignInit, uintptr(m.session), addr(mech.buf), uintptr(handle))
	runtime.KeepAlive(mech)
	runtime.KeepAlive(paramBuf)
	if rv != ckrOk {
		return nil, &ckError{"C_SignInit", rv}
	}

	input := cBytes(data)
	signature := cMem(1024)
	length := cMem(ulongSize)
	putUlong(length, uint64(len(signature)))
	rv = m.ck(fnSign, uintptr(m.session), addr(input), uintptr(len(data)), addr(signature), addr(length))
	if rv == ckrBufferTooSmall {
		signature = cMem(int(getUlong(length)))
		rv = m.ck(fnSign, uintptr(m.session), addr(input), uintptr(len(data)), addr(signature), addr(length))
	}
	runtime.KeepAlive(input)
	runtime.KeepAlive(signature)
	if rv != ckrOk {
		return nil, &ckError{"C_Sign", rv}
	}
	return signature[:getUlong(length)], nil
}
