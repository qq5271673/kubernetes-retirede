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

//go:generate ./hooks/run_extpoints.sh

package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/heapster/common/flags"
	"k8s.io/heapster/metrics/core"
	"k8s.io/heapster/metrics/manager"
	"k8s.io/heapster/metrics/processors"
	"k8s.io/heapster/metrics/sinks"
	"k8s.io/heapster/metrics/sources"
	"k8s.io/heapster/version"
)

var (
	argMetricResolution = flag.Duration("metric_resolution", 30*time.Second, "The resolution at which heapster will retain metrics.")
	argPort             = flag.Int("port", 8082, "port to listen to")
	argIp               = flag.String("listen_ip", "", "IP to listen on, defaults to all IPs")
	argMaxProcs         = flag.Int("max_procs", 0, "max number of CPUs that can be used simultaneously. Less than 1 for default (number of cores)")
	argTLSCertFile      = flag.String("tls_cert", "", "file containing TLS certificate")
	argTLSKeyFile       = flag.String("tls_key", "", "file containing TLS key")
	argTLSClientCAFile  = flag.String("tls_client_ca", "", "file containing TLS client CA for client cert validation")
	argAllowedUsers     = flag.String("allowed_users", "", "comma-separated list of allowed users")
	argSources          flags.Uris
	argSinks            flags.Uris
)

func main() {
	defer glog.Flush()
	flag.Var(&argSources, "source", "source(s) to watch")
	flag.Var(&argSinks, "sink", "external sink(s) that receive data")
	flag.Parse()
	setMaxProcs()
	glog.Infof(strings.Join(os.Args, " "))
	glog.Infof("Heapster version %v", version.HeapsterVersion)
	if err := validateFlags(); err != nil {
		glog.Fatal(err)
	}

	// sources
	if len(argSources) != 1 {
		glog.Fatal("wrong number of sources specified")
	}
	sourceFactory := sources.NewSourceFactory()
	sourceProvider, err := sourceFactory.BuildAll(argSources)
	if err != nil {
		glog.Fatal(err)
	}
	sourceManager, err := sources.NewSourceManager(sourceProvider, sources.DefaultMetricsScrapeTimeout)
	if err != nil {
		glog.Fatal(err)
	}

	// sinks
	sinksFactory := sinks.NewSinkFactory()
	metricSink, sinkList := sinksFactory.BuildAll(argSinks)
	if metricSink == nil {
		glog.Fatalf("Failed to create metric sink")
	}
	for _, sink := range sinkList {
		glog.Infof("Starting with %s sink", sink.Name())
	}
	sinkManager, err := sinks.NewDataSinkManager(sinkList, sinks.DefaultSinkExportDataTimeout, sinks.DefaultSinkStopTimeout)
	if err != nil {
		glog.Fatal(err)
	}

	// data processors
	metricsToAggregate := []string{
		core.MetricCpuUsageRate.Name,
		core.MetricMemoryUsage.Name,
		core.MetricCpuRequest.Name,
		core.MetricCpuLimit.Name,
		core.MetricMemoryRequest.Name,
		core.MetricMemoryLimit.Name,
	}

	dataProcessors := []core.DataProcessor{}

	// pod enricher goes first
	if url, err := getKubernetesAddress(argSources); err == nil {
		podBasedEnricher, err := processors.NewPodBasedEnricher(url)
		if err != nil {
			glog.Fatalf("Failed to create PodBasedEnricher: %v", err)
		} else {
			dataProcessors = append(dataProcessors, podBasedEnricher)
		}

		namespaceBasedEnricher, err := processors.NewNamespaceBasedEnricher(url)
		if err != nil {
			glog.Fatalf("Failed to create NamespaceBasedEnricher: %v", err)
		} else {
			dataProcessors = append(dataProcessors, namespaceBasedEnricher)
		}
	}

	// then aggregators
	dataProcessors = append(dataProcessors,
		&processors.PodAggregator{},
		&processors.NamespaceAggregator{
			MetricsToAggregate: metricsToAggregate,
		},
		&processors.NodeAggregator{
			MetricsToAggregate: metricsToAggregate,
		},
		&processors.ClusterAggregator{
			MetricsToAggregate: metricsToAggregate,
		})

	// pod enricher goes first
	if url, err := getKubernetesAddress(argSources); err == nil {
		nodeAutoscalingEnricher, err := processors.NewNodeAutoscalingEnricher(url)
		if err != nil {
			glog.Fatalf("Failed to create NodeAutoscalingEnricher: %v", err)
		} else {
			dataProcessors = append(dataProcessors, nodeAutoscalingEnricher)
		}
	}

	// main manager
	manager, err := manager.NewManager(sourceManager, dataProcessors, sinkManager, *argMetricResolution,
		manager.DefaultScrapeOffset, manager.DefaultMaxParallelism)
	if err != nil {
		glog.Fatal(err)
	}
	manager.Start()

	handler := setupHandlers(metricSink)
	addr := fmt.Sprintf("%s:%d", *argIp, *argPort)
	glog.Infof("Starting heapster on port %d", *argPort)

	mux := http.NewServeMux()
	promHandler := prometheus.Handler()
	if len(*argTLSCertFile) > 0 && len(*argTLSKeyFile) > 0 {
		if len(*argTLSClientCAFile) > 0 {
			authPprofHandler, err := newAuthHandler(handler)
			if err != nil {
				glog.Fatal(err)
			}
			handler = authPprofHandler

			authPromHandler, err := newAuthHandler(promHandler)
			if err != nil {
				glog.Fatal(err)
			}
			promHandler = authPromHandler
		}
		mux.Handle("/", handler)
		mux.Handle("/metrics", promHandler)
		glog.Fatal(http.ListenAndServeTLS(addr, *argTLSCertFile, *argTLSKeyFile, mux))
	} else {
		mux.Handle("/", handler)
		mux.Handle("/metrics", promHandler)
		glog.Fatal(http.ListenAndServe(addr, mux))
	}
}

func getKubernetesAddress(args flags.Uris) (*url.URL, error) {
	for _, uri := range args {
		if uri.Key == "kubernetes" {
			return &uri.Val, nil
		}
	}
	return nil, fmt.Errorf("No kubernetes source found.")
}

func validateFlags() error {
	if *argMetricResolution < 5*time.Second {
		return fmt.Errorf("metric resolution needs to be greater than 5 seconds - %d", *argMetricResolution)
	}
	if (len(*argTLSCertFile) > 0 && len(*argTLSKeyFile) == 0) || (len(*argTLSCertFile) == 0 && len(*argTLSKeyFile) > 0) {
		return fmt.Errorf("both TLS certificate & key are required to enable TLS serving")
	}
	if len(*argTLSClientCAFile) > 0 && len(*argTLSCertFile) == 0 {
		return fmt.Errorf("client cert authentication requires TLS certificate & key")
	}
	return nil
}

func setMaxProcs() {
	// Allow as many threads as we have cores unless the user specified a value.
	var numProcs int
	if *argMaxProcs < 1 {
		numProcs = runtime.NumCPU()
	} else {
		numProcs = *argMaxProcs
	}
	runtime.GOMAXPROCS(numProcs)

	// Check if the setting was successful.
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != numProcs {
		glog.Warningf("Specified max procs of %d but using %d", numProcs, actualNumProcs)
	}
}
