// Copyright 2022 The casbin Authors. All Rights Reserved.
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

package ip

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net"
)

const Null = "N/A"

var (
	ErrInvalidIp = errors.New("invalid ip format")
	std          *Locator
)

// Init default locator with dataFile
func Init(dataFile string) (err error) {
	if std != nil {
		return
	}
	std, err = NewLocator(dataFile)
	return
}

// Init default locator with data
func InitWithData(data []byte) {
	if std != nil {
		return
	}
	std = NewLocatorWithData(data)
	return
}

// Find locationInfo by ip string
// It will return err when ipstr is not a valid format
func Find(ipstr string) (*LocationInfo, error) {
	return std.Find(ipstr)
}

// Find locationInfo by uint32
func FindByUint(ip uint32) *LocationInfo {
	return std.FindByUint(ip)
}

//-----------------------------------------------------------------------------

// New locator with dataFile
func NewLocator(dataFile string) (loc *Locator, err error) {
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return
	}
	loc = NewLocatorWithData(data)
	return
}

// New locator with data
func NewLocatorWithData(data []byte) (loc *Locator) {
	loc = new(Locator)
	loc.init(data)
	return
}

type Locator struct {
	textData   []byte
	indexData1 []uint32
	indexData2 []int
	indexData3 []int
	index      []int
}

type LocationInfo struct {
	Country string
	Region  string
	City    string
	Isp     string
}

// Find locationInfo by ip string
// It will return err when ipstr is not a valid format
func (loc *Locator) Find(ipstr string) (info *LocationInfo, err error) {
	ip := net.ParseIP(ipstr).To4()
	if ip == nil || ip.To4() == nil {
		err = ErrInvalidIp
		return
	}
	info = loc.FindByUint(binary.BigEndian.Uint32([]byte(ip)))
	return
}

// Find locationInfo by uint32
func (loc *Locator) FindByUint(ip uint32) (info *LocationInfo) {
	end := len(loc.indexData1) - 1
	if ip>>24 != 0xff {
		end = loc.index[(ip>>24)+1]
	}
	idx := loc.findIndexOffset(ip, loc.index[ip>>24], end)
	off := loc.indexData2[idx]
	return newLocationInfo(loc.textData[off : off+loc.indexData3[idx]])
}

// binary search
func (loc *Locator) findIndexOffset(ip uint32, start, end int) int {
	for start < end {
		mid := (start + end) / 2
		if ip > loc.indexData1[mid] {
			start = mid + 1
		} else {
			end = mid
		}
	}

	if loc.indexData1[end] >= ip {
		return end
	}

	return start
}

func (loc *Locator) init(data []byte) {
	textoff := int(binary.BigEndian.Uint32(data[:4]))

	loc.textData = data[textoff-1024:]

	loc.index = make([]int, 256)
	for i := 0; i < 256; i++ {
		off := 4 + i*4
		loc.index[i] = int(binary.LittleEndian.Uint32(data[off : off+4]))
	}

	nidx := (textoff - 4 - 1024 - 1024) / 8

	loc.indexData1 = make([]uint32, nidx)
	loc.indexData2 = make([]int, nidx)
	loc.indexData3 = make([]int, nidx)

	for i := 0; i < nidx; i++ {
		off := 4 + 1024 + i*8
		loc.indexData1[i] = binary.BigEndian.Uint32(data[off : off+4])
		loc.indexData2[i] = int(uint32(data[off+4]) | uint32(data[off+5])<<8 | uint32(data[off+6])<<16)
		loc.indexData3[i] = int(data[off+7])
	}
	return
}

func newLocationInfo(str []byte) *LocationInfo {

	var info *LocationInfo

	fields := bytes.Split(str, []byte("\t"))
	switch len(fields) {
	case 4:
		// free version
		info = &LocationInfo{
			Country: string(fields[0]),
			Region:  string(fields[1]),
			City:    string(fields[2]),
		}
	case 5:
		// pay version
		info = &LocationInfo{
			Country: string(fields[0]),
			Region:  string(fields[1]),
			City:    string(fields[2]),
			Isp:     string(fields[4]),
		}
	default:
		panic("unexpected ip info:" + string(str))
	}

	if len(info.Country) == 0 {
		info.Country = Null
	}
	if len(info.Region) == 0 {
		info.Region = Null
	}
	if len(info.City) == 0 {
		info.City = Null
	}
	if len(info.Isp) == 0 {
		info.Isp = Null
	}
	return info
}
