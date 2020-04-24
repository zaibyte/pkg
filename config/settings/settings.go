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

// Settings is the global settings of zai.
// Don't modify it unless you totally know what will happen.
package settings

// ----- Redundancy ----- //
const (
	// Erasure Codes Policy.
	// I hide this config in const, because the benefit of changing them
	// is not much, the cost maybe higher or the durability maybe lower,
	// 9+4 is a balanced choice.
	//
	// Default: 9+4, it can save 33.3% I/O when one data extent lost.
	// ECData is the num of data extents in a group.
	ECData = 9
	// ECParity is the num of parity extents in a group.
	ECParity = 4
	// ECTotal is the num of all extents in a group.
	ECTotal = ECData + ECParity
)

// MntRoot
const MntRoot = "/zai" // Disk device mount path.

// <DefaultLogPath>/<appName>/access.log
// & <DefaultLogPath>/<appName>/error.log
const DefaultLogPath = "/var/log/zai"

// MaxObjSize is the maximum size of an object.
const MaxObjSize = 32 * 1024 * 1024

const (
	kb = 1024
	mb = kb * 1024
	gb = mb * 1024
)

const (
	// MaxTinySize is a global setting,
	// FrontEnd need it to choice different Write Buffer:
	// 1. If object size <= MaxTinySize, put it into Buffer Tiny
	// (If there is no writable group, upload will be failed. Actually we can use Buffer for this situation,
	// but it maybe dangerous if I allow this action, because if there are too many small objects in Buffer,
	// it may damage the performance hugely.).
	// TODO there should be a option, let tiny object get into normal buffer
	// 2. If object size > MaxTinySize, put it into Buffer.
	//
	// Default value is 256KB.
	// So the max number of objects in a single Buffer Tiny or Buffer extent is 65536,
	// this small number will help to compress the index size.
	// Maybe it's too big or too small for your business model,
	// make sure all things could work with the new size.
	// You can get help by develop docs/issue/email.
	MaxTinySize  = kb * 256
	BufExtSize   = gb * 16
	StoreExtSize = mb * 256 // Don't set it < 256MB, because the bigger the less rows in Finder.
)
