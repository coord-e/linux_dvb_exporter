// Copyright ord_e
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  	 http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frontend

// #include <linux/dvb/frontend.h>
import "C"

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/paypal/gatt/linux/gioctl"
)

type Frontend struct {
	File *os.File
}

func Open(adapter uint, frontend uint) (*Frontend, error) {
	dev := fmt.Sprintf("/dev/dvb/adapter%d/frontend%d", adapter, frontend)
	file, err := os.Open(dev)
	if err != nil {
		return nil, err
	}

	return &Frontend{
		File: file,
	}, nil
}

func (f *Frontend) Close() error {
	return f.File.Close()
}

var (
	feReadStatus  = gioctl.IoR('o', 69, C.sizeof_enum_fe_status)
	feGetProperty = gioctl.IoR('o', 83, C.sizeof_struct_dtv_properties)
)

type Status struct {
	HasSignal  bool
	HasCarrier bool
	HasViterbi bool
	HasSync    bool
	HasLock    bool
	Timedout   bool
	Reinit     bool
}

func (f *Frontend) ReadStatus() (Status, error) {
	var data uint32
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.File.Fd(),
		feReadStatus,
		uintptr(unsafe.Pointer(&data)),
	)

	var status Status
	if errno != 0 {
		return status, errno
	}

	if data&C.FE_HAS_SIGNAL != 0 {
		status.HasSignal = true
	}

	if data&C.FE_HAS_CARRIER != 0 {
		status.HasCarrier = true
	}

	if data&C.FE_HAS_VITERBI != 0 {
		status.HasViterbi = true
	}

	if data&C.FE_HAS_SYNC != 0 {
		status.HasSync = true
	}

	if data&C.FE_HAS_LOCK != 0 {
		status.HasLock = true
	}

	if data&C.FE_TIMEDOUT != 0 {
		status.Timedout = true
	}

	if data&C.FE_REINIT != 0 {
		status.Reinit = true
	}

	return status, nil
}

type DecibelStat struct {
	Decibel *float64
	Ratio   *float64
}

type Stats struct {
	SignalStrength    DecibelStat
	CNR               DecibelStat
	PreErrorBitCount  *uint64
	PreTotalBitCount  *uint64
	PostErrorBitCount *uint64
	PostTotalBitCount *uint64
	ErrorBlockCount   *uint64
	TotalBlockCount   *uint64
}

const (
	dtvStatSignalStrength    = 62
	dtvStatCNR               = 63
	dtvStatPreErrorBitCount  = 64
	dtvStatPreTotalBitCount  = 65
	dtvStatPostErrorBitCount = 66
	dtvStatPostTotalBitCount = 67
	dtvStatErrorBlockCount   = 68
	dtvStatTotalBlockCount   = 69
)

func (f *Frontend) GetStats() (Stats, error) {
	props := []C.struct_dtv_property{
		{cmd: dtvStatSignalStrength},
		{cmd: dtvStatCNR},
		{cmd: dtvStatPreErrorBitCount},
		{cmd: dtvStatPreTotalBitCount},
		{cmd: dtvStatPostErrorBitCount},
		{cmd: dtvStatPostTotalBitCount},
		{cmd: dtvStatErrorBlockCount},
		{cmd: dtvStatTotalBlockCount},
	}
	data := C.struct_dtv_properties{
		num:   C.uint(len(props)),
		props: &props[0],
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.File.Fd(),
		feGetProperty,
		uintptr(unsafe.Pointer(&data)),
	)

	if errno != 0 {
		return Stats{}, errno
	}

	stats := Stats{}
	for _, p := range props {
		st := (*C.struct_dtv_fe_stats)(unsafe.Pointer(&p.u))
		if st.len <= 0 {
			// property not supported
			continue
		}
		stat := st.stat[0]

		switch p.cmd {
		case dtvStatSignalStrength:
			stats.SignalStrength = decibelStatOf(stat)
		case dtvStatCNR:
			stats.CNR = decibelStatOf(stat)
		case dtvStatPreErrorBitCount:
			stats.PreErrorBitCount = countOf(stat)
		case dtvStatPreTotalBitCount:
			stats.PreTotalBitCount = countOf(stat)
		case dtvStatPostErrorBitCount:
			stats.PostErrorBitCount = countOf(stat)
		case dtvStatPostTotalBitCount:
			stats.PostTotalBitCount = countOf(stat)
		case dtvStatErrorBlockCount:
			stats.ErrorBlockCount = countOf(stat)
		case dtvStatTotalBlockCount:
			stats.TotalBlockCount = countOf(stat)
		}
	}

	return stats, nil
}

func decibelStatOf(stat C.struct_dtv_stats) DecibelStat {
	dstat := DecibelStat{}
	switch stat.scale {
	case C.FE_SCALE_DECIBEL:
		decibel := float64(svalueOf(stat)) * 0.001
		dstat.Decibel = &decibel
	case C.FE_SCALE_RELATIVE:
		ratio := float64(uvalueOf(stat)) / 0xffff
		dstat.Ratio = &ratio
	}
	return dstat
}

func countOf(stat C.struct_dtv_stats) *uint64 {
	var count uint64
	if stat.scale == C.FE_SCALE_COUNTER {
		count = uvalueOf(stat)
	}
	return &count
}

func uvalueOf(stat C.struct_dtv_stats) uint64 {
	data := C.GoBytes(unsafe.Pointer(&stat), C.sizeof_struct_dtv_stats)
	return *(*uint64)(unsafe.Pointer(&data[1]))
}

func svalueOf(stat C.struct_dtv_stats) int64 {
	data := C.GoBytes(unsafe.Pointer(&stat), C.sizeof_struct_dtv_stats)
	return *(*int64)(unsafe.Pointer(&data[1]))
}
