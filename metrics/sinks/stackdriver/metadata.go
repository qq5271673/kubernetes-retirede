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

package stackdriver

var (
	// Known metrics metadata

	// Container metrics

	containerUptimeMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "DOUBLE",
		Name:       "kubernetes.io/container/uptime",
	}

	cpuContainerCoreUsageTimeMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "DOUBLE",
		Name:       "kubernetes.io/container/cpu/core_usage_time",
	}

	cpuRequestedCoresMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "DOUBLE",
		Name:       "kubernetes.io/container/cpu/requested_cores",
	}

	cpuLimitCoresMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "DOUBLE",
		Name:       "kubernetes.io/container/cpu/limit_cores",
	}

	memoryContainerUsedBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/container/memory/used_bytes",
	}

	memoryRequestedBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/container/memory/requested_bytes",
	}

	memoryLimitBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/container/memory/limit_bytes",
	}

	restartCountMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/container/restart_count",
	}

	// Pod metrics

	volumeUsedBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/pod/volume/used_bytes",
	}

	volumeRequestedBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/pod/volume/requested_bytes",
	}

	networkPodRxMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/pod/network/bytes_rx",
	}

	networkPodTxMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/pod/network/bytes_tx",
	}

	// Node metrics

	cpuNodeCoreUsageTimeMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "DOUBLE",
		Name:       "kubernetes.io/node/cpu/core_usage_time",
	}

	cpuTotalCoresMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "DOUBLE",
		Name:       "kubernetes.io/node/cpu/total_cores",
	}

	cpuAllocatableCoresMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "DOUBLE",
		Name:       "kubernetes.io/node/cpu/allocatable_cores",
	}

	memoryNodeUsedBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/node/memory/used_bytes",
	}

	memoryTotalBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/node/memory/total_bytes",
	}

	memoryAllocatableBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/node/memory/allocatable_bytes",
	}

	networkNodeRxMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/node/network/bytes_rx",
	}

	networkNodeTxMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/node/network/bytes_tx",
	}

	cpuNodeDaemonCoreUsageTimeMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "DOUBLE",
		Name:       "kubernetes.io/node_daemon/cpu/core_usage_time",
	}

	memoryNodeDaemonUsedBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "kubernetes.io/node_daemon/memory/used_bytes",
	}

	// Old resource model metrics

	legacyUptimeMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "DOUBLE",
		Name:       "container.googleapis.com/container/uptime",
	}

	legacyCPUReservedCoresMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "DOUBLE",
		Name:       "container.googleapis.com/container/cpu/reserved_cores",
	}

	legacyCPUUsageTimeMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "DOUBLE",
		Name:       "container.googleapis.com/container/cpu/usage_time",
	}

	legacyNetworkRxMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "INT64",
		Name:       "container.googleapis.com/container/network/received_bytes_count",
	}

	legacyNetworkTxMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "INT64",
		Name:       "container.googleapis.com/container/network/sent_bytes_count",
	}

	legacyMemoryLimitMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "container.googleapis.com/container/memory/bytes_total",
	}

	legacyMemoryBytesUsedMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "container.googleapis.com/container/memory/bytes_used",
	}

	legacyMemoryAllocatableBytesMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "container.googleapis.com/container/memory/allocatable_bytes",
	}

	legacyMemoryPageFaultsMD = &metricMetadata{
		MetricKind: "CUMULATIVE",
		ValueType:  "INT64",
		Name:       "container.googleapis.com/container/memory/page_fault_count",
	}

	legacyDiskBytesUsedMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "container.googleapis.com/container/disk/bytes_used",
	}

	legacyDiskBytesTotalMD = &metricMetadata{
		MetricKind: "GAUGE",
		ValueType:  "INT64",
		Name:       "container.googleapis.com/container/disk/bytes_total",
	}
)
