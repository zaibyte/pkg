/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xmath

import "testing"

func TestRound(t *testing.T) {
	f := 1.1
	var i float64
	for i = 0; i < 0.05; i += 0.01 {
		if Round(f+i, 1) != 1.1 {
			t.Fatal("mismatch")
		}
	}
	for i = 0.05; i < 0.1; i += 0.01 {
		if Round(f+i, 1) != 1.2 {
			t.Fatal("mismatch")
		}
	}
}
