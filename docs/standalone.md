Runnig Heapster standalone on a host with cAdvisor
================================

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
    - [Prepare Heapster binaries](#prepare-heapster-binaries)
    - [Running a Cadvisor](#running-a-cadvisor)
- [Starting a heapster](#starting-a-heapster)
    - [Config Cadvisor source](#config-cadvisor-source)
    - [Start heapster](#start-heapster)
    - [Test it out](#test-it-out)

## Introduction

This document describes how to to run Heapster standalone with [cAdvisor](https://github.com/google/cadvisor). These guide is tested OK on Ubuntu 14.04 LTS 64bit server. I'm working around testing this with some other system . The assumptions here are that [influxdb](https://github.com/influxdb/influxdb), and [grafana](https://github.com/grafana/grafana.git) run on one machine, which heapster can visit.

## Prerequisites

### Prepare Heapster binaries
We can get the heapster binary from the [Release Page](https://github.com/kubernetes/heapster/releases/) or build the heapster from source code.The build steps are as follows, [go](https://github.com/golang) and [godep](https://github.com/tools/godep) need be installed in system first:


**Step 1: Clone the heapster github repo**

``` 
$ git clone https://github.com/kubernetes/heapster.git
```

**Step 2: Build the source code**

``` 
$ cd kubernetes/heapster
$ make
```

Then you can get the heapster binary in current path.

### Running a cAdvisor
cAdvisor can run in a docker or standalone outside of Docker to monitor the whole machine. This guide run a cAdvisor standalone out of Docker.cAdvisor is a static Go binary with no external dependencies. To run it standalone all you should need to do is run it! Note that some data sources may require root priviledges.You can get cAdvisor binary from the [Release Page](https://github.com/google/cadvisor/releases)

``` 
$ ./cadvisor  
```

cAdvisor is now running (in the foreground) on http://localhost:8080/.

## Starting a heapster

### Config Cadvisor source
Heapster support two types of Cadvisor source: standalone & CoreOS; Doc of source configuration is [HERE](https://github.com/kubernetes/heapster/blob/master/docs/source-configuration.md)

External cadvisor source "discovers" hosts from the specified file. Use it like this:

```
--source=cadvisor:external[?<OPTIONS>]
```

The following options are available:

* `standalone` - only use `localhost` (default: `false`)
* `hostsFile` - file containing list of hosts to gather cadvisor metrics from (default: `/var/run/heapster/hosts`)
* `cadvisorPort` - cadvisor port to use (default: `8080`)

Here is an example:
```shell
$ ./heapster --source="cadvisor:external?cadvisorPort=4194"
```

Because heapster needs a specified file to config external cadvisor source, we need to provide one.
The `hostsFile` parameter defines a list of hosts to poll for metrics and must be in JSON format. See below for an example:

```
{
  "Items": [
    {
      "Name": "server-105",
      "IP": "192.168.99.105"
    },
    {
      "Name": "server-106",
      "IP": "192.168.99.105"
    }
  ]
}
```
### Start heapster
Now you can run a heapster, here is an example:

```
./heapster-master --source="cadvisor:external?cadvisorPort=8080" --use_model=true --model_resolution=10s --cache_duration=1m --stats_resolution=1s"
```

The `hostsFile` use the default one `/var/run/heapster/hosts`.Then the heapster is running on the default port 8082,like this:

```
I0911 13:42:03.657739   21041 heapster.go:57] ./heapster --source=cadvisor:external?cadvisorPort=8080 --use_model=true --model_resolution=10s --cache_duration=1m --stats_resolution=1s
I0911 13:42:03.657918   21041 heapster.go:58] Heapster version 0.17.0
I0911 13:42:03.662594   21041 heapster.go:68] Starting heapster on port 8082
```

###Test it out
If you got debug information as before, means that the heapster is working normally.You can test it by requeting its [Reatful api](https://github.com/kubernetes/heapster/blob/master/docs/model.md). You can use some restful tools to test, or just run a `curl` command, 

```

curl http://0.0.0.0:8082/api/v1/model/stats/ | gunzip

```

and you will get the response like this:

```
{
  "uptime": 2543160,
  "stats": {
   "cpu-limit": {
    "minute": {
     "average": 1000,
     "percentile": 1000,
     "max": 1000
    },
    "hour": {
     "average": 1000,
     "percentile": 1000,
     "max": 1000
    },
    "day": {
     "average": 1000,
     "percentile": 1000,
     "max": 1000
    }
   },
   "cpu-usage": {
    "minute": {
     "average": 10,
     "percentile": 10,
     "max": 10
    },
    "hour": {
     "average": 10,
     "percentile": 10,
     "max": 10
    },
    "day": {
     "average": 10,
     "percentile": 10,
     "max": 10
    }
   },
   "fs-limit-dev-disk-by-uuid-dcaa07b0-d2ad-4a32-bd61-6584d0da68c0": {
    "minute": {
     "average": 21103243264,
     "percentile": 21103243264,
     "max": 21103243264
    },
    "hour": {
     "average": 21103243300,
     "percentile": 21103243300,
     "max": 21103243300
    },
    "day": {
     "average": 21103243300,
     "percentile": 21103243300,
     "max": 21103243300
    }
   },
   "fs-usage-dev-disk-by-uuid-dcaa07b0-d2ad-4a32-bd61-6584d0da68c0": {
    "minute": {
     "average": 12974346240,
     "percentile": 12974346240,
     "max": 12974346240
    },
    "hour": {
     "average": 12974346300,
     "percentile": 12974346300,
     "max": 12974346300
    },
    "day": {
     "average": 12974346300,
     "percentile": 12974346300,
     "max": 12974346300
    }
   },
   "memory-limit": {
    "minute": {
     "average": 0,
     "percentile": 18446744073709551615,
     "max": 18446744073709551615
    },
    "hour": {
     "average": 1366425486941603612,
     "percentile": 12297829382474432512,
     "max": 18446744073709551615
    },
    "day": {
     "average": 1366425486941603612,
     "percentile": 12297829382474432512,
     "max": 18446744073709551615
    }
   },
   "memory-usage": {
    "minute": {
     "average": 1824296960,
     "percentile": 1824296960,
     "max": 1824296960
    },
    "hour": {
     "average": 1820327936,
     "percentile": 1820327936,
     "max": 1824296960
    },
    "day": {
     "average": 1820327936,
     "percentile": 1820327936,
     "max": 1824296960
    }
   },
   "memory-working": {
    "minute": {
     "average": 447021056,
     "percentile": 447021056,
     "max": 447021056
    },
    "hour": {
     "average": 444596224,
     "percentile": 444596224,
     "max": 447021056
    },
    "day": {
     "average": 444596224,
     "percentile": 444596224,
     "max": 447021056
    }
   }
  }
 }

```




