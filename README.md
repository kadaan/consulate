# Consulate [![Build Status](https://travis-ci.org/kadaan/consulate.svg?branch=master)](https://travis-ci.org/kadaan/consulate) [![Go Report Card](https://goreportcard.com/badge/github.com/kadaan/consulate)](https://goreportcard.com/report/github.com/kadaan/consulate) [![Coverage Status](https://coveralls.io/repos/github/kadaan/consulate/badge.svg?branch=master)](https://coveralls.io/github/kadaan/consulate?branch=master)

Consulate provides a normalized HTTP endpoint that responds with
200 / non-200 according to Consul checks specified in the requested URL.
This endpoint can then be composed with existing HTTP based monitoring tools, like
AWS ELB health checks, to enable service monitoring based on Consul checks.

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
>    "BuildUser": "user@computer.local",
>    "BuildDate": "2018-02-15T07:10:39Z",
>    "GoVersion": "go1.9"
>}
>```
>
>##### Status Codes
>* `200`: Successful call
>* `500`: Unexpected failure

---

#### /health

>The `/health` route returns 200 if Consulate is running and able to communicate with 
>Consul.  Otherwise, a non-200 status code is returned.
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
>
>##### Status Codes
>* `200`: Successful call
>* `422`: Could not parse the response from Consul
>* `500`: Unexpected failure
>* `503`: Consul unavailable

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
>
>##### Status Codes
>* `200`: Successful call
>* `500`: Unexpected failure

---

### Verify
Consulate verifies Consul checks by inspecting the status.  The possible status values, in increasing severity are:
* `passing`
* `maintenance`
* `warning`
* `critical`

By default, verification will fail checks whose status is not `passing`.

This can be changed by specifying the `status` query string parameter like `?status=warning`.  Only checks whose status is worse than the specified value will cause a failure.

The following table shows all possible status query string parameter values and, as a result, which checks will be considered failing. 

| Status Query String Parameter                | Fails When                       | Default? |
| -------------------------------------------- | -------------------------------- | -------- |
| _Not specified_ or `?status=passing`         | Check is not `passing`           | Yes      |
| `?status=maintenance`                        | Check is `warning` or `critical` | No       |
| `?status=warning`                            | Check is `critical`              | No       |
| `?status=critical`                           | Never                            | No       |

#### /verify/checks

>The `/verify/checks` route returns 200 if all Consul checks ok.  Otherwise, a non-200 status code is returned and the failing checks will be in the response.
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

>>##### Unhealthy
>>```
>>HTTP/1.1 429 Too Many Requests
>>Content-Type: application/json; charset=utf-8
>>```
>>```json
>>{
>>    "Status": "Failed",
>>    "Checks": {
>>        "check1b": {
>>            "Node": "computer.local",
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
>
>##### Status Codes
>* `200`: Successful call
>* `404`: No checks
>* `422`: Could not parse the response from Consul
>* `429`: One or more Consul checks have failed
>* `500`: Unexpected failure
>* `503`: Consul unavailable

---

#### /verify/checks/id/:checkId

>The `/verify/checks/id/:checkId` route returns 200 if the specified Consul check is ok.  Otherwise, a non-200 status code is returned and the failing check will be in the response.
>
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/verify/checks/id/check1b\?pretty
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

>> ###### Unhealthy
>>```
>>HTTP/1.1 429 Too Many Requests
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>    "Status": "Failed",
>>    "Checks": {
>>        "check1b": {
>>            "Node": "computer.local",
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
>
>##### Status Codes
>* `200`: Successful call
>* `404`: 
>   * No checks
>   * No checks matching specified _CheckID_
>* `422`: Could not parse the response from Consul
>* `429`: One or more Consul checks have failed
>* `500`: Unexpected failure
>* `503`: Consul unavailable

---

#### /verify/checks/id/:checkName

>The `/verify/checks/id/:checkName` route returns 200 if the specified Consul check is ok.  Otherwise, a non-200 status code is returned and the failing check will be in the response.
>More then one check will be verified if multiple checks share the same name.
>
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/verify/checks/id/check%201\?pretty
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

>> ###### Unhealthy
>>```
>>HTTP/1.1 429 Too Many Requests
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>    "Status": "Failed",
>>    "Checks": {
>>        "check1b": {
>>            "Node": "computer.local",
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
>
>##### Status Codes
>* `200`: Successful call
>* `404`: 
>   * No checks
>   * No checks matching specified _CheckName_
>* `422`: Could not parse the response from Consul
>* `429`: One or more Consul checks have failed
>* `500`: Unexpected failure
>* `503`: Consul unavailable

---

#### /verify/service/id/:serviceId

>The `/verify/service/id/:serviceId` route returns 200 if all Consul checks for the specified service are ok.  Otherwise, a non-200 status code is returned and the failing checks will be in the response.
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/verify/service/id/service2\?pretty
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

>> ###### Unhealthy
>>```
>>HTTP/1.1 429 Too Many Requests
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>    "Status": "Failed",
>>    "Checks": {
>>        "check1b": {
>>            "Node": "computer.local",
>>            "CheckID": "check1b",
>>            "Name": "check 1",
>>            "Status": "critical",
>>            "Notes": "Check 1",
>>            "Output": "Timed out (1s) running check",
>>            "ServiceID": "service2",
>>            "ServiceName": "service 2",
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
>
>##### Status Codes
>* `200`: Successful call
>* `404`: 
>   * No checks
>   * No checks for services matching specified _ServiceID_
>* `422`: Could not parse the response from Consul
>* `429`: One or more Consul checks have failed
>* `500`: Unexpected failure
>* `503`: Consul unavailable

---

#### /verify/service/name/:serviceName

>The `/verify/service/name/:serviceName` route returns 200 if all Consul checks for the specified service are ok.  Otherwise, a non-200 status code is returned and the failing checks will be in the response.
>More than one service will be verified if multiple services share the same name.
>
>##### Request
>
>```bash
>curl -X GET http:/localhost:8080/verify/service/name/service%202\?pretty
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

>> ###### Unhealthy
>>```
>>HTTP/1.1 429 Too Many Requests
>>Content-Type: application/json; charset=utf-8
>>...
>>```
>>```json
>>{
>>    "Status": "Failed",
>>    "Checks": {
>>        "check1b": {
>>            "Node": "computer.local",
>>            "CheckID": "check1b",
>>            "Name": "check 1",
>>            "Status": "critical",
>>            "Notes": "Check 1",
>>            "Output": "Timed out (1s) running check",
>>            "ServiceID": "service2",
>>            "ServiceName": "service 2",
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
>
>##### Status Codes
>* `200`: Successful call
>* `404`: 
>   * No checks
>   * No checks for services matching specified _ServiceName_
>* `422`: Could not parse the response from Consul
>* `429`: One or more Consul checks have failed
>* `500`: Unexpected failure
>* `503`: Consul unavailable

## License

Apache License 2.0, see [LICENSE](https://github.com/kadaan/consulate/blob/master/LICENSE).
