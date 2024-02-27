// Copyright 2015 Google Inc. All Rights Reserved.
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

package cache

import (
	"time"

	"github.com/golang/glog"
)

type gcStore struct {
	bufferDuration time.Duration
	store          TimeStore
}

func (gcs *gcStore) Put(timestamp time.Time, data interface{}) error {
	if err := gcs.store.Put(timestamp, data); err != nil {
		return err
	}
	gcs.reapOldData()
	return nil
}

func (gcs *gcStore) Get(start, end time.Time) ([]interface{}, error) {
	return gcs.store.Get(start, end)
}

func (gcs *gcStore) GetAll() []interface{} {
	return gcs.store.GetAll()
}

func (gcs *gcStore) Last() interface{} {
	return gcs.store.Last()
}

func (gcs *gcStore) Delete(start, end time.Time) error {
	return gcs.store.Delete(start, end)
}

func (gcs *gcStore) reapOldData() {
	end := time.Now().Add(-gcs.bufferDuration)
	start := time.Unix(0, 0)
	if err := gcs.store.Delete(start, end); err != nil {
		glog.Fatalf("failed to delete old data")
	}
}

func NewGCStore(store TimeStore, bufferDuration, gcDuration time.Duration) TimeStore {
	gcStore := &gcStore{
		bufferDuration: bufferDuration,
		store:          store,
	}
	return gcStore
}
