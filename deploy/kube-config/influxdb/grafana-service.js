{
  "apiVersion": "v1beta1",
  "kind": "Service",
  "id": "monitoring-grafana",
  "port": 80,
  "containerPort": 8080,
  "labels": {
    "name": "monitoring-grafana",
    "kubernetes.io/cluster-service": "true"
  },
  "selector": {
    "name": "influxGrafana"
  }
}