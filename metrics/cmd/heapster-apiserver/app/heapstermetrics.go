// Copyright 2016 Google Inc. All Rights Reserved.
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

package app

import (
	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/genericapiserver"
	genericoptions "k8s.io/kubernetes/pkg/genericapiserver/options"

	"k8s.io/heapster/metrics/apis/metrics"
	_ "k8s.io/heapster/metrics/apis/metrics/install"
)

func installMetricsAPIs(s *genericoptions.ServerRunOptions, g *genericapiserver.GenericAPIServer) {
	heapsterResources := map[string]rest.Storage{
	// TODO: Register resources with storage
	}
	heapsterGroupMeta := registered.GroupOrDie(metrics.GroupName)
	apiGroupInfo := genericapiserver.APIGroupInfo{
		GroupMeta: *heapsterGroupMeta,
		VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
			"v1alpha1": heapsterResources,
		},
		OptionsExternalVersion: &registered.GroupOrDie(api.GroupName).GroupVersion,
		Scheme:                 api.Scheme,
		ParameterCodec:         api.ParameterCodec,
		NegotiatedSerializer:   api.Codecs,
	}
	if err := g.InstallAPIGroup(&apiGroupInfo); err != nil {
		glog.Fatalf("Error in registering group versions: %v", err)
	}
}
