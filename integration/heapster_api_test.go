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

package integration

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/stretchr/testify/require"
	api_v1 "k8s.io/heapster/api/v1/types"
	sink_api "k8s.io/heapster/sinks/api"
	"k8s.io/heapster/sinks/cache"
	kube_api "k8s.io/kubernetes/pkg/api"
	apiErrors "k8s.io/kubernetes/pkg/api/errors"
)

const (
	kTestZone        = "us-central1-b"
	targetTags       = "kubernetes-minion"
	heapsterBuildDir = "../deploy/docker"
)

var (
	kubeVersions           = flag.String("kube_versions", "", "Comma separated list of kube versions to test against. By default will run the test against an existing cluster")
	heapsterControllerFile = flag.String("heapster_controller", "../deploy/kube-config/standalone/heapster-controller.json", "Path to heapster replication controller file.")
	heapsterServiceFile    = flag.String("heapster_service", "../deploy/kube-config/standalone/heapster-service.json", "Path to heapster service file.")
	heapsterImage          = flag.String("heapster_image", "heapster:e2e_test", "heapster docker image that needs to be tested.")
	avoidBuild             = flag.Bool("nobuild", false, "When true, a new heapster docker image will not be created and pushed to test cluster nodes.")
	namespace              = flag.String("namespace", "default", "namespace to be used for testing")
	maxRetries             = flag.Int("retries", 100, "Number of attempts before failing this test.")
	runForever             = flag.Bool("run_forever", false, "If true, the tests are run in a loop forever.")
)

func deleteAll(fm kubeFramework, ns string, service *kube_api.Service, rc *kube_api.ReplicationController) error {
	glog.V(2).Infof("Deleting rc %s/%s...", ns, rc.Name)
	if err := fm.DeleteRC(ns, rc); err != nil {
		glog.V(2).Infof("Failed to delete rc: %v", err)
		return err
	}
	glog.V(2).Infof("Deleted rc %s/%s.", ns, rc.Name)

	glog.V(2).Infof("Deleting service %s/%s.", ns, service.Name)
	if err := fm.DeleteService(ns, service); err != nil {
		glog.V(2).Infof("Failed to delete service: %v", err)
		return err
	}
	glog.V(2).Infof("Deleted service %s/%s.", ns, service.Name)
	return nil
}

func createAll(fm kubeFramework, ns string, service **kube_api.Service, rc **kube_api.ReplicationController) error {
	glog.V(2).Infof("Creating rc %s/%s...", ns, (*rc).Name)
	if newRc, err := fm.CreateRC(ns, *rc); err != nil {
		glog.V(2).Infof("Failed to create rc: %v", err)
		return err
	} else {
		*rc = newRc
	}
	glog.V(2).Infof("Created rc %s/%s.", ns, (*rc).Name)

	glog.V(2).Infof("Creating service %s/%s...", ns, (*service).Name)
	if newSvc, err := fm.CreateService(ns, *service); err != nil {
		glog.V(2).Infof("Failed to create service: %v", err)
		return err
	} else {
		*service = newSvc
	}
	glog.V(2).Infof("Created servuce %s/%s.", ns, (*service).Name)

	return nil
}

func removeHeapsterImage(fm kubeFramework) error {
	glog.V(2).Infof("Removing heapster image.")
	if err := removeDockerImage(*heapsterImage); err != nil {
		glog.Errorf("Failed to remove Heapster image: %v", err)
	} else {
		glog.V(2).Infof("Heapster image removed.")
	}
	if nodes, err := fm.GetNodes(); err == nil {
		for _, node := range nodes {
			host := strings.Split(node, ".")[0]
			cleanupRemoteHost(host, kTestZone)
		}
	} else {
		glog.Errorf("failed to cleanup nodes - %v", err)
	}
	return nil
}

func buildAndPushHeapsterImage(hostnames []string) error {
	glog.V(2).Info("Building and pushing Heapster image...")
	curwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(heapsterBuildDir); err != nil {
		return err
	}
	if err := buildDockerImage(*heapsterImage); err != nil {
		return err
	}
	for _, host := range hostnames {
		if err := copyDockerImage(*heapsterImage, host, kTestZone); err != nil {
			return err
		}
	}
	glog.V(2).Info("Heapster image pushed.")
	return os.Chdir(curwd)
}

func getHeapsterRcAndSvc(fm kubeFramework) (*kube_api.Service, *kube_api.ReplicationController, error) {
	// Add test docker image
	rc, err := fm.ParseRC(*heapsterControllerFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse heapster controller - %v", err)
	}
	rc.Spec.Template.Spec.Containers[0].Image = *heapsterImage
	rc.Spec.Template.Spec.Containers[0].ImagePullPolicy = kube_api.PullNever
	// increase logging level
	rc.Spec.Template.Spec.Containers[0].Env = append(rc.Spec.Template.Spec.Containers[0].Env, kube_api.EnvVar{Name: "FLAGS", Value: "--vmodule=*=3"})
	svc, err := fm.ParseService(*heapsterServiceFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse heapster service - %v", err)
	}

	return svc, rc, nil
}

func buildAndPushDockerImages(fm kubeFramework) error {
	if *avoidBuild {
		return nil
	}
	nodes, err := fm.GetNodes()
	if err != nil {
		return err
	}
	hostnames := []string{}
	for _, node := range nodes {
		hostnames = append(hostnames, strings.Split(node, ".")[0])
	}

	return buildAndPushHeapsterImage(hostnames)
}

const (
	metricsEndpoint       = "/api/v1/metric-export"
	metricsSchemaEndpoint = "/api/v1/metric-export-schema"
	sinksEndpoint         = "/api/v1/sinks"
)

func getTimeseries(fm kubeFramework, svc *kube_api.Service) ([]*api_v1.Timeseries, error) {
	body, err := fm.Client().Get().
		Namespace(svc.Namespace).
		Prefix("proxy").
		Resource("services").
		Name(svc.Name).
		Suffix(metricsEndpoint).
		Do().Raw()
	if err != nil {
		return nil, err
	}
	var timeseries []*api_v1.Timeseries
	if err := json.Unmarshal(body, &timeseries); err != nil {
		glog.V(2).Infof("response body: %v", string(body))
		return nil, err
	}
	return timeseries, nil
}

func getSchema(fm kubeFramework, svc *kube_api.Service) (*api_v1.TimeseriesSchema, error) {
	body, err := fm.Client().Get().
		Namespace(svc.Namespace).
		Prefix("proxy").
		Resource("services").
		Name(svc.Name).
		Suffix(metricsSchemaEndpoint).
		Do().Raw()
	if err != nil {
		return nil, err
	}
	var timeseriesSchema api_v1.TimeseriesSchema
	if err := json.Unmarshal(body, &timeseriesSchema); err != nil {
		glog.V(2).Infof("response body: %v", string(body))
		return nil, err
	}
	return &timeseriesSchema, nil
}

func getModelMetrics(fm kubeFramework, svc *kube_api.Service, pod *kube_api.Pod) (*api_v1.MetricResultList, error) {
	url := fmt.Sprintf("/api/v1/model/namespaces/%s/pod-list/%s/metrics/%s",
		pod.Namespace,
		pod.Name,
		"cpu-usage")

	body, err := fm.Client().Get().
		Namespace(svc.Namespace).
		Prefix("proxy").
		Resource("services").
		Name(svc.Name).
		Suffix(url).
		Do().Raw()
	if err != nil {
		return nil, err
	}
	var metrics api_v1.MetricResultList
	if err := json.Unmarshal(body, &metrics); err != nil {
		glog.V(2).Infof("response body: %v", string(body))
		return nil, err
	}
	return &metrics, nil

}

var expectedSystemContainers = map[string]struct{}{
	"machine":       {},
	"kubelet":       {},
	"kube-proxy":    {},
	"system":        {},
	"docker-daemon": {},
}

func runHeapsterMetricsTest(fm kubeFramework, svc *kube_api.Service, expectedNodes, expectedPods []string) error {
	timeseries, err := getTimeseries(fm, svc)
	if err != nil {
		return err
	}
	if len(timeseries) == 0 {
		return fmt.Errorf("expected non zero timeseries")
	}
	schema, err := getSchema(fm, svc)
	if err != nil {
		return err
	}
	// Build a map of metric names to metric descriptors.
	mdMap := map[string]*api_v1.MetricDescriptor{}
	for idx := range schema.Metrics {
		mdMap[schema.Metrics[idx].Name] = &schema.Metrics[idx]
	}
	actualPods := map[string]bool{}
	actualNodes := map[string]bool{}
	actualSystemContainers := map[string]map[string]struct{}{}
	for _, ts := range timeseries {
		// Verify the relevant labels are present.
		// All common labels must be present.
		for _, label := range sink_api.CommonLabels() {
			_, exists := ts.Labels[label.Key]
			if !exists {
				return fmt.Errorf("timeseries: %v does not contain common label: %v", ts, label)
			}
		}
		podName, podMetric := ts.Labels[sink_api.LabelPodName.Key]
		if podMetric {
			for _, label := range sink_api.PodLabels() {
				_, exists := ts.Labels[label.Key]
				if !exists {
					return fmt.Errorf("timeseries: %v does not contain pod label: %v", ts, label)
				}
			}
		}
		if podMetric {
			actualPods[podName] = true
		} else {
			if cName, ok := ts.Labels[sink_api.LabelContainerName.Key]; ok {
				hostname, ok := ts.Labels[sink_api.LabelHostname.Key]
				if !ok {
					return fmt.Errorf("hostname label missing on container %+v", ts)
				}

				if cName == cache.NodeContainerName {
					actualNodes[hostname] = true
				}
				if _, exists := expectedSystemContainers[cName]; exists {
					if actualSystemContainers[cName] == nil {
						actualSystemContainers[cName] = map[string]struct{}{}
					}
					actualSystemContainers[cName][hostname] = struct{}{}
				}
			} else {
				return fmt.Errorf("container_name label missing on timeseries - %v", ts)
			}
		}

		for metricName, points := range ts.Metrics {
			md, exists := mdMap[metricName]
			if !exists {
				return fmt.Errorf("unexpected metric %q", metricName)
			}
			for _, point := range points {
				for _, label := range md.Labels {
					_, exists := point.Labels[label.Key]
					if !exists {
						return fmt.Errorf("metric %q point %v does not contain metric label: %v", metricName, point, label)
					}
				}
			}

		}
	}
	// Validate that system containers are running on all the nodes.
	// This test could fail if one of the containers was down while the metrics sample was collected.
	for cName, hosts := range actualSystemContainers {
		for _, host := range expectedNodes {
			if _, ok := hosts[host]; !ok {
				return fmt.Errorf("System container %q not found on host: %q - %v", cName, host, actualSystemContainers)
			}
		}
	}

	if !expectedItemsExist(expectedPods, actualPods) {
		return fmt.Errorf("expected pods don't exist.\nExpected: %v\nActual:%v", expectedPods, actualPods)
	}
	if !expectedItemsExist(expectedNodes, actualNodes) {
		return fmt.Errorf("expected nodes don't exist.\nExpected: %v\nActual:%v", expectedNodes, actualNodes)
	}

	return nil
}

func expectedItemsExist(expectedItems []string, actualItems map[string]bool) bool {
	if len(actualItems) < len(expectedItems) {
		return false
	}
	for _, item := range expectedItems {
		if _, found := actualItems[item]; !found {
			return false
		}
	}
	return true
}

func getSinks(fm kubeFramework, svc *kube_api.Service) ([]string, error) {
	body, err := fm.Client().Get().
		Namespace(svc.Namespace).
		Prefix("proxy").
		Resource("services").
		Name(svc.Name).
		Suffix(sinksEndpoint).
		Do().Raw()
	if err != nil {
		return nil, err
	}
	var sinks []string
	if err := json.Unmarshal(body, &sinks); err != nil {
		glog.V(2).Infof("response body: %v", string(body))
		return nil, err
	}
	return sinks, nil
}

func setSinks(fm kubeFramework, svc *kube_api.Service, sinks []string) error {
	data, err := json.Marshal(sinks)
	if err != nil {
		return err
	}
	return fm.Client().Post().
		Namespace(svc.Namespace).
		Prefix("proxy").
		Resource("services").
		Name(svc.Name).
		Suffix(sinksEndpoint).
		SetHeader("Content-Type", "application/json").
		Body(data).
		Do().Error()
}

func getErrorCauses(err error) string {
	serr, ok := err.(*apiErrors.StatusError)
	if !ok {
		return ""
	}
	var causes []string
	for _, c := range serr.ErrStatus.Details.Causes {
		causes = append(causes, c.Message)
	}
	return strings.Join(causes, ", ")
}

func runSinksTest(fm kubeFramework, svc *kube_api.Service) error {
	for _, newSinks := range [...][]string{
		{},
		{
			"gcm",
		},
		{},
	} {
		if err := setSinks(fm, svc, newSinks); err != nil {
			glog.Errorf("Could not set sinks. Causes: %s", getErrorCauses(err))
			return err
		}
		sinks, err := getSinks(fm, svc)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(sinks, newSinks) {
			return fmt.Errorf("expected %v sinks, found %v", newSinks, sinks)
		}
	}
	return nil
}

func runModelTest(fm kubeFramework, svc *kube_api.Service) error {
	podList, err := fm.GetPodList()
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("Empty pod list")
	}
	for _, pod := range podList.Items {
		metrics, err := getModelMetrics(fm, svc, &pod)
		if err != nil {
			return fmt.Errorf("error while getting metrics for %s/%s: %v", pod.Namespace, pod.Name, err)
		}
		if len(metrics.Items) == 0 {
			return fmt.Errorf("empty metrics for: %s/%s", pod.Namespace, pod.Name)
		}
		if len(metrics.Items[0].Metrics) == 0 {
			return fmt.Errorf("empty metrics for: %s/%s", pod.Namespace, pod.Name)
		}
		if time.Now().Sub(metrics.Items[0].LatestTimestamp).Seconds() > 30 {
			return fmt.Errorf("Corrupted last timestamp for: %s/%s", pod.Namespace, pod.Name)
		}
		if time.Now().Sub(metrics.Items[0].Metrics[0].Timestamp).Seconds() > 30 {
			return fmt.Errorf("Corrupted timestamp for: %s/%s", pod.Namespace, pod.Name)
		}
		if metrics.Items[0].Metrics[0].Value > 10000 {
			return fmt.Errorf("Value too big for: %s/%s", pod.Namespace, pod.Name)
		}
	}
	return nil
}

func apiTest(kubeVersion string) error {
	fm, err := newKubeFramework(kubeVersion)
	if err != nil {
		return err
	}
	if err := buildAndPushDockerImages(fm); err != nil {
		return err
	}
	// Create heapster pod and service.
	svc, rc, err := getHeapsterRcAndSvc(fm)
	if err != nil {
		return err
	}
	ns := *namespace
	if err := deleteAll(fm, ns, svc, rc); err != nil {
		return err
	}
	if err := createAll(fm, ns, &svc, &rc); err != nil {
		return err
	}
	if err := fm.WaitUntilPodRunning(ns, rc.Spec.Template.Labels, time.Minute); err != nil {
		return err
	}
	if err := fm.WaitUntilServiceActive(svc, time.Minute); err != nil {
		return err
	}
	expectedPods, err := fm.GetPodNames()
	if err != nil {
		return err
	}
	expectedNodes, err := fm.GetNodes()
	if err != nil {
		return err
	}
	testFuncs := []func() error{
		func() error {
			glog.V(2).Infof("Heapster metrics test...")
			err := runHeapsterMetricsTest(fm, svc, expectedNodes, expectedPods)
			if err == nil {
				glog.V(2).Infof("Heapster metrics test: OK")
			} else {
				glog.V(2).Infof("Heapster metrics test error: %v", err)
			}
			return err
		},
		func() error {
			glog.V(2).Infof("Sinks test...")
			err := runSinksTest(fm, svc)
			if err == nil {
				glog.V(2).Infof("Sinks test: OK")
			} else {
				glog.V(2).Infof("Sinks test error: %v", err)
			}
			return err
		},
		func() error {
			glog.V(2).Infof("Model test")
			err := runModelTest(fm, svc)
			if err == nil {
				glog.V(2).Infof("Model test: OK")
			} else {
				glog.V(2).Infof("Model test error: %v", err)
			}
			return err
		},
	}
	attempts := *maxRetries
	glog.Infof("Starting tests")
	for {
		var err error
		for _, testFunc := range testFuncs {
			if err = testFunc(); err != nil {
				break
			}
		}
		if *runForever {
			continue
		}
		if err == nil {
			glog.V(2).Infof("All tests passed.")
			break
		}
		if attempts == 0 {
			glog.V(2).Info("Too many attempts.")
			return err
		}
		glog.V(2).Infof("Some tests failed. Retrying.")
		attempts--
		time.Sleep(time.Second * 10)
	}
	deleteAll(fm, ns, svc, rc)
	removeHeapsterImage(fm)
	return nil
}

func runApiTest() error {
	tempDir, err := ioutil.TempDir("", "deploy")
	if err != nil {
		return nil
	}
	defer os.RemoveAll(tempDir)
	if *kubeVersions == "" {
		return apiTest("")
	}
	kubeVersionsList := strings.Split(*kubeVersions, ",")
	for _, kubeVersion := range kubeVersionsList {
		if err := apiTest(kubeVersion); err != nil {
			return err
		}
	}
	return nil
}

func TestHeapster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping heapster kubernetes integration test.")
	}
	require.NoError(t, runApiTest())
}
