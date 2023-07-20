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

package nodes

import (
	"fmt"
	"net"
	"sync"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/cache"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/golang/glog"
)

type kubeNodes struct {
	client *client.Client
	// a means to list all minions
	nodeLister *cache.StoreToNodeLister
	reflector  *cache.Reflector
	// Used to stop the existing reflector.
	stopChan   chan struct{}
	goodNodes  []string       // guarded by stateLock
	nodeErrors map[string]int // guarded by stateLock
	stateLock  sync.RWMutex
}

func (self *kubeNodes) recordNodeError(name string) {
	self.stateLock.Lock()
	defer self.stateLock.Unlock()

	self.nodeErrors[name]++
}

func (self *kubeNodes) recordGoodNodes(nodes []string) {
	self.stateLock.Lock()
	defer self.stateLock.Unlock()

	self.goodNodes = nodes
}

func parseSelectorOrDie(s string) labels.Selector {
	selector, err := labels.Parse(s)
	if err != nil {
		panic(err)
	}
	return selector
}

func (self *kubeNodes) List() (*NodeList, error) {
	nodeList := newNodeList()
	allNodes, err := self.nodeLister.List()
	if err != nil {
		glog.Errorf("failed to list minions via watch interface - %v", err)
		return nil, fmt.Errorf("failed to list minions via watch interface - %v", err)
	}
	glog.V(5).Infof("all kube nodes: %+v", allNodes)

	goodNodes := []string{}
	for _, node := range allNodes.Items {
		nodeInfo := Info{}
		hostname := ""
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case api.NodeExternalIP:
				nodeInfo.PublicIP = addr.Address
			case api.NodeInternalIP:
				nodeInfo.InternalIP = addr.Address
			case api.NodeHostName:
				hostname = addr.Address
			}
		}
		if hostname == "" {
			hostname = node.Name
		}
		if nodeInfo.InternalIP == "" {
			addrs, err := net.LookupIP(hostname)
			if err == nil {
				nodeInfo.InternalIP = addrs[0].String()
			} else {
				glog.Errorf("Skipping host %s since looking up its IP failed - %s", node.Name, err)
				self.recordNodeError(node.Name)
			}
		}

		nodeList.Items[Host(hostname)] = nodeInfo
		goodNodes = append(goodNodes, node.Name)
	}
	self.recordGoodNodes(goodNodes)
	glog.V(5).Infof("kube nodes found: %+v", nodeList)
	return nodeList, nil
}

func (self *kubeNodes) getState() string {
	self.stateLock.RLock()
	defer self.stateLock.RUnlock()

	state := "\tHealthy Nodes:\n"
	for _, node := range self.goodNodes {
		state += fmt.Sprintf("\t\t%s\n", node)
	}
	if len(self.nodeErrors) > 0 {
		state += fmt.Sprintf("\tNode Errors: %+v\n", self.nodeErrors)
	} else {
		state += "\tNo node errors\n"
	}
	return state
}

func (self *kubeNodes) DebugInfo() string {
	desc := "Kubernetes Nodes plugin: \n"
	desc += self.getState()
	desc += "\n"

	return desc
}

func NewKubeNodes(client *client.Client) (NodesApi, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	lw := cache.NewListWatchFromClient(client, "minions", api.NamespaceAll, fields.Everything())
	nodeLister := &cache.StoreToNodeLister{Store: cache.NewStore(cache.MetaNamespaceKeyFunc)}
	reflector := cache.NewReflector(lw, &api.Node{}, nodeLister.Store, 0)
	stopChan := make(chan struct{})
	reflector.RunUntil(stopChan)

	return &kubeNodes{
		client:     client,
		nodeLister: nodeLister,
		reflector:  reflector,
		stopChan:   stopChan,
		nodeErrors: make(map[string]int),
	}, nil
}
