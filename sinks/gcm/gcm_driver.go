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

package gcm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/gcloud-golang/compute/metadata"
)

type MetricType int

const (
	// A cumulative metric.
	MetricCumulative MetricType = iota

	// An instantaneous value metric.
	MetricGauge
)

func (self MetricType) String() string {
	switch self {
	case MetricCumulative:
		return "cumulative"
	case MetricGauge:
		return "gauge"
	}
	return ""
}

type MetricValueType int

const (
	// An int64 value.
	ValueInt64 MetricValueType = iota
	// A boolean value
	ValueBool
	// A double-precision floating point number.
	ValueDouble
)

func (self MetricValueType) String() string {
	switch self {
	case ValueInt64:
		return "int64"
	case ValueBool:
		return "bool"
	case ValueDouble:
		return "double"
	}
	return ""
}

type LabelDescriptor struct {
	// Key to use for the label.
	Key string `json:"key,omitempty"`

	// Description of the label.
	Description string `json:"description,omitempty"`
}

type MetricDescriptor struct {
	// The unique name of the metric.
	Name string

	// Description of the metric.
	Description string

	// Descriptor of the labels used by this metric.
	Labels []LabelDescriptor

	// Type and value of metric data.
	Type      MetricType
	ValueType MetricValueType
}

type Metric struct {
	// The name of the metric. Must match an existing descriptor.
	Name string

	// The labels and values for the metric. The keys must match those in the descriptor.
	Labels map[string]string

	// The start and end time for which this data is representative.
	Start time.Time
	End   time.Time

	// The value of the metric. Must match the type in the descriptor.
	Value interface{}
}

type gcmDriver struct {
	// Token to use for authentication.
	token string

	// When the token expires.
	tokenExpiration time.Time

	// TODO(vmarmol): Make this configurable and not only detected.
	// GCE project.
	project string

	// TODO(vmarmol): Also store labels?
	// Map of metrics we currently export.
	exportedMetrics map[string]MetricDescriptor
}

// Returns a thread-compatible implementation of GCM interactions.
func NewDriver() (*gcmDriver, error) {
	// Only support GCE for now.
	if !metadata.OnGCE() {
		return nil, fmt.Errorf("the GCM sink is currently only supported on GCE")
	}

	// Detect project.
	project, err := metadata.ProjectID()
	if err != nil {
		return nil, err
	}

	// Check required service accounts
	err = checkServiceAccounts()
	if err != nil {
		return nil, err
	}

	impl := &gcmDriver{
		project:         project,
		exportedMetrics: make(map[string]MetricDescriptor),
	}

	// Get an initial token.
	err = impl.refreshToken()
	if err != nil {
		return nil, err
	}

	return impl, nil
}

func (self *gcmDriver) refreshToken() error {
	if time.Now().After(self.tokenExpiration) {
		token, err := getToken()
		if err != nil {
			return nil
		}

		// Expire the token a bit early.
		const earlyRefreshSeconds = 60
		if token.ExpiresIn > earlyRefreshSeconds {
			token.ExpiresIn -= earlyRefreshSeconds
		}
		self.token = token.AccessToken
		self.tokenExpiration = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}
	return nil
}

// GCM request structures for a MetricDescriptor.
type typeDescriptor struct {
	MetricType string `json:"metricType,omitempty"`
	ValueType  string `json:"valueType,omitempty"`
}

type metricDescriptor struct {
	Name           string            `json:"name,omitempty"`
	Project        string            `json:"project,omitempty"`
	Description    string            `json:"description,omitempty"`
	Labels         []LabelDescriptor `json:"labels,omitempty"`
	TypeDescriptor typeDescriptor    `json:"typeDescriptor,omitempty"`
}

const maxNumLabels = 10

// Adds the specified metrics or updates them if they already exist.
func (self *gcmDriver) AddMetrics(metrics []MetricDescriptor) error {
	for _, metric := range metrics {
		// Enforce the most labels that GCM allows.
		if len(metric.Labels) > maxNumLabels {
			return fmt.Errorf("metrics cannot have more than %d labels and %q has %d", maxNumLabels, metric.Name, len(metric.Labels))
		}

		// Ensure all labels are in the correct format.
		for i := range metric.Labels {
			metric.Labels[i].Key = fullLabelName(metric.Labels[i].Key)
		}

		request := metricDescriptor{
			Name:        fullMetricName(metric.Name),
			Project:     self.project,
			Description: metric.Description,
			Labels:      metric.Labels,
			TypeDescriptor: typeDescriptor{
				MetricType: metric.Type.String(),
				ValueType:  metric.ValueType.String(),
			},
		}

		err := sendRequest(fmt.Sprintf("https://www.googleapis.com/cloudmonitoring/v2beta2/projects/%s/metricDescriptors", self.project), self.token, request)
		if err != nil {
			return err
		}

		// Add metric to exportedMetrics.
		self.exportedMetrics[metric.Name] = metric
	}

	return nil
}

// GCM request structures for writing time-series data.
type timeseriesDescriptor struct {
	Project string            `json:"project,omitempty"`
	Metric  string            `json:"metric,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type point struct {
	Start      time.Time `json:"start,omitempty"`
	End        time.Time `json:"end,omitempty"`
	Int64Value int64     `json:"int64Value"`
}

type timeseries struct {
	TimeseriesDescriptor timeseriesDescriptor `json:"timeseriesDesc,omitempty"`
	Point                point                `json:"point,omitempty"`
}

type metricWriteRequest struct {
	Timeseries []timeseries `json:"timeseries,omitempty"`
}

// The largest number of timeseries we can write to per request.
const maxTimeseriesPerRequest = 200

// Pushes the specified metric values. The metrics must already exist.
func (self *gcmDriver) PushMetrics(metrics []Metric) error {
	// Check we're not being asked to write more timeseries than we can..
	if len(metrics) > maxTimeseriesPerRequest {
		return fmt.Errorf("unable to write more than %d metrics at once and %d were provided", maxTimeseriesPerRequest, len(metrics))
	}

	// Ensure the metrics exist.
	for _, metric := range metrics {
		if _, ok := self.exportedMetrics[metric.Name]; !ok {
			return fmt.Errorf("unable to push unknown metric %q", metric.Name)
		}
	}

	// Push the metrics.
	var request metricWriteRequest
	for _, metric := range metrics {
		// Use full label names.
		labels := make(map[string]string, len(metric.Labels))
		for key, value := range metric.Labels {
			labels[fullLabelName(key)] = value
		}

		// TODO(vmarmol): Validation and cleanup of data.
		// TODO(vmarmol): Handle non-int64 data types. There is an issue with using omitempty since 0 is a valid value for us.
		if _, ok := metric.Value.(int64); !ok {
			return fmt.Errorf("non-int64 data not implemented. Seen for metric %q", metric.Name)
		}
		request.Timeseries = append(request.Timeseries, timeseries{
			TimeseriesDescriptor: timeseriesDescriptor{
				Metric: fullMetricName(metric.Name),
				Labels: labels,
			},
			Point: point{
				Start:      metric.Start,
				End:        metric.End,
				Int64Value: metric.Value.(int64),
			},
		})
	}

	// Refresh token.
	err := self.refreshToken()
	if err != nil {
		return err
	}

	return sendRequest(fmt.Sprintf("https://www.googleapis.com/cloudmonitoring/v2beta2/projects/%s/timeseries:write", self.project), self.token, request)
}

// Domain for the metrics.
const metricDomain = "kubernetes.io"

func fullLabelName(name string) string {
	if !strings.Contains(name, "custom.cloudmonitoring.googleapis.com/") {
		return fmt.Sprintf("custom.cloudmonitoring.googleapis.com/%s/label/%s", metricDomain, name)
	}
	return name
}

func fullMetricName(name string) string {
	if !strings.Contains(name, "custom.cloudmonitoring.googleapis.com/") {
		return fmt.Sprintf("custom.cloudmonitoring.googleapis.com/%s/%s", metricDomain, name)
	}
	return name
}

func sendRequest(url string, token string, request interface{}) error {
	rawRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(rawRequest))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("request to %q failed with status %q and response: %q", url, resp.Status, string(out))
	}

	return nil
}
