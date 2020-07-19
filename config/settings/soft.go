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
// Copyright 2017-2019 Lei Ni (nilei81@gmail.com) and other Dragonboat authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// The idea of soft settings is from Dragonboat.
// And some settings are copied from Dragonboat.

package settings

// Tuning configuration parameters here will impact the performance of your
// system. It will not corrupt your data. Only tune these parameters when
// you know what you are doing.

// Soft is the soft settings that can be changed after the deployment of a
// system.
var Soft = getDefaultSoftSettings()

type soft struct {

	//
	// transport
	//
	// SendQueueLength is the length of the send queue used to hold messages
	// exchanged between nodehosts. You may need to increase this value when
	// you want to host large number nodes per nodehost.
	SendQueueLength uint64

	// PerConnBufSize is the size of the per connection buffer used for
	// receiving incoming messages.
	PerConnectionSendBufSize uint64
	// PerConnectionRecvBufSize is the size of the recv buffer size.
	PerConnectionRecvBufSize uint64
}

func getDefaultSoftSettings() soft {
	return soft{

		SendQueueLength:          1024 * 2,
		PerConnectionSendBufSize: 2 * 1024 * 1024,
		PerConnectionRecvBufSize: 2 * 1024 * 1024,
	}
}
