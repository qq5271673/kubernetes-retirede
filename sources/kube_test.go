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
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/heapster/sources/api"
	"github.com/GoogleCloudPlatform/heapster/sources/datasource"
	"github.com/GoogleCloudPlatform/heapster/sources/nodes"
	"github.com/stretchr/testify/require"
)

type fakePodsApi struct {
	podList []api.Pod
}

func (self *fakePodsApi) List(nodeList *nodes.NodeList) ([]api.Pod, error) {
	return self.podList, nil
}

func (self *fakePodsApi) DebugInfo() string {
	return ""
}

type fakeKubeletApi struct {
	container *api.Container
}

func (self *fakeKubeletApi) GetContainer(host datasource.Host, numStats int) (*api.Container, error) {
	return self.container, nil
}

func TestKubeSourceBasic(t *testing.T) {
	nodesApi := &fakeNodesApi{nodes.NodeList{}}
	podsApi := &fakePodsApi{[]api.Pod{}}
	kubeSource := &kubeSource{
		lastQuery:   time.Now(),
		kubeletPort: "10250",
		nodesApi:    nodesApi,
		podsApi:     podsApi,
		kubeletApi:  &fakeKubeletApi{nil},
	}
	_, err := kubeSource.GetInfo()
	require.NoError(t, err)
	require.NotEmpty(t, kubeSource.DebugInfo())
}

func TestKubeSourceDetail(t *testing.T) {
	nodeList := nodes.NodeList{
		Items: map[nodes.Host]nodes.Info{
			nodes.Host("test-machine-b"): {InternalIP: "10.10.10.1"},
			nodes.Host("test-machine-1"): {InternalIP: "10.10.10.0"},
		},
	}
	podList := []api.Pod{
		{
			Name: "blah",
		},
		{
			Name: "blah1",
		},
	}
	container := &api.Container{
		Name: "test",
	}
	nodesApi := &fakeNodesApi{nodeList}
	podsApi := &fakePodsApi{podList}
	kubeletApi := &fakeKubeletApi{container}

	kubeSource := &kubeSource{
		lastQuery:   time.Now(),
		kubeletPort: "10250",
		nodesApi:    nodesApi,
		podsApi:     podsApi,
		kubeletApi:  kubeletApi,
	}
	data, err := kubeSource.GetInfo()
	require.NoError(t, err)
	require.NotEmpty(t, data)
}
