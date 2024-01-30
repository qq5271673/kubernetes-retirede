Runnning Heapster on CoreOS
================================

Heapster enables cluster monitoring in a CoreOS cluster using [cAdvisor](https://github.com/google/cadvisor). 

NOTE: Some of the following steps should be handled using fleet and unit files.

**Step 1: Start cAdvisor on all hosts by default**

gce-cloud-config-with-cadvisor.yml contains 'cadvisor.service' which can be specified as part of your cloud config to bring up cAdvisor by default on all hosts.

**Step 2: Start InfluxDB**

On a CoreOS machine start InfluxDB and grafana

```shell
$ docker run -d -p 8083:8083 -p 8086:8086 --name influxdb kubernetes/heapster_influxdb:v0.3
```

**Step 3: Start heapster**

Pass the host where heapster is running via the 'INFLUXDB_HOST' environment variable.

```shell
$ docker run --name heapster --link influxdb:influxdb -d kubernetes/heapster:v0.13.0 --source="cadvisor:coreos?fleetEndpoint=http://127.0.0.1:4001&cadvisorPort=8080" --sink="influxdb:http://influxdb:8086"
```

**Step 4: Start Grafana**

```
docker run -d -p 80:80 -e INFLUXDB_HOST=<host_ip> kubernetes/heapster_grafana:v0.7
```

### Unit files for the various components.
cadvisor (globally deployed)
```
[Unit]
Description=cAdvisor Service
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=10m
Restart=always
ExecStartPre=-/usr/bin/docker kill cadvisor
ExecStartPre=-/usr/bin/docker rm -f cadvisor
ExecStartPre=/usr/bin/docker pull google/cadvisor
ExecStart=/usr/bin/docker run --volume=/:/rootfs:ro --volume=/var/run:/var/run:rw --volume=/sys:/sys:ro --volume=/var/lib/docker/:/var/lib/docker:ro --publish=4194:4194 --name=cadvisor --net=host google/cadvisor:latest --logtostderr --port=4194
ExecStop=/usr/bin/docker stop -t 2 cadvisor

[X-Fleet]
Global=true
MachineMetadata=role=kubernetes
```
