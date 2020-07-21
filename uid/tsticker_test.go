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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.uber.org/goleak"
)

func TestTicker(t *testing.T) {
	defer goleak.VerifyNone(t)

	StopTicker()
}

func TestTs2Time(t *testing.T) {
	assert.Equal(t, time.Unix(epoch, 0), TsToTime(0))
	assert.Equal(t, time.Unix(epoch+1, 0), TsToTime(1))
}

func TestTsNanoToTime(t *testing.T) {
	assert.Equal(t, time.Unix(0, epochNano), TsNanoToTime(0))
	assert.Equal(t, time.Unix(0, epochNano+1), TsNanoToTime(1))
}

func TestTickerMov(t *testing.T) {
	defer StopTicker()

	t0 := atomic.LoadUint32(&ticker.ts)
	time.Sleep(3 * time.Second)
	t1 := atomic.LoadUint32(&ticker.ts)

	if t1 <= t0 {
		t.Fatal("ticker backwards")
	}
	if t1-t0 == 1 {
		t.Fatal("ticker is slower than expect")
	}

	if t1-t0 > 3+1 {
		t.Fatal("ticker is faster than expect")
	}
}
