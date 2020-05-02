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

const (
	kb = 1024
	mb = kb * 1024
	gb = mb * 1024
)

const (
	// MaxObjSize is the maximum size of an object.
	//
	// Zai is designed for LOSF(lots of small files), although it can support
	// much larger object, but it's better to split it then put the chunks
	// of this object to Zai.
	// That's because too big file may block other read/write ops
	// and the cost of retries is higher.
	MaxObjSize = 4 * mb

	// MntRoot
	MntRoot = "/zai" // Disk device mount path.

	// DefaultLogRoot is the default log files path root.
	// e.g.:
	// <DefaultLogRoot>/<appName>/access.log
	// & <DefaultLogRoot>/<appName>/error.log
	DefaultLogRoot = "/var/log/zai"
)

/* ----- ZBuffer ----- */
const (
	// Replicas is the number of replicas for each extent.
	//
	// ZBuffer is temporary storage layer, it's meaningless to keep more than 3 replicas.
	// But if less than it, we won't lose the ability of data consistency.
	Replicas = 3

	// BufExtSize is the size of ZBuffer's extent.
	// There will be 131,072 extents at most in a single box with 16PB usable capacity.
	// In practice, only part of capacity will be used as ZBuffer,
	// for example, with 2PB usable capacity, only 16,384 extents.
	// All extents infos could be loaded in Zai's memory cache.
	BufExtSize = 128 * gb
)

/* ----- ZStore ----- */
const (
	// Erasure-Codes Policy.
	// Because the benefit of changing them is not much,
	// the cost maybe higher or the durability maybe lower,
	// 12+4 is a balanced choice.
	//
	// It can save 33.3% I/O in reconstruction process
	// when one data extent lost.
	//
	// ECDataExts is the num of data extents in a Erasure-Codes extent.
	ECDataExts = 12
	// ECParityExts is the num of parity extents in a Erasure-Codes extent.
	ECParityExts = 4
	// ECExts is the num of all extents in a Erasure-Codes extent.
	ECExts = ECDataExts + ECParityExts

	// MaxStoreExtSize is the size of ZStore's extent.
	// There will be 67,108,864 extents at most in a single box with 16PB usable capacity.
	// It won't be hard to search the extent by oid.
	//
	// The benefits of keeping ZStore extent size small:
	// 1. Making extent could be in blocking mode. The cost of restart isn't high.
	// 2. GC process could be in blocking mode too.
	MaxStoreExtSize = 256 * mb
)
