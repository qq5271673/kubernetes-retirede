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

package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	restful "github.com/emicklei/go-restful"

	"k8s.io/heapster/metrics/api/v1/types"
	"k8s.io/heapster/metrics/core"
	"k8s.io/heapster/metrics/sinks/metric"
)

// errModelNotActivated is the error that is returned by the API handlers
// when manager.model has not been initialized.
var errModelNotActivated = errors.New("the model is not activated")

var metricNamesConversion = map[string]string{
	"cpu-usage":      "cpu/usage_rate",
	"cpu-limit":      "cpu/limit",
	"memory-limit":   "memory/limit",
	"memory-usage":   "memory/usage",
	"memory-working": "memory/working_set",
}

// RegisterModel registers the Model API endpoints.
// All endpoints that end with a {metric-name} also receive a start time query parameter.
// The start and end times should be specified as a string, formatted according to RFC 3339.
func (a *Api) RegisterModel(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/api/v1/model").
		Doc("Root endpoint of the stats model").
		Consumes("*/*").
		Produces(restful.MIME_JSON)

	// The /metrics/ endpoint returns a list of all available metrics for the Cluster entity of the model.
	ws.Route(ws.GET("/metrics/").
		To(a.availableClusterMetrics).
		Doc("Get a list of all available metrics for the Cluster entity").
		Operation("availableMetrics"))

	// The /metrics/{metric-name} endpoint exposes an aggregated metric for the Cluster entity of the model.
	ws.Route(ws.GET("/metrics/{metric-name}").
		To(a.clusterMetrics).
		Doc("Export an aggregated cluster-level metric").
		Operation("clusterMetrics").
		Param(ws.PathParameter("metric-name", "The name of the requested metric").DataType("string")).
		Param(ws.QueryParameter("start", "Start time for requested metric").DataType("string")).
		Param(ws.QueryParameter("end", "End time for requested metric").DataType("string")).
		Writes(types.MetricResult{}))

	// The /nodes/{node-name}/metrics endpoint returns a list of all available metrics for a Node entity.
	ws.Route(ws.GET("/nodes/{node-name}/metrics/").
		To(a.availableNodeMetrics).
		Doc("Get a list of all available metrics for a Node entity").
		Operation("availableMetrics").
		Param(ws.PathParameter("node-name", "The name of the node to lookup").DataType("string")))

	// The /nodes/{node-name}/metrics/{metric-name} endpoint exposes a metric for a Node entity of the model.
	// The {node-name} parameter is the hostname of a specific node.
	ws.Route(ws.GET("/nodes/{node-name}/metrics/{metric-name}").
		To(a.nodeMetrics).
		Doc("Export a node-level metric").
		Operation("nodeMetrics").
		Param(ws.PathParameter("node-name", "The name of the node to lookup").DataType("string")).
		Param(ws.PathParameter("metric-name", "The name of the requested metric").DataType("string")).
		Param(ws.QueryParameter("start", "Start time for requested metric").DataType("string")).
		Param(ws.QueryParameter("end", "End time for requested metric").DataType("string")).
		Writes(types.MetricResult{}))

	if a.runningInKubernetes {
		// The /namespaces/{namespace-name}/metrics endpoint returns a list of all available metrics for a Namespace entity.
		ws.Route(ws.GET("/namespaces/{namespace-name}/metrics").
			To(a.availableNamespaceMetrics).
			Doc("Get a list of all available metrics for a Namespace entity").
			Operation("availableMetrics").
			Param(ws.PathParameter("namespace-name", "The name of the namespace to lookup").DataType("string")))

		// The /namespaces/{namespace-name}/metrics/{metric-name} endpoint exposes an aggregated metrics
		// for a Namespace entity of the model.
		ws.Route(ws.GET("/namespaces/{namespace-name}/metrics/{metric-name}").
			To(a.namespaceMetrics).
			Doc("Export an aggregated namespace-level metric").
			Operation("namespaceMetrics").
			Param(ws.PathParameter("namespace-name", "The name of the namespace to lookup").DataType("string")).
			Param(ws.PathParameter("metric-name", "The name of the requested metric").DataType("string")).
			Param(ws.QueryParameter("start", "Start time for requested metrics").DataType("string")).
			Param(ws.QueryParameter("end", "End time for requested metric").DataType("string")).
			Writes(types.MetricResult{}))

		// The /namespaces/{namespace-name}/pods/{pod-name}/metrics endpoint returns a list of all available metrics for a Pod entity.
		ws.Route(ws.GET("/namespaces/{namespace-name}/pods/{pod-name}/metrics").
			To(a.availablePodMetrics).
			Doc("Get a list of all available metrics for a Pod entity").
			Operation("availableMetrics").
			Param(ws.PathParameter("namespace-name", "The name of the namespace to lookup").DataType("string")).
			Param(ws.PathParameter("pod-name", "The name of the pod to lookup").DataType("string")))

		// The /namespaces/{namespace-name}/pods/{pod-name}/metrics/{metric-name} endpoint exposes
		// an aggregated metric for a Pod entity of the model.
		ws.Route(ws.GET("/namespaces/{namespace-name}/pods/{pod-name}/metrics/{metric-name}").
			To(a.podMetrics).
			Doc("Export an aggregated pod-level metric").
			Operation("podMetrics").
			Param(ws.PathParameter("namespace-name", "The name of the namespace to lookup").DataType("string")).
			Param(ws.PathParameter("pod-name", "The name of the pod to lookup").DataType("string")).
			Param(ws.PathParameter("metric-name", "The name of the requested metric").DataType("string")).
			Param(ws.QueryParameter("start", "Start time for requested metrics").DataType("string")).
			Param(ws.QueryParameter("end", "End time for requested metric").DataType("string")).
			Writes(types.MetricResult{}))

		// The /namespaces/{namespace-name}/pods/{pod-name}/containers/metrics/{container-name}/metrics endpoint
		// returns a list of all available metrics for a Pod Container entity.
		ws.Route(ws.GET("/namespaces/{namespace-name}/pods/{pod-name}/containers/{container-name}/metrics").
			To(a.availablePodContainerMetrics).
			Doc("Get a list of all available metrics for a Pod entity").
			Operation("availableMetrics").
			Param(ws.PathParameter("namespace-name", "The name of the namespace to lookup").DataType("string")).
			Param(ws.PathParameter("pod-name", "The name of the pod to lookup").DataType("string")).
			Param(ws.PathParameter("container-name", "The name of the namespace to use").DataType("string")))

		// The /namespaces/{namespace-name}/pods/{pod-name}/containers/{container-name}/metrics/{metric-name} endpoint exposes
		// a metric for a Container entity of the model.
		ws.Route(ws.GET("/namespaces/{namespace-name}/pods/{pod-name}/containers/{container-name}/metrics/{metric-name}").
			To(a.podContainerMetrics).
			Doc("Export an aggregated metric for a Pod Container").
			Operation("podContainerMetrics").
			Param(ws.PathParameter("namespace-name", "The name of the namespace to use").DataType("string")).
			Param(ws.PathParameter("pod-name", "The name of the pod to use").DataType("string")).
			Param(ws.PathParameter("container-name", "The name of the namespace to use").DataType("string")).
			Param(ws.PathParameter("metric-name", "The name of the requested metric").DataType("string")).
			Param(ws.QueryParameter("start", "Start time for requested metrics").DataType("string")).
			Param(ws.QueryParameter("end", "End time for requested metric").DataType("string")).
			Writes(types.MetricResult{}))
	}

	// The /nodes/{node-name}/freecontainers/{container-name}/metrics endpoint
	// returns a list of all available metrics for a Free Container entity.
	ws.Route(ws.GET("/nodes/{node-name}/freecontainers/{container-name}/metrics").
		To(a.availableFreeContainerMetrics).
		Doc("Get a list of all available metrics for a free Container entity").
		Operation("availableMetrics").
		Param(ws.PathParameter("node-name", "The name of the namespace to lookup").DataType("string")).
		Param(ws.PathParameter("container-name", "The name of the namespace to use").DataType("string")))

	// The /nodes/{node-name}/freecontainers/{container-name}/metrics/{metric-name} endpoint exposes
	// a metric for a free Container entity of the model.
	ws.Route(ws.GET("/nodes/{node-name}/freecontainers/{container-name}/metrics/{metric-name}").
		To(a.freeContainerMetrics).
		Doc("Export a container-level metric for a free container").
		Operation("freeContainerMetrics").
		Param(ws.PathParameter("node-name", "The name of the node to use").DataType("string")).
		Param(ws.PathParameter("container-name", "The name of the container to use").DataType("string")).
		Param(ws.PathParameter("metric-name", "The name of the requested metric").DataType("string")).
		Param(ws.QueryParameter("start", "Start time for requested metrics").DataType("string")).
		Param(ws.QueryParameter("end", "End time for requested metric").DataType("string")).
		Writes(types.MetricResult{}))

	if a.runningInKubernetes {
		// The /namespaces/{namespace-name}/pod-list/{pod-list}/metrics/{metric-name} endpoint exposes
		// metrics for a list od pods of the model.
		ws.Route(ws.GET("/namespaces/{namespace-name}/pod-list/{pod-list}/metrics/{metric-name}").
			To(a.podListMetrics).
			Doc("Export a metric for all pods from the given list").
			Operation("podListMetric").
			Param(ws.PathParameter("namespace-name", "The name of the namespace to lookup").DataType("string")).
			Param(ws.PathParameter("pod-list", "Comma separated list of pod names to lookup").DataType("string")).
			Param(ws.PathParameter("metric-name", "The name of the requested metric").DataType("string")).
			Param(ws.QueryParameter("start", "Start time for requested metrics").DataType("string")).
			Param(ws.QueryParameter("end", "End time for requested metric").DataType("string")).
			Writes(types.MetricResult{}))
	}

	container.Add(ws)
}

// availableMetrics returns a list of available cluster metric names.
func (a *Api) availableClusterMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricNamesRequest(core.ClusterKey(), response)
}

// availableMetrics returns a list of available node metric names.
func (a *Api) availableNodeMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricNamesRequest(core.NodeKey(request.PathParameter("node-name")), response)
}

// availableMetrics returns a list of available namespace metric names.
func (a *Api) availableNamespaceMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricNamesRequest(core.NamespaceKey(request.PathParameter("namespace-name")), response)
}

// availableMetrics returns a list of available pod metric names.
func (a *Api) availablePodMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricNamesRequest(
		core.PodKey(request.PathParameter("namespace-name"),
			request.PathParameter("pod-name")), response)
}

// availableMetrics returns a list of available pod metric names.
func (a *Api) availablePodContainerMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricNamesRequest(
		core.PodContainerKey(request.PathParameter("namespace-name"),
			request.PathParameter("pod-name"),
			request.PathParameter("container-name"),
		), response)
}

// availableMetrics returns a list of available pod metric names.
func (a *Api) availableFreeContainerMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricNamesRequest(
		core.NodeContainerKey(request.PathParameter("node-name"),
			request.PathParameter("container-name"),
		), response)
}

// clusterMetrics returns a metric timeseries for a metric of the Cluster entity.
func (a *Api) clusterMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricRequest(core.ClusterKey(), request, response)
}

// nodeMetrics returns a metric timeseries for a metric of the Node entity.
func (a *Api) nodeMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricRequest(core.NodeKey(request.PathParameter("node-name")),
		request, response)
}

// namespaceMetrics returns a metric timeseries for a metric of the Namespace entity.
func (a *Api) namespaceMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricRequest(core.NamespaceKey(request.PathParameter("namespace-name")),
		request, response)
}

// podMetrics returns a metric timeseries for a metric of the Pod entity.
func (a *Api) podMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricRequest(
		core.PodKey(request.PathParameter("namespace-name"),
			request.PathParameter("pod-name")),
		request, response)
}

func (a *Api) podListMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	start, end, err := getStartEndTime(request)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	ns := request.PathParameter("namespace-name")
	keys := []string{}
	for _, podName := range strings.Split(request.PathParameter("pod-list"), ",") {
		keys = append(keys, core.PodKey(ns, podName))
	}
	metricName := getMetricName(request)
	if metricName == "" {
		response.WriteError(http.StatusBadRequest, fmt.Errorf("Metric not supported: %v", request.PathParameter("metric-name")))
		return
	}
	metrics := a.metricSink.GetMetric(metricName, keys, start, end)
	result := types.MetricResultList{
		Items: make([]types.MetricResult, 0, len(keys)),
	}
	for _, key := range keys {
		result.Items = append(result.Items, exportTimestampedMetricValue(metrics[key]))
	}
	response.WriteEntity(result)
}

// podContainerMetrics returns a metric timeseries for a metric of a Pod Container entity.
// podContainerMetrics uses the namespace-name/pod-name/container-name path.
func (a *Api) podContainerMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricRequest(
		core.PodContainerKey(request.PathParameter("namespace-name"),
			request.PathParameter("pod-name"),
			request.PathParameter("container-name"),
		),
		request, response)
}

// freeContainerMetrics returns a metric timeseries for a metric of the Container entity.
// freeContainerMetrics addresses only free containers, by using the node-name/container-name path.
func (a *Api) freeContainerMetrics(request *restful.Request, response *restful.Response) {

	//number of http model api requests add 1
	core.ModelApiRequestCount.Inc()
	a.processMetricRequest(
		core.NodeContainerKey(request.PathParameter("node-name"),
			request.PathParameter("container-name"),
		),
		request, response)
}

// parseRequestParam parses a time.Time from a named QueryParam, using the RFC3339 format.
func parseTimeParam(queryParam string, defaultValue time.Time) (time.Time, error) {
	if queryParam != "" {
		reqStamp, err := time.Parse(time.RFC3339, queryParam)
		if err != nil {
			return time.Time{}, fmt.Errorf("timestamp argument cannot be parsed: %s", err)
		}
		return reqStamp, nil
	}
	return defaultValue, nil
}

func (a *Api) processMetricRequest(key string, request *restful.Request, response *restful.Response) {
	start, end, err := getStartEndTime(request)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}
	metricName := getMetricName(request)
	if metricName == "" {
		response.WriteError(http.StatusBadRequest, fmt.Errorf("Metric not supported: %v", request.PathParameter("metric-name")))
		return
	}
	metrics := a.metricSink.GetMetric(metricName, []string{key}, start, end)
	converted := exportTimestampedMetricValue(metrics[key])
	response.WriteEntity(converted)
}

func (a *Api) processMetricNamesRequest(key string, response *restful.Response) {
	metricNames := a.metricSink.GetMetricNames(key)
	response.WriteEntity(metricNames)
}

func getMetricName(request *restful.Request) string {
	param := request.PathParameter("metric-name")
	return metricNamesConversion[param]
}

func getStartEndTime(request *restful.Request) (time.Time, time.Time, error) {
	start, err := parseTimeParam(request.QueryParameter("start"), time.Time{})
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseTimeParam(request.QueryParameter("end"), time.Now())
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func exportTimestampedMetricValue(values []metricsink.TimestampedMetricValue) types.MetricResult {
	result := types.MetricResult{
		Metrics: make([]types.MetricPoint, 0, len(values)),
	}
	for _, value := range values {
		if result.LatestTimestamp.Before(value.Timestamp) {
			result.LatestTimestamp = value.Timestamp
		}
		// TODO: clean up types in model api
		var intValue int64
		if value.ValueType == core.ValueInt64 {
			intValue = value.IntValue
		} else {
			intValue = int64(value.FloatValue)
		}

		result.Metrics = append(result.Metrics, types.MetricPoint{
			Timestamp: value.Timestamp,
			Value:     uint64(intValue),
		})
	}
	return result
}
