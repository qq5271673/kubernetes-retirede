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

package processors

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/heapster/metrics/core"
)

type NodeAggregator struct {
	MetricsToAggregate []string
}

func (this *NodeAggregator) Name() string {
	return "node_aggregator"
}

func (this *NodeAggregator) Process(batch *core.DataBatch) (*core.DataBatch, error) {
	result := core.DataBatch{
		Timestamp:  batch.Timestamp,
		MetricSets: make(map[string]*core.MetricSet),
	}

	for key, metricSet := range batch.MetricSets {
		result.MetricSets[key] = metricSet
		if metricSetType, found := metricSet.Labels[core.LabelMetricSetType.Key]; found && metricSetType == core.MetricSetTypePod {
			// Aggregating pods
			if nodeName, found := metricSet.Labels[core.LabelNodename.Key]; found {
				nodeKey := core.NodeKey(nodeName)
				node, found := result.MetricSets[nodeKey]
				if !found {
					if node, found = batch.MetricSets[nodeKey]; !found {
						glog.Warningf("Failed to find node: %s", nodeKey)
						continue
					}
				}
				if err := aggregate(metricSet, node, this.MetricsToAggregate); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("No node info in pod %s: %v", key, metricSet.Labels)
			}
		}
	}

	return &result, nil
}
