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
//
// MIT License
//
// Copyright (c) 2017 Zach Bjornson

#include "textflag.h"

#define dst R8
#define src R9
#define mask16 X2
#define mask32 Y2
#define low Y5
#define hextable Y6

// func hexEncAVX2(dst, src *byte)
TEXT ·hexEncAVX2(SB), NOSPLIT, $0
	MOVQ         d+0(FP), dst
	MOVQ         s+8(FP), src

	// Prepare a 32Bytes mask filled with $0x0f.
	MOVB         $0x0f, DX
	VPINSRB      $0x00, DX, mask16, mask16
	VPBROADCASTB mask16, mask16
	VINSERTI128  $1, mask16, mask32, mask32

	//
	MOVQ         ·lows(SB), R10
	VMOVDQU      (R10), low

	// Load hex table.
	MOVQ         ·hextable(SB), R11
	VMOVDQU      (R11), hextable

    // Split every byte in src: [a >> 4, a & 0x0f]
	VMOVDQU    (src), X1
	VPMOVZXBW X1, Y1
	VPSRLQ    $4, Y1, Y3
	VPSHUFB   low, Y1, Y1
	VPOR      Y1, Y3, Y1
	VPAND     mask32, Y1, Y1

	// Shuffle according hex table.
	VPSHUFB   Y1, hextable, Y1

	VMOVDQU Y1, (dst)
	RET

