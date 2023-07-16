{
  "apiVersion": "v1beta1",
  "kind": "ReplicationController",
  "id": "monitoring-influx-grafana-controller",
  "desiredState": {
    "replicas": 1,
    "replicaSelector": {
      "name": "influxGrafana"
    },
    "podTemplate": {
      "labels": {
        "name": "influxGrafana",
        "kubernetes.io/cluster-service": "true"
      },
      "desiredState": {
        "manifest": {
          "version": "v1beta1",
          "id": "monitoring-influx-grafana-controller",
          "containers": [
            {
              "name": "influxdb",
              "image": "kubernetes/heapster_influxdb:v0.3",
              "ports": [
                {
                  "containerPort": 8083
                },
                {
                  "containerPort": 8086
                }
              ]
            },
            {
              "name": "grafana",
              "image": "kubernetes/heapster_grafana:v0.6",
              "ports": [
                {
                  "containerPort": 8080
                }
              ],
              "env": [
                {
                  "name": "INFLUXDB_EXTERNAL_URL",
                  "value": "/api/v1beta1/proxy/services/monitoring-grafana/db/"
                },
                {
                  "name": "INFLUXDB_HOST",
                  "value": "monitoring-influxdb"
                },
                {
                  "name": "INFLUXDB_PORT",
                  "value": "80"
                }
              ]
            }
          ]
        }
      }
    }
  },
  "labels": {
    "name": "influxGrafana"
  }
}