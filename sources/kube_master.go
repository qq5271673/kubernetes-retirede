package sources

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	kube_api "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/golang/glog"
)

var (
	argMaster     = flag.String("kubernetes_master", "", "Kubernetes master IP")
	argMasterAuth = flag.String("kubernetes_master_auth", "", "username:password to access the master")
)

type KubeMasterSource struct {
	master         string
	authMasterUser string
	authMasterPass string
}

func PostRequestAndGetValue(client *http.Client, req *http.Request, value interface{}) error {
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, value)
	if err != nil {
		return fmt.Errorf("Got '%s': %v", string(body), err)
	}
	return nil
}

func (self *KubeMasterSource) masterListMinionsUrl() string {
	return self.master + "/api/v1beta1/minions"
}

// Returns a map of minion hostnames to their corresponding IPs.
func (self *KubeMasterSource) ListMinions() (map[string]string, error) {
	var minions kube_api.MinionList
	req, err := http.NewRequest("GET", self.masterListMinionsUrl(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(self.authMasterUser, self.authMasterPass)
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	err = PostRequestAndGetValue(httpClient, req, &minions)
	if err != nil {
		return nil, err
	}
	var hosts map[string]string
	for _, value := range minions.Items {
		addrs, err := net.LookupIP(value.ID)
		if err == nil {
			hosts[value.ID] = addrs[0].String()
		} else {
			glog.Errorf("Skipping host %s since looking up its IP failed - %s", value.ID, err)
		}
	}

	return hosts, nil
}

func (self *KubeMasterSource) masterListPodsUrl() string {
	return self.master + "/api/v1beta1/pods"
}

func (self *KubeMasterSource) parsePod(pod *kube_api.Pod) (string, *Pod) {
	hostname := pod.CurrentState.Host
	localPod := Pod{
		Status:     string(pod.CurrentState.Status),
		PodIP:      pod.CurrentState.PodIP,
		Labels:     make(map[string]string, 0),
		Containers: make([]ContainerDesc, 0),
	}
	for key, value := range pod.Labels {
		localPod.Labels[key] = value
	}
	for _, container := range pod.CurrentState.Manifest.Containers {
		localContainer := ContainerDesc{
			Name: container.Name,
			ID:   pod.CurrentState.Info[container.Name].ID,
		}
		localPod.Containers = append(localPod.Containers, localContainer)
	}
	return hostname, &localPod
}

// Returns a map of minion hostnames to the Pods running in them.
func (self *KubeMasterSource) ListPods() (map[string][]Pod, error) {
	var pods kube_api.PodList
	req, err := http.NewRequest("GET", self.masterListPodsUrl(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(self.authMasterUser, self.authMasterPass)
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	err = PostRequestAndGetValue(httpClient, req, &pods)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]Pod, 0)
	for _, pod := range pods.Items {
		hostname, pod := self.parsePod(&pod)
		out[hostname] = append(out[hostname], *pod)
	}
	return out, nil
}

func NewKubeMasterSource() (*KubeMasterSource, error) {
	if len(*argMaster) == 0 {
		return nil, fmt.Errorf("kubernetes_master flag not specified")
	}
	if len(*argMasterAuth) == 0 || len(strings.Split(*argMasterAuth, ":")) != 2 {
		return nil, fmt.Errorf("kubernetes_master_auth invalid")
	}
	authInfo := strings.Split(*argMasterAuth, ":")
	return &KubeMasterSource{
		master:         "https://" + *argMaster,
		authMasterUser: authInfo[0],
		authMasterPass: authInfo[1],
	}, nil
}
