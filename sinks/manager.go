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

package sinks

import (
	"sync"
	"time"

	"github.com/golang/glog"
	. "k8s.io/heapster/core"
)

type DataSink interface {
	Name() string

	// Exports data to the external storge. The funciton should be synchronous/blocking and finish only
	// after the given DataBatch was written. This will allow sink manager to push data only to these
	// sinks that finished writing the previous data.
	ExportData(*DataBatch)
	Stop()
}

const (
	DefaultSinkExportDataTimeout = 20 * time.Second
	DefaultSinkStopTimeout       = 60 * time.Second
)

type sinkHolder struct {
	sink             DataSink
	dataBatchChannel chan *DataBatch
	stopChannel      chan bool
}

// Sink Manager - a special sink that distributes data to other sinks. It pushes data
// only to these sinks that completed their previous exports. Data that could not be
// pushed in the defined time is dropped and not retried.
type sinkManager struct {
	sinkHolders       []sinkHolder
	exportDataTimeout time.Duration
	stopTimeout       time.Duration
}

func NewDataSinkManager(sinks []DataSink, exportDataTimeout, stopTimeout time.Duration) (DataSink, error) {
	sinkHolders := []sinkHolder{}
	for _, sink := range sinks {
		sh := sinkHolder{
			sink:             sink,
			dataBatchChannel: make(chan *DataBatch),
			stopChannel:      make(chan bool),
		}
		sinkHolders = append(sinkHolders, sh)
		go func(sh sinkHolder) {
			for {
				select {
				case data := <-sh.dataBatchChannel:
					sh.sink.ExportData(data)
				case isStop := <-sh.stopChannel:
					glog.V(2).Infof("Stop received: %s", sh.sink.Name())
					if isStop {
						sh.sink.Stop()
						return
					}
				}
			}
		}(sh)
	}
	return &sinkManager{
		sinkHolders:       sinkHolders,
		exportDataTimeout: exportDataTimeout,
		stopTimeout:       stopTimeout,
	}, nil
}

// Guarantees that the export will complete in sinkExportDataTimeout.
func (this *sinkManager) ExportData(data *DataBatch) {
	var wg sync.WaitGroup
	for _, sh := range this.sinkHolders {
		wg.Add(1)
		go func(sh sinkHolder, wg *sync.WaitGroup) {
			defer wg.Done()
			glog.V(2).Infof("Pushing data to: %s", sh.sink.Name())
			select {
			case sh.dataBatchChannel <- data:
				glog.V(2).Infof("Data push completed: %s", sh.sink.Name())
				// everything ok
			case <-time.After(this.exportDataTimeout):
				glog.Warningf("Failed to push data to sink: %s", sh.sink.Name())
			}
		}(sh, &wg)
	}
	// Wait for all pushes to complete or timeout.
	wg.Wait()
}

func (this *sinkManager) Name() string {
	return "Manager"
}

func (this *sinkManager) Stop() {
	for _, sh := range this.sinkHolders {
		glog.V(2).Infof("Running stop for: %s", sh.sink.Name())

		go func(sh sinkHolder) {
			select {
			case sh.stopChannel <- true:
				// everything ok
				glog.V(2).Infof("Stop sent to sink: %s", sh.sink.Name())

			case <-time.After(this.stopTimeout):
				glog.Warningf("Failed to stop sink: %s", sh.sink.Name())
			}
			return
		}(sh)
	}
}
