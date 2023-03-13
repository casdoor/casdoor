// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package sync

import "sync"

func startSyncJob(db1 *Database, db2 *Database) error {
	var wg sync.WaitGroup

	// start canal1 replication
	go db1.startCanal(db2)
	wg.Add(1)

	// start canal2 replication
	go db2.startCanal(db1)
	wg.Add(1)

	wg.Wait()
	return nil
}
