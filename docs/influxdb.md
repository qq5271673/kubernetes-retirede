# Run Heapster in a Kubernetes cluster with an InfluxDB backend and a Grafana UI

_Warning: Virtual machines need to have at least 2 cores for InfluxDB to perform optimally._

### Setup a Kubernetes cluster
[Bring up a Kubernetes cluster](https://github.com/kubernetes/kubernetes), if you haven't already. Ensure that `kubecfg.sh` is exported.

### Start all of the pods and services
```shell
$ kubectl.sh create -f deploy/kube-config/influxdb/
```

Grafana service by default requests for a LoadBalancer. If that is not available in your cluster, consider changing that to NodePort. Use the external IP assigned to the Grafana service,
to access Grafana.
The default user name and password is 'admin'.
Once you login to Grafana, add a datasource that is InfluxDB. The URL for InfluxDB will be http://<Influxdb service name>:8086. Database name is 'k8s'. Default user name and password is 'root'.

Take a look at the [storage schema](./storage-schema.md) to understand how metrics are stored in InfluxDB.

## Troubleshooting guide
1. If the Grafana service is not accessible, chances are it might not be running. Use `kubectl.sh` to verify that the `heapster` and `influxdb & grafana` pods are alive.

	```shell
$ kubectl.sh get pods
	```

	```shell
$ kubectl.sh get services
	```

2. If the default Grafana dashboard doesn't show any graphs, check the Heapster logs with `kubectl.sh log <heapster pod name>`. Look for any errors related to accessing the Kubernetes master or the kubelet.

3. To access the InfluxDB UI, you will have to make the InfluxDB service externally visible, similar to how Grafana is made publicly accessible.

4. If you find InfluxDB to be using up a lot of CPU or memory, consider placing resource restrictions on the `InfluxDB & Grafana` pod. You can add `cpu: <millicores>` and `memory: <bytes>` in the [Controller Spec](../deploy/kube-config/influxdb/influxdb-grafana-controller.json) and relaunch the controllers:

	```shell
$ deploy/kube.sh restart
	```
