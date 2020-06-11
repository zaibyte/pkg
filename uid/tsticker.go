// Copyright (c) 2020. Temple3x (temple3x@gmail.com)
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

package uid

import (
	"sync/atomic"
	"time"
)

const (
	// epoch is an Unix time.
	// 2020-06-03T08:39:34.000+0800.
	epoch int64 = 1591144774
	// doom is the zai's max Unix time.
	// It will reach the end after 136 years from epoch.
	doom int64 = 5880040774 // epoch + 136 years (about 2^32 seconds).
	// maxTS is the zai's max timestamp.
	maxTS uint32 = uint32(doom - epoch)
)

var ticker *tsTicker

type tsTicker struct {
	_padding0 [64]byte
	ts        uint32
	_padding1 [64]byte

	_padding2 [64]byte
	seqID     uint32
	_padding3 [64]byte

	ticker *time.Ticker

	closed chan bool
}

// StartTicker starts the ticker which running in background.
// TODO may put it in init.
func StartTicker() {

	now := time.Now().Unix()
	if now > doom {
		panic("zai met its doom")
	}

	ticker = &tsTicker{
		ts:     uint32(now - epoch),
		ticker: time.NewTicker(time.Second),
		closed: make(chan bool),
	}

	go logicalTimeMovLoop()
}

func logicalTimeMovLoop() {

	for {
		select {
		case <-ticker.ticker.C:
			if atomic.AddUint32(&ticker.ts, 1) >= maxTS {
				panic("zai met its doom")
			}
		case <-ticker.closed:
			return
		}
	}
}

// StopTicker releases the resource.
func StopTicker() {
	ticker.ticker.Stop()
	ticker.closed <- true
}

// Ts2Time converts zai ts to time.
func Ts2Time(ts uint32) time.Time {
	sec := int64(ts) + epoch
	return time.Unix(sec, 0)
}
