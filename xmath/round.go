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

import "math"

// Round rounds a float64 and cuts it by n.
// n: decimal places.
// e.g.
// f = 1.001, n = 2, return 1.00
func Round(f float64, n int) float64 {
	pow10n := math.Pow10(n)
	return math.Trunc(f*pow10n+0.5) / pow10n
}
