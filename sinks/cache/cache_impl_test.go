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
	"testing"
	"time"

	source_api "github.com/GoogleCloudPlatform/heapster/sources/api"
	cadvisor "github.com/google/cadvisor/info/v1"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuzz(t *testing.T) {
	cache := NewCache(time.Hour)
	var (
		pods       []source_api.Pod
		containers []source_api.Container
	)
	f := fuzz.New().NumElements(2, 10).NilChance(0)
	f.Fuzz(&pods)
	f.Fuzz(&containers)
	assert := assert.New(t)
	assert.NoError(cache.StorePods(pods))
	assert.NoError(cache.StoreContainers(containers))
	zeroTime := time.Time{}
	assert.NotEmpty(cache.GetFreeContainers(zeroTime, zeroTime))
	assert.NotEmpty(cache.GetPods(zeroTime, zeroTime))
}

func getContainer(name string) source_api.Container {
	f := fuzz.New().NumElements(2, 10).NilChance(0)
	containerSpec := cadvisor.ContainerSpec{
		CreationTime:  time.Now(),
		HasCpu:        true,
		HasMemory:     true,
		HasNetwork:    true,
		HasFilesystem: true,
		HasDiskIo:     true,
	}
	containerStats := make([]*cadvisor.ContainerStats, 1)
	f.Fuzz(&containerStats)
	for idx := range containerStats {
		containerStats[idx].Timestamp = time.Now()
	}
	return source_api.Container{
		Name:  name,
		Spec:  containerSpec,
		Stats: containerStats,
	}
}

func TestRealCacheData(t *testing.T) {
	containers := []source_api.Container{
		getContainer("container1"),
	}
	pods := []source_api.Pod{
		{
			PodMetadata: source_api.PodMetadata{
				Name:      "pod1",
				ID:        "123",
				Namespace: "test",
				Hostname:  "1.2.3.4",
				Status:    "Running",
			},
			Containers: containers,
		},
		{
			PodMetadata: source_api.PodMetadata{
				Name:      "pod2",
				ID:        "1234",
				Namespace: "test",
				Hostname:  "1.2.3.5",
				Status:    "Running",
			},
			Containers: containers,
		},
	}
	cache := NewCache(time.Hour)
	assert := assert.New(t)
	assert.NoError(cache.StorePods(pods))
	assert.NoError(cache.StoreContainers(containers))
	actualPods := cache.GetPods(time.Time{}, time.Time{})
	actualContainer := cache.GetNodes(time.Time{}, time.Now())
	actualContainer = append(actualContainer, cache.GetFreeContainers(time.Time{}, time.Now())...)
	actualPodsMap := map[string]*PodElement{}
	for _, pod := range actualPods {
		actualPodsMap[pod.Name] = pod
	}
	for _, expectedPod := range pods {
		pod, exists := actualPodsMap[expectedPod.Name]
		require.True(t, exists)
		require.NotEmpty(t, pod.Containers)
		assert.NotEmpty(pod.Containers[0].Metrics)
	}
	actualContainerMap := map[string]*ContainerElement{}
	for _, cont := range actualContainer {
		actualContainerMap[cont.Name] = cont
	}
	for _, expectedContainer := range containers {
		ce, exists := actualContainerMap[expectedContainer.Name]
		assert.True(exists)
		assert.NotEmpty(ce.Metrics)
	}
}
