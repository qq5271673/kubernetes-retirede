package sinks

import (
	"fmt"
	"strings"

	bigquery "code.google.com/p/google-api-go-client/bigquery/v2"
	"github.com/GoogleCloudPlatform/heapster/sources"
	"github.com/golang/glog"
	cadvisor "github.com/google/cadvisor/info"
	bigquery_client "github.com/google/cadvisor/storage/bigquery/client"
)

// Big query related flags defined in bigquery_client
// clientId       = flag.String("bq_id", "", "Client ID")
// clientSecret   = flag.String("bq_secret", "notasecret", "Client Secret")
// projectId      = flag.String("bq_project_id", "", "Bigquery project ID")
// serviceAccount = flag.String("bq_account", "", "Service account email")
// pemFile        = flag.String("bq_credentials_file", "", "Credential Key file (pem)")

type bigquerySink struct {
	client *bigquery_client.Client
	rows   []map[string]interface{}
}

const (
	// Bigquery schema types
	typeTimestamp string = "TIMESTAMP"
	typeString    string = "STRING"
	typeInteger   string = "INTEGER"
)

// TODO(jnagal): Infer schema through reflection. (See bigquery/client/example)
func (self *bigquerySink) GetSchema() *bigquery.TableSchema {
	fields := make([]*bigquery.TableFieldSchema, 0)

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeTimestamp,
		Name: colTimestamp,
		Mode: "REQUIRED",
	})
	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeString,
		Name: colHostName,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeString,
		Name: colPodStatus,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeString,
		Name: colPodStatus,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeString,
		Name: colPodIP,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeString,
		Name: colLabels,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeString,
		Name: colContainerName,
		Mode: "REQUIRED",
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colCpuCumulativeUsage,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colCpuInstantUsage,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colMemoryUsage,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colMemoryWorkingSet,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colMemoryPgFaults,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colCpuInstantUsage,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colRxBytes,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colRxErrors,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colTxBytes,
	})

	fields = append(fields, &bigquery.TableFieldSchema{
		Type: typeInteger,
		Name: colTxErrors,
	})

	return &bigquery.TableSchema{
		Fields: fields,
	}
}

func (self *bigquerySink) containerStatsToValues(
	pod *sources.Pod,
	hostname,
	containerName string,
	stat *cadvisor.ContainerStats) (row map[string]interface{}) {
	row = make(map[string]interface{})

	// Timestamp
	row[colTimestamp] = stat.Timestamp.Unix()

	if pod != nil {
		// Pod name
		row[colPodName] = pod.Name

		// Pod Status
		row[colPodStatus] = pod.Status

		// Pod IP
		row[colPodIP] = pod.PodIP

		labels := []string{}
		for key, value := range pod.Labels {
			labels = append(labels, fmt.Sprintf("%s:%s", key, value))
		}
		row[colLabels] = strings.Join(labels, ",")
	}

	// Hostname
	row[colHostName] = hostname

	// Container name
	row[colContainerName] = containerName

	if stat.Cpu != nil {
		// Cumulative Cpu Usage
		row[colCpuCumulativeUsage] = stat.Cpu.Usage.Total
	}

	if stat.Memory != nil {
		// Memory Usage
		row[colMemoryUsage] = stat.Memory.Usage

		row[colMemoryPgFaults] = stat.Memory.ContainerData.Pgfault

		// Working set size
		row[colMemoryWorkingSet] = stat.Memory.WorkingSet
	}

	// Optional: Network stats.
	if stat.Network != nil {
		row[colRxBytes] = stat.Network.RxBytes
		row[colRxErrors] = stat.Network.RxErrors
		row[colTxBytes] = stat.Network.TxBytes
		row[colTxErrors] = stat.Network.TxErrors
	}

	return
}

func (self *bigquerySink) handlePods(pods []sources.Pod) {
	for _, pod := range pods {
		for _, container := range pod.Containers {
			for _, stat := range container.Stats {
				self.rows = append(self.rows, self.containerStatsToValues(&pod, pod.Hostname, container.Name, stat))
			}
		}
	}
}

func (self *bigquerySink) handleContainers(container sources.AnonContainer) {
	for _, stat := range container.Stats {
		self.rows = append(self.rows, self.containerStatsToValues(nil, container.Hostname, container.Name, stat))
	}
}

func (self *bigquerySink) StoreData(ip Data) error {
	if data, ok := ip.([]sources.Pod); ok {
		self.handlePods(data)
	} else if data, ok := ip.(sources.AnonContainer); ok {
		self.handleContainers(data)
	} else {
		return fmt.Errorf("Requesting unrecognized type to be stored in InfluxDB")
	}

	// TODO(vishh): Modify the big query client to take in a series of rows.
	for _, row := range self.rows {
		err := self.client.InsertRow(row)
		if err != nil {
			glog.Error(err)
		}
	}
	self.rows = self.rows[:0]
	return nil
}

// Create a new bigquery storage driver.
// machineName: A unique identifier to identify the host that current cAdvisor
// instance is running on.
// tableName: BigQuery table used for storing stats.
func NewBigQuerySink() (Sink, error) {
	bqClient, err := bigquery_client.NewClient()
	if err != nil {
		return nil, err
	}
	err = bqClient.CreateDataset("cadvisor")
	if err != nil {
		return nil, err
	}

	ret := &bigquerySink{
		client: bqClient,
		rows:   make([]map[string]interface{}, 0),
	}
	schema := ret.GetSchema()
	err = bqClient.CreateTable(statsTable, schema)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
