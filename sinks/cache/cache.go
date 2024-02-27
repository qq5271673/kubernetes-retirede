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
	source_api "github.com/GoogleCloudPlatform/heapster/sources/api"
	cadvisor_api "github.com/google/cadvisor/info/v1"
)

type Metadata struct {
	Name      string
	Namespace string
	UID       string
	Hostname  string
	Labels    map[string]string
}

type ContainerMetricElement struct {
	Spec  *cadvisor_api.ContainerSpec
	Stats *cadvisor_api.ContainerStats
}

type ContainerElement struct {
	Metadata
	Metrics []*ContainerMetricElement
}

type PodElement struct {
	Metadata
	// map of container name to container element.
	Containers map[string]*ContainerElement
	// TODO: Cache history of Spec and Status.
}

type Cache interface {
	StorePods([]source_api.Pod) error
	StoreContainers([]source_api.Container) error
}
