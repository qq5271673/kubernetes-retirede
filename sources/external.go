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

// This file is dedicated to heapster running outside of kubernetes. Heapster
// will poll a file to get the hosts that it needs to monitor and will collect
// stats from cadvisor running on those hosts.

package sources

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/heapster/sources/datasource"
	"github.com/GoogleCloudPlatform/heapster/sources/nodes"
	"github.com/golang/glog"
)

var argCadvisorPort = flag.Int("cadvisor_port", 8080, "The port on which cadvisor binds to on all nodes.")

type externalSource struct {
	cadvisorPort string
	cadvisorApi  datasource.Cadvisor
	pollDuration time.Duration
	nodesApi     nodes.NodesApi
	lastQuery    time.Time
}

func (self *externalSource) GetInfo() (ContainerData, error) {
	nodeList, err := self.nodesApi.List()
	if err != nil {
		return ContainerData{}, err
	}

	result := ContainerData{}
	for hostname, info := range nodeList.Items {
		host := datasource.Host{
			IP:   info.InternalIP,
			Port: self.cadvisorPort,
		}
		subcontainers, node, err := self.cadvisorApi.GetAllContainers(host, self.numStatsToFetch())
		if err != nil {
			glog.Error(err)
			continue
		}
		for _, cont := range subcontainers {
			if cont != nil {
				result.Containers = append(result.Containers, Container{
					Hostname: string(hostname),
					Name:     cont.Name,
					Spec:     cont.Spec,
					Stats:    cont.Stats,
				})
			}
		}
		if node != nil {
			result.Machine = append(result.Machine, Container{
				Hostname: string(hostname),
				Name:     node.Name,
				Spec:     node.Spec,
				Stats:    node.Stats,
			})
		}
	}
	self.lastQuery = time.Now()

	return result, nil
}

func (self *externalSource) numStatsToFetch() int {
	numStats := int(self.pollDuration / time.Second)
	if time.Since(self.lastQuery) > self.pollDuration {
		numStats = int(time.Since(self.lastQuery) / time.Second)
	}
	return numStats
}

func newExternalSource(pollDuration time.Duration) (Source, error) {
	if *argCadvisorPort <= 0 {
		return nil, fmt.Errorf("invalid cadvisor port - %d", *argCadvisorPort)
	}
	nodesApi, err := nodes.NewExternalNodes()
	if err != nil {
		return nil, err
	}
	return &externalSource{
		pollDuration: pollDuration,
		cadvisorApi:  datasource.NewCadvisor(),
		nodesApi:     nodesApi,
		cadvisorPort: strconv.Itoa(*argCadvisorPort),
		lastQuery:    time.Now(),
	}, nil
}

func (self *externalSource) DebugInfo() string {
	desc := "Source type: External\n"
	// TODO(rjnagal): Cache config?
	nodeList, err := self.nodesApi.List()
	if err != nil {
		desc += fmt.Sprintf("\tFailed to read host config: %s", err)
	}
	desc += fmt.Sprintf("\tNodeList: %+v\n", *nodeList)
	desc += "\n"
	return desc
}
