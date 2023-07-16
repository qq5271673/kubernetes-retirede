{
  "apiVersion": "v1beta1",
  "id": "monitoring-heapster-controller",
  "kind": "ReplicationController",
  "desiredState": {
    "replicas": 1,
    "replicaSelector": {
      "name": "heapster"
    },
    "podTemplate": {
      "desiredState": {
        "manifest": {
          "version": "v1beta1",
          "id": "monitoring-heapster-controller",
          "containers": [
            {
              "name": "heapster",
              "image": "kubernetes/heapster:v0.10.0",
              "env": [
                {
                  "name": "SINK",
                  "value": "gcm"
                },
                {
                  "name": "FLAGS",
                  "value": "--poll_duration=2m --stats_resolution=1m"
                }
              ],
              "volumeMounts": [
                {
                  "name": "ssl-certs",
                  "mountPath": "/etc/ssl/certs",
                  "readOnly": true
                }
              ]
            }
          ],
          "volumes": [
            {
              "name": "ssl-certs",
              "source": {
                "hostDir": {
                  "path": "/etc/ssl/certs"
                }
              }
            }
          ]
        }
      },
      "labels": {
        "name": "heapster",
        "kubernetes.io/cluster-service": "true"
      }
    }
  },
  "labels": {
    "name": "heapster"
  }
}