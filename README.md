# Consulate [![Build Status](https://travis-ci.org/kadaan/consulate.svg?branch=master)](https://travis-ci.org/kadaan/consulate)

Consulate provides a normalized HTTP endpoint that responds with
200 / non-200 according to Consul checks specified in the requested URL.
This endpoint can then be composed with existing HTTP-URL-monitoring tools like
AWS ELB health checks to enable service QoS monitoring based on Consul checks.

### Building from source

To build Consulate from the source code yourself you need to have a working
Go environment with [version 1.9 or greater installed](http://golang.org/doc/install).

    $ mkdir -p $GOPATH/src/github.com/kadaan
    $ cd $GOPATH/src/github.com/kadaan
    $ git clone https://github.com/kadaan/consulate.git
    $ cd consulate
    $ ./build.sh

### Routes

All routes accept the following query string parameters:

1. `pretty`: when present, pretty prints json responses
1. `verbose`: when present, additional details are include in responses

#### /about

The `/about` route returns detailed version information about Consulate.

##### Request

```bash
curl -X GET http:/localhost:8080/about
```

##### Response

```json
{"Version":"0.0.1","Revision":"ea16a636cc89bec04f36cee1799265485c72cac6","Branch":"master","BuildUser":"jbaranick@ensadmins-MacBook-Pro.local","BuildDate":"2018-02-12T10:17:03Z","GoVersion":"go1.9"}
```

#### /health

The `/health` route returns 200 if Consulate is running and able to communicate with Consul.  Otherwise, a non-200 status code is returned.

##### Request

```bash
curl -X GET http:/localhost:8080/health
```

##### Response

```json
{"Status":"Ok"}
```

#### /metrics

The `/metrics` route returns Prometheus metrics.

##### Request

```bash
curl -X GET http:/localhost:8080/metrics
```

##### Response

```
# HELP consulate_request_duration_seconds The HTTP request latencies in seconds.
# TYPE consulate_request_duration_seconds histogram
consulate_request_duration_seconds_bucket{code="200",method="GET",url="/about",le="0.5"} 1
consulate_request_duration_seconds_bucket{code="200",method="GET",url="/about",le="1"} 1
consulate_request_duration_seconds_bucket{code="200",method="GET",url="/about",le="2"} 1
consulate_request_duration_seconds_bucket{code="200",method="GET",url="/about",le="3"} 1
...
```

#### /verify/checks

The `/verify/checks` route returns 200 if all Consul checks are in the PASSING state.  Otherwise, a non-200 status code is returned and the failing checks will be in the response.

##### Request

```bash
curl -X GET http:/localhost:8080/verify/checks
```

##### Response

```json
{"Status":"Ok"}
```

```json
{"Status":"Failed","Checks":{"port_check":{"Node":"macbook-pro.local","CheckID":"port_check","Name":"port check","Status":"critical","Notes":"Port Check","Output":"Timed out (1s) running check","ServiceID":"test","ServiceName":"test","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}
```

#### /verify/checks/:check

The `/verify/checks/:check` route returns 200 if the specified Consul check is in the PASSING state.  Otherwise, a non-200 status code is returned and the failing check will be in the response.

##### Request

```bash
curl -X GET http:/localhost:8080/verify/checks/port_check
```

##### Response

```json
{"Status":"Ok"}
```

```json
{"Status":"Failed","Checks":{"port_check":{"Node":"macbook-pro.local","CheckID":"port_check","Name":"port check","Status":"critical","Notes":"Port Check","Output":"Timed out (1s) running check","ServiceID":"test","ServiceName":"test","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}
```

#### /verify/service/:service

The `/verify/service/:service` route returns 200 if all Consul checks for the specified service are in the PASSING state.  Otherwise, a non-200 status code is returned and the failing checks will be in the response.

##### Request

```bash
curl -X GET http:/localhost:8080/verify/service/test
```

##### Response

```json
{"Status":"Ok"}
```

```json
{"Status":"Failed","Checks":{"port_check":{"Node":"macbook-pro.local","CheckID":"port_check","Name":"port check","Status":"critical","Notes":"Port Check","Output":"Timed out (1s) running check","ServiceID":"test","ServiceName":"test","ServiceTags":[],"Definition":{"HTTP":"","Header":null,"Method":"","TLSSkipVerify":false,"TCP":"","Interval":0,"Timeout":0,"DeregisterCriticalServiceAfter":0},"CreateIndex":0,"ModifyIndex":0}}}
```

## License

Apache License 2.0, see [LICENSE](https://github.com/kadaan/consulate/blob/master/LICENSE).
