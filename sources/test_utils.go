// Copyright 2014 Google Inc. All Rights Reserved.
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

package sources

import (
	"time"

	"k8s.io/heapster/sources/api"
	"k8s.io/heapster/sources/datasource"
	"k8s.io/heapster/sources/nodes"
	"k8s.io/kubernetes/pkg/api/latest"
	"k8s.io/kubernetes/pkg/runtime"
)

type fakeNodesApi struct {
	nodeList nodes.NodeList
}

func (self *fakeNodesApi) List() (*nodes.NodeList, error) {
	return &self.nodeList, nil
}

func (self *fakeNodesApi) DebugInfo() string {
	return "fake nodes plugin: no data"
}

type fakeKubeletApi struct {
	container  *api.Container
	containers []api.Container
}

func (self *fakeKubeletApi) GetContainer(host datasource.Host, start, end time.Time) (*api.Container, error) {
	return self.container, nil
}

func (self *fakeKubeletApi) GetAllRawContainers(host datasource.Host, start, end time.Time) ([]api.Container, error) {
	return self.containers, nil
}

func body(obj runtime.Object) string {
	if obj != nil {
		bs, _ := latest.GroupOrDie("").Codec.Encode(obj)
		body := string(bs)
		return body
	}
	return ""
}
