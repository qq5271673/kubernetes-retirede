Heapster
===========

Heapster collects resource usage of Pods running in a [Kubernetes](https://github.com/GoogleCloudPlatform/kubernetes) cluster.
Heapster is *demo app* that demonstrates one possible way of monitoring a Kubernetes cluster. It also serves to showcase the power of core Kubernetes concepts like labels and pods and the awesomeness that is [cAdvisor](https://https://github.com/google/cadvisor).

####How Heapster works:
1. Discovers all minions in a Kubernetes cluster
2. Collects container statistics from the cadvisors running on the minions
2. Organizes stats into Pods
3. Stores Pod stats in a configurable backend

Along with each container stat entry, it's Pod ID, Container name, Pod IP, Hostname and Labels are also stored. Labels are stored as key:value pairs.

Heapster currently supports in-memory and [InfluxDB](http://influxdb.com) backends. Patches are welcome for adding more storage backends.

####Run Heapster in a Kubernetes cluster along with an Influxdb backend and [Grafana](http://grafana.org/docs/features/influxdb).

**Step 1: Kube cluster**
Fork the Kubernetes repository and [turn up a Kubernetes cluster](https://github.com/GoogleCloudPlatform/kubernetes-new#contents), if you haven't already.

**Step 2: Start a Pod that contains Influxdb**
```shell
cluster/kubecfg.sh -c influx-grafana/deploy/grafana-influxdb-pod.json create pods
```

**Step 3: Start Influxdb service**
```shell
cluster/kubecfg.sh -c influx-grafana/deploy/grafana-influxdb-service.json create services
```

**Step 4: Update firewall rules**
Open up ports tcp:80,8083,8086,9200.
```shell
gcutil addfirewall --allowed=tcp:80,tcp:8083,tcp:8086,tcp:9200 --target_tags=kubernetes-minion heapster
```

**Step 5: Configure cluster information for heapster Pod**
Open deploy/heapster-pod.json and update the following environment variables:
- Set 'KUBE_MASTER' to the IP address of the master.
- Set 'KUBE_MASTER_AUTH' to be the the username and password of the master. The format is username:password.
```shell
cat ~/.kubernetes_auth
```

**Step 6: Start Heapster Pod**
```shell
cluster/kubecfg.sh -c deploy/heapster-pod.json create pods
```

Once that's up you can list the pods in the cluster, to verify that all the pods and services are up and running:

```shell
cluster/kubecfg.sh list pods
```

To start monitoring the cluster using grafana, find out the the external IP of the minion where the 'influx-grafana' Pod is running from the [Google Cloud Console][cloud-console] or the `gcutil` tool, and visit `http://<minion-ip>:80`. 

To access the Influxdb UI visit  `http://<minion-ip>:8083`.


```shell
$ gcutil listinstances
```
