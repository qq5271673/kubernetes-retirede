{
  "apiVersion": "v1beta1",
  "kind": "Service",
  "id": "monitoring-influxdb",
  "port": 80,
  "containerPort": 8086,
  "selector": {
    "name": "influxGrafana"
  }
}