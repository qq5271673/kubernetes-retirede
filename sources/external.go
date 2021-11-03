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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
)

// While updating this, also update heapster/deploy/Dockerfile.
const HostsFile = "/var/run/heapster/hosts"

type ExternalSource struct {
	cadvisor *cadvisorSource
}

func (self *ExternalSource) getCadvisorHosts() (*CadvisorHosts, error) {
	fi, err := os.Stat(HostsFile)
	if err != nil {
		return nil, fmt.Errorf("cannot stat hosts_file %q: %s", HostsFile, err)
	}
	if fi.Size() == 0 {
		return &CadvisorHosts{}, nil
	}
	contents, err := ioutil.ReadFile(HostsFile)
	if err != nil {
		return nil, err
	}
	var cadvisorHosts CadvisorHosts
	err = json.Unmarshal(contents, &cadvisorHosts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal contents of file %s. Error: %s", HostsFile, err)
	}
	glog.V(1).Infof("Using cAdvisor hosts %+v", cadvisorHosts)
	return &cadvisorHosts, nil
}

func (self *ExternalSource) GetPods() ([]Pod, error) {
	return []Pod{}, nil
}

func (self *ExternalSource) GetInfo() (ContainerData, error) {
	hosts, err := self.getCadvisorHosts()
	if err != nil {
		return ContainerData{}, err
	}

	containers, nodes, err := self.cadvisor.fetchData(hosts)
	if err != nil {
		glog.Error(err)
		return ContainerData{}, nil
	}

	return ContainerData{
		Containers: containers,
		Machine:    nodes,
	}, nil
}

func newExternalSource() (Source, error) {
	if _, err := os.Stat(HostsFile); err != nil {
		return nil, fmt.Errorf("cannot stat hosts_file %s. Error: %s", HostsFile, err)
	}
	cadvisorSource := newCadvisorSource()
	return &ExternalSource{
		cadvisor: cadvisorSource,
	}, nil
}

func (self *ExternalSource) GetConfig() string {
	desc := "Source type: External\n"
	// TODO(rjnagal): Cache config?
	hosts, err := self.getCadvisorHosts()
	if err != nil {
		desc += fmt.Sprintf("\tFailed to read host config: %s", err)
	}
	desc += fmt.Sprintf("\tHosts: %+v\n", *hosts)
	desc += "\n"
	return desc
}
