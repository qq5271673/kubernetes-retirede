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
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	kube_api "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	kube_client "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	kube_labels "github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/golang/glog"
	cadvisor "github.com/google/cadvisor/info"
)

// Kubernetes released supported and tested against.
var kubeVersions = []string{"v0.3"}

const (
	// Cadvisor port in kubernetes.
	cadvisorPort = 4194

	kubeClientVersion = "v1beta1"
)

type PodInstance struct {
	Pod    string
	PodId  string
	HostIp string
}

type KubeSource struct {
	client      *kube_client.Client
	lastQuery   time.Time
	kubeletPort string
	stateLock   sync.RWMutex
	goodNodes   []string            // guarded by stateLock
	nodeErrors  map[string]int      // guarded by stateLock
	podErrors   map[PodInstance]int // guarded by stateLock
}

type nodeList CadvisorHosts

func (self *KubeSource) recordNodeError(name string) {
	self.stateLock.Lock()
	defer self.stateLock.Unlock()

	self.nodeErrors[name]++
}

func (self *KubeSource) recordPodError(pod Pod) {
	// Heapster knows about pods before they are up and running on a node.
	// Ignore errors for Pods that are not Running.
	if pod.Status != "Running" {
		return
	}

	self.stateLock.Lock()
	defer self.stateLock.Unlock()

	podInstance := PodInstance{Pod: pod.Name, PodId: pod.ID, HostIp: pod.HostIP}
	self.podErrors[podInstance]++
}

func (self *KubeSource) recordGoodNodes(nodes []string) {
	self.stateLock.Lock()
	defer self.stateLock.Unlock()

	self.goodNodes = nodes
}

func (self *KubeSource) getState() string {
	self.stateLock.RLock()
	defer self.stateLock.RUnlock()

	state := "\tHealthy Nodes:\n"
	for _, node := range self.goodNodes {
		state += fmt.Sprintf("\t\t%s\n", node)
	}
	if len(self.nodeErrors) != 0 {
		state += fmt.Sprintf("\tNode Errors: %+v\n", self.nodeErrors)
	} else {
		state += "\tNo node errors\n"
	}
	if len(self.podErrors) != 0 {
		state += fmt.Sprintf("\tPod Errors: %+v\n", self.podErrors)
	} else {
		state += "\tNo pod errors\n"
	}
	return state
}

// Returns a map of minion hostnames to their corresponding IPs.
func (self *KubeSource) listMinions() (*nodeList, error) {
	nodeList := &nodeList{
		Port:  cadvisorPort,
		Hosts: make(map[string]string, 0),
	}
	minions, err := self.client.Nodes().List()
	if err != nil {
		return nil, err
	}
	goodNodes := []string{}
	for _, minion := range minions.Items {
		addrs, err := net.LookupIP(minion.Name)
		if err == nil {
			nodeList.Hosts[minion.Name] = addrs[0].String()
			goodNodes = append(goodNodes, minion.Name)
		} else {
			glog.Errorf("Skipping host %s since looking up its IP failed - %s", minion.Name, err)
			self.recordNodeError(minion.Name)
		}
	}
	self.recordGoodNodes(goodNodes)

	return nodeList, nil
}

func (self *KubeSource) parsePod(pod *kube_api.Pod) *Pod {
	localPod := Pod{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		ID:         string(pod.UID),
		Hostname:   pod.Status.Host,
		Status:     string(pod.Status.Phase),
		PodIP:      pod.Status.PodIP,
		Labels:     make(map[string]string, 0),
		Containers: make([]*Container, 0),
	}
	for key, value := range pod.Labels {
		localPod.Labels[key] = value
	}
	for _, container := range pod.Spec.Containers {
		localContainer := newContainer()
		localContainer.Name = container.Name
		localPod.Containers = append(localPod.Containers, localContainer)
	}
	glog.V(2).Infof("found pod: %+v", localPod)

	return &localPod
}

// Returns a map of minion hostnames to the Pods running in them.
func (self *KubeSource) getPods() ([]Pod, error) {
	pods, err := self.client.Pods(kube_api.NamespaceAll).List(kube_labels.Everything())
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("got pods from api server %+v", pods)
	// TODO(vishh): Add API Version check. Fail if Kubernetes returns an invalid API Version.
	out := make([]Pod, 0)
	for _, pod := range pods.Items {
		glog.V(2).Infof("Got Kube Pod: %+v", pod)
		pod := self.parsePod(&pod)
		addrs, err := net.LookupIP(pod.Hostname)
		if err != nil {
			glog.Errorf("Skipping host %s since looking up its IP failed - %s", pod.Hostname, err)
			self.recordNodeError(pod.Hostname)
			continue
		}
		pod.HostIP = addrs[0].String()
		out = append(out, *pod)
	}

	return out, nil
}

func (self *KubeSource) getStatsFromKubelet(pod Pod, containerName string) (cadvisor.ContainerSpec, []*cadvisor.ContainerStats, error) {
	var containerInfo cadvisor.ContainerInfo
	values := url.Values{}
	values.Add("num_stats", strconv.Itoa(int(time.Since(self.lastQuery)/time.Second)))
	url := "http://" + pod.HostIP + ":" + self.kubeletPort + filepath.Join("/stats", pod.Namespace, pod.Name, pod.ID, containerName) + "?" + values.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return cadvisor.ContainerSpec{}, []*cadvisor.ContainerStats{}, err
	}
	err = PostRequestAndGetValue(&http.Client{}, req, &containerInfo)
	if err != nil {
		glog.Errorf("failed to get stats from kubelet url: %s - %s\n", url, err)
		self.recordPodError(pod)
		return cadvisor.ContainerSpec{}, []*cadvisor.ContainerStats{}, nil
	}

	return containerInfo.Spec, containerInfo.Stats, nil
}

func (self *KubeSource) getNodesInfo() ([]RawContainer, error) {
	kubeNodes, err := self.listMinions()
	if err != nil {
		return []RawContainer{}, err
	}
	nodesInfo := []RawContainer{}
	for node, ip := range kubeNodes.Hosts {
		spec, stats, err := self.getStatsFromKubelet(Pod{HostIP: ip}, "/")
		if err != nil {
			glog.V(1).Infof("Failed to get machine stats from kubelet for node %s", node)
			return []RawContainer{}, err
		}
		if len(stats) > 0 {
			container := RawContainer{node, Container{"/", spec, stats}}
			nodesInfo = append(nodesInfo, container)
		}
	}

	return nodesInfo, nil
}

func (self *KubeSource) GetInfo() (ContainerData, error) {
	pods, err := self.getPods()
	if err != nil {
		return ContainerData{}, err
	}
	for _, pod := range pods {
		for _, container := range pod.Containers {
			spec, stats, err := self.getStatsFromKubelet(pod, container.Name)
			if err != nil {
				return ContainerData{}, err
			}
			glog.V(2).Infof("Fetched stats from kubelet for container %s in pod %s", container.Name, pod.Name)
			container.Stats = stats
			container.Spec = spec
		}
	}
	nodesInfo, err := self.getNodesInfo()
	if err != nil {
		return ContainerData{}, err
	}
	glog.V(2).Info("Fetched list of nodes from the master")
	self.lastQuery = time.Now()

	return ContainerData{Pods: pods, Machine: nodesInfo}, nil
}

func newKubeSource() (*KubeSource, error) {
	if len(*argMaster) == 0 {
		return nil, fmt.Errorf("kubernetes_master flag not specified")
	}
	kubeClient := kube_client.NewOrDie(&kube_client.Config{
		Host:     "http://" + *argMaster,
		Version:  kubeClientVersion,
		Insecure: true,
	})

	glog.Infof("Using Kubernetes client with master %q and version %s\n", *argMaster, kubeClientVersion)
	glog.Infof("Using kubelet port %q", *argKubeletPort)
	glog.Infof("Support kubelet versions %v", kubeVersions)

	return &KubeSource{
		client:      kubeClient,
		lastQuery:   time.Now(),
		kubeletPort: *argKubeletPort,
		nodeErrors:  make(map[string]int),
		podErrors:   make(map[PodInstance]int),
	}, nil
}

func (self *KubeSource) GetConfig() string {
	desc := "Source type: Kube\n"
	desc += fmt.Sprintf("\tClient config: master ip %q, version %s\n", *argMaster, kubeClientVersion)
	desc += fmt.Sprintf("\tUsing kubelet port %q\n", self.kubeletPort)
	desc += fmt.Sprintf("\tSupported kubelet versions %v\n", kubeVersions)
	desc += self.getState()
	desc += "\n"
	return desc
}
