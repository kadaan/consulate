# Consulate [![Build Status](https://travis-ci.org/kadaan/consulate.svg?branch=master)](https://travis-ci.org/kadaan/consulate) [![Go Report Card](https://goreportcard.com/badge/github.com/kadaan/consulate)](https://goreportcard.com/report/github.com/kadaan/consulate)

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

---

#### /about
>The `/about` route returns detailed version information about Consulate.
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/about\?pretty
>```
>
>##### Response
>```
>HTTP/1.1 200 OK
>Content-Type: application/json; charset=utf-8
>...
>```
>```json
>{
>    "Version": "0.0.1",
>    "Revision": "c216c1294676cdaac0b018244be31ebb6e404b92",
>    "Branch": "master",
>    "BuildUser": "jbaranick@ensadmins-MacBook-Pro.local",
>    "BuildDate": "2018-02-15T07:10:39Z",
>    "GoVersion": "go1.9"
>}
>```

---

#### /health

>The `/health` route returns 200 if Consulate is running and able to communicate with >Consul.  Otherwise, a non-200 status code is returned.
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/health\?pretty
>```
>
>##### Responses
>>###### Healthy
>>```
>>HTTP/1.1 200 OK
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>    "Status": "Ok"
>>}
>>```

> &nbsp;

>>###### Unhealthy
>>```
>>HTTP/1.1 503 Service Unavailable
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>    "Status": "Failed",
>>    "Detail": "Get http://localhost:8500/v1/agent/checks: dial tcp [::1]:8500: getsockopt: connection refused"
>>}
>>```

---

#### /metrics

>The `/metrics` route returns Prometheus metrics.
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/metrics
>```
>
>##### Response
>
>```
># HELP consulate_request_duration_seconds The HTTP request latencies in seconds.
># TYPE consulate_request_duration_seconds histogram
>consulate_request_duration_seconds_bucket{code="200",method="GET",url="/about",le="0.5"} 1
>consulate_request_duration_seconds_bucket{code="200",method="GET",url="/about",le="1"} 1
>consulate_request_duration_seconds_bucket{code="200",method="GET",url="/about",le="2"} 1
>consulate_request_duration_seconds_bucket{code="200",method="GET",url="/about",le="3"} 1
>...
>```

---

#### /verify/checks

>The `/verify/checks` route returns 200 if all Consul checks are in the PASSING state.  >Otherwise, a non-200 status code is returned and the failing checks will be in the >response.
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/verify/checks\?pretty
>```
>
>##### Responses
>> ###### Healthy
>>```
>>HTTP/1.1 200 OK
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>  "Status": "Ok"
>>}
>>```

> &nbsp;

>>##### Unhealthy
>>```
>>HTTP/1.1 500 Internal Server Error
>>Content-Type: application/json; charset=utf-8
>>```
>>```json
>>{
>>    "Status": "Failed",
>>    "Checks": {
>>        "check1b": {
>>            "Node": "ensadmins-MacBook-Pro.local",
>>            "CheckID": "check1b",
>>            "Name": "check 1",
>>            "Status": "critical",
>>            "Notes": "Check 1",
>>            "Output": "Timed out (1s) running check",
>>            "ServiceID": "service2",
>>            "ServiceName": "service2",
>>            "ServiceTags": [],
>>            "Definition": {
>>                "HTTP": "",
>>                "Header": null,
>>                "Method": "",
>>                "TLSSkipVerify": false,
>>                "TCP": "",
>>                "Interval": 0,
>>                "Timeout": 0,
>>                "DeregisterCriticalServiceAfter": 0
>>            },
>>            "CreateIndex": 0,
>>            "ModifyIndex": 0
>>        }
>>    }
>>}
>>```

---

#### /verify/checks/:check

>The `/verify/checks/:check` route returns 200 if the specified Consul check is in the >PASSING state.  Otherwise, a non-200 status code is returned and the failing check will >be in the response.
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/verify/checks/check%201\?pretty
>```
>
>##### Responses
>> ###### Healthy
>>```
>>HTTP/1.1 200 OK
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>    "Status": "Ok"
>>}
>>```

> &nbsp;

>> ###### Unhealthy
>>```
>>HTTP/1.1 500 Internal Server Error
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>    "Status": "Failed",
>>    "Checks": {
>>        "check1b": {
>>            "Node": "ensadmins-MacBook-Pro.local",
>>            "CheckID": "check1b",
>>            "Name": "check 1",
>>            "Status": "critical",
>>            "Notes": "Check 1",
>>            "Output": "Timed out (1s) running check",
>>            "ServiceID": "service2",
>>            "ServiceName": "service2",
>>            "ServiceTags": [],
>>            "Definition": {
>>                "HTTP": "",
>>                "Header": null,
>>                "Method": "",
>>                "TLSSkipVerify": false,
>>                "TCP": "",
>>                "Interval": 0,
>>                "Timeout": 0,
>>                "DeregisterCriticalServiceAfter": 0
>>            },
>>            "CreateIndex": 0,
>>            "ModifyIndex": 0
>>        }
>>    }
>>}
>>```

---

#### /verify/service/:service

>The `/verify/service/:service` route returns 200 if all Consul checks for the specified >service are in the PASSING state.  Otherwise, a non-200 status code is returned and the >failing checks will be in the response.
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/verify/service/test
>```
>
>##### Responses
>> ###### Healthy
>>```
>>HTTP/1.1 200 OK
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
{
    "Status": "Ok"
}
>>```

> &nbsp;

>> ###### Unhealthy
>>```
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=utf-8
...
>>```
>>```json
{
    "Status": "Failed",
    "Checks": {
        "check1b": {
            "Node": "ensadmins-MacBook-Pro.local",
            "CheckID": "check1b",
            "Name": "check 1",
            "Status": "critical",
            "Notes": "Check 1",
            "Output": "Timed out (1s) running check",
            "ServiceID": "service2",
            "ServiceName": "service2",
            "ServiceTags": [],
            "Definition": {
                "HTTP": "",
                "Header": null,
                "Method": "",
                "TLSSkipVerify": false,
                "TCP": "",
                "Interval": 0,
                "Timeout": 0,
                "DeregisterCriticalServiceAfter": 0
            },
            "CreateIndex": 0,
            "ModifyIndex": 0
        }
    }
}
>```

## License

Apache License 2.0, see [LICENSE](https://github.com/kadaan/consulate/blob/master/LICENSE).
