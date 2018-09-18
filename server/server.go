// Copyright © 2018 Joel Baranick <jbaranick@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"github.com/kadaan/consulate/caching"
	"github.com/kadaan/consulate/checks"
	"github.com/kadaan/consulate/client"
	"github.com/kadaan/consulate/config"
	"github.com/kadaan/consulate/spi"
	"github.com/kadaan/consulate/version"
	"github.com/kadaan/go-gin-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"strings"
)

const (
	consulChecksUrl        = "http://%s/v1/agent/checks"
	verifyCheckParamKey    = "check"
	verifyCheckParamTag    = ":" + verifyCheckParamKey
	verifyServiceParamKey  = "service"
	verifyServiceParamTag  = ":" + verifyServiceParamKey
	prettyQueryStringKey   = "pretty"
	verboseQueryStringKey  = "verbose"
	statusQueryStringKey   = "status"
	aboutRoute             = "/about"
	healthRoute            = "/health"
	verifyAllChecksRoute   = "/verify/checks"
	verifyCheckIdRoute     = verifyAllChecksRoute + "/id/" + verifyCheckParamTag
	verifyCheckNameRoute   = verifyAllChecksRoute + "/name/" + verifyCheckParamTag
	verifyServiceIdRoute   = "/verify/service/id/" + verifyServiceParamTag
	verifyServiceNameRoute = "/verify/service/name/" + verifyServiceParamTag
)

var (
	state = stopped
)

type serverState int

const (
	stopped serverState = iota
	started
)

type server struct {
	config     config.ServerConfig
	httpServer http.Server
	httpClient http.Client
	jsonApi    jsoniter.API
	cache      caching.Cache
}

// NewServer create a new Consulate server.
func NewServer(c *config.ServerConfig) spi.Server {
	svr := &server{config: *c}
	return svr
}

// Start begins the Server.
func (r *server) Start() (spi.RunningServer, error) {
	if state == stopped {
		state = started
		r.createJsonAPI()
		r.createCache()
		r.createServer()
		r.createClient()
		var err error
		go func() {
			log.Printf("Started Consulate server on %s\n", r.config.ListenAddress)
			if err = r.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Failed to start Consulate server: %s\n", err)
			}
		}()
		return r, err
	}
	return nil, errors.New("cannot start server because it is already running")
}

func (r *server) Stop() {
	if state == started {
		log.Println("Shutting down Consulate server...")
		ctx, cancel := context.WithTimeout(context.Background(), r.config.ShutdownTimeout)
		defer func() {
			cancel()
			state = stopped
		}()
		if err := r.httpServer.Shutdown(ctx); err != nil {
			log.Panicf("Consulate server shutdown failed:%s", err)
		}
		log.Println("Consulate server shutdown")
	}
}

func (r *server) createRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.UseRawPath = true
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	r.attachPrometheusMiddleware(router)
	r.handle(router, aboutRoute, r.about)
	r.handle(router, healthRoute, r.health)
	r.handle(router, verifyAllChecksRoute, r.verifyAllChecks)
	r.handle(router, verifyCheckIdRoute, r.verifyCheckId)
	r.handle(router, verifyCheckNameRoute, r.verifyCheckName)
	r.handle(router, verifyServiceIdRoute, r.verifyServiceId)
	r.handle(router, verifyServiceNameRoute, r.verifyServiceName)
	return router
}

func (r *server) attachPrometheusMiddleware(engine *gin.Engine) {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "consulate",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests made.",
		},
		[]string{"code", "method", "url"},
	)

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "consulate",
			Name:      "request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
			Buckets:   prometheus.ExponentialBuckets(0.5, 2, 6),
		},
		[]string{"code", "method", "url"},
	)

	requestSize := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "consulate",
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
			Buckets:   prometheus.ExponentialBuckets(128, 2, 4),
		},
		[]string{"code", "method", "url"},
	)

	responseSize := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "consulate",
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
			Buckets:   prometheus.ExponentialBuckets(512, 4, 4),
		},
		[]string{"code", "method", "url"},
	)

	prometheus.MustRegister(counter, duration, requestSize, responseSize)

	b := ginprometheus.NewBuilder()
	b.Counter(counter)
	b.Duration(duration)
	b.RequestSize(requestSize)
	b.ResponseSize(responseSize)
	b.UrlMapping(func(c *gin.Context) string {
		url := strings.TrimSuffix(c.Request.URL.Path, "/")
		for _, param := range c.Params {
			if param.Key == verifyCheckParamKey {
				url = strings.Replace(url, param.Value, verifyCheckParamTag, 1)
				break
			} else if param.Key == verifyServiceParamKey {
				url = strings.Replace(url, param.Value, verifyServiceParamTag, 1)
				break
			}
		}
		return url
	})
	b.Use(engine)
}

func (r *server) handle(router *gin.Engine, relativePath string, handler gin.HandlerFunc) {
	router.GET(relativePath, handler)
	router.GET(relativePath+"/", handler)
	router.HEAD(relativePath, handler)
	router.HEAD(relativePath+"/", handler)
}

func (r *server) createClient() {
	r.httpClient = *client.CreateClient(&r.config.ClientConfig)
}

func (r *server) createJsonAPI() {
	r.jsonApi = jsoniter.ConfigCompatibleWithStandardLibrary
}

func (r *server) createServer() {
	router := r.createRouter()

	r.httpServer = http.Server{
		Addr:         r.config.ListenAddress,
		Handler:      router,
		ReadTimeout:  r.config.ReadTimeout,
		WriteTimeout: r.config.WriteTimeout,
	}
}

func (r *server) createCache() {
	r.cache = *caching.NewCache(r.config.CacheConfig)
}

func (r *server) json(context *gin.Context, code int, obj interface{}) {
	_, ok := context.GetQuery(prettyQueryStringKey)
	if ok {
		context.IndentedJSON(code, obj)
	} else {
		context.JSON(code, obj)
	}
}

func (r *server) abortWithStatusJSON(context *gin.Context, code int, obj interface{}) {
	var message string
	switch obj.(type) {
	case checks.Result:
		b, _ := json.Marshal(obj)
		message = string(b)
	default:
		message = fmt.Sprint(obj)
	}
	context.Error(errors.New(message)).SetType(gin.ErrorTypePrivate)
	context.Abort()
	r.json(context, code, obj)
}

func (r *server) about(context *gin.Context) {
	r.json(context, r.config.SuccessStatusCode, version.NewInfo())
}

func (r *server) health(context *gin.Context) {
	r.processChecks(context, func(allChecks *map[string]*checks.Check) {
		r.json(context, r.config.SuccessStatusCode, checks.Result{Status: checks.Ok})
	})
}

type checkHandler func(allChecks *map[string]*checks.Check)

type checkMatcher struct {
	noChecksErrorMessage string
	matcher              func(c *checks.Check) bool
}

func (m *checkMatcher) match(c *checks.Check) bool {
	return m.matcher(c)
}

func (r *server) verifyAllChecks(context *gin.Context) {
	matcher := checkMatcher{
		noChecksErrorMessage: "No checks",
		matcher:              func(c *checks.Check) bool { return true },
	}
	r.verifyChecks(context, matcher)
}

func (r *server) verifyCheckId(context *gin.Context) {
	check := context.Param(verifyCheckParamKey)
	matcher := checkMatcher{
		noChecksErrorMessage: fmt.Sprintf("No checks with CheckID: %s", check),
		matcher:              func(c *checks.Check) bool { return c.IsCheckId(check) },
	}
	r.verifyChecks(context, matcher)
}

func (r *server) verifyCheckName(context *gin.Context) {
	check := context.Param(verifyCheckParamKey)
	matcher := checkMatcher{
		noChecksErrorMessage: fmt.Sprintf("No checks with CheckName: %s", check),
		matcher:              func(c *checks.Check) bool { return c.IsCheckName(check) },
	}
	r.verifyChecks(context, matcher)
}

func (r *server) verifyServiceId(context *gin.Context) {
	service := context.Param(verifyServiceParamKey)
	matcher := checkMatcher{
		noChecksErrorMessage: fmt.Sprintf("No checks for services with ServiceId: %s", service),
		matcher:              func(c *checks.Check) bool { return c.IsServiceId(service) },
	}
	r.verifyChecks(context, matcher)
}

func (r *server) verifyServiceName(context *gin.Context) {
	service := context.Param(verifyServiceParamKey)
	matcher := checkMatcher{
		noChecksErrorMessage: fmt.Sprintf("No checks for services with ServiceName: %s", service),
		matcher:              func(c *checks.Check) bool { return c.IsServiceName(service) },
	}
	r.verifyChecks(context, matcher)
}

func (r *server) verifyChecks(context *gin.Context, matcher checkMatcher) {
	r.processChecks(context, func(allChecks *map[string]*checks.Check) {
		_, isVerbose := context.GetQuery(verboseQueryStringKey)
		s := r.getStatus(context)

		var checkCount = 0
		var verifiedCheckCount = 0
		var statusCounts = map[checks.Status]int{
			checks.StatusPassing: 0,
			checks.StatusWarning: 0,
			checks.StatusFailing: 0,
		}
		var matchedChecks map[string]*checks.Check
		matchedChecks = make(map[string]*checks.Check)
		for k, v := range *allChecks {
			checkCount++
			if matcher.match(v) {
				verifiedCheckCount++
				m, e := v.MatchStatus(s)
				if e != nil {
					r.abortWithStatusJSON(context, r.config.UnprocessableStatusCode,
						checks.Result{Status: checks.Failed, Detail: e.Error()})
				}
				statusCounts[m] = statusCounts[m] + 1
				if m != checks.StatusPassing || isVerbose {
					matchedChecks[k] = v
				}
			}
		}
		if statusCounts[checks.StatusFailing] > 0 {
			r.abortWithStatusJSON(context, r.config.ErrorStatusCode,
				checks.Result{Status: checks.Failed, Counts: statusCounts, Checks: matchedChecks})
		} else if statusCounts[checks.StatusPassing] == 0 && statusCounts[checks.StatusWarning] > 0 {
			r.abortWithStatusJSON(context, r.config.WarningStatusCode,
				checks.Result{Status: checks.Failed, Counts: statusCounts, Checks: matchedChecks})
		} else if statusCounts[checks.StatusWarning] > 0 {
			r.abortWithStatusJSON(context, r.config.PartialSuccessStatusCode,
				checks.Result{Status: checks.Warning, Counts: statusCounts, Checks: matchedChecks})
		} else if checkCount == 0 || verifiedCheckCount == 0 {
			r.abortWithStatusJSON(context, r.config.NoCheckStatusCode,
				checks.Result{Status: checks.NoChecks, Detail: matcher.noChecksErrorMessage})
		} else {
			r.json(context, r.config.SuccessStatusCode, checks.Result{Status: checks.Ok, Checks: matchedChecks})
		}
	})
}

func (r *server) processChecks(context *gin.Context, handler checkHandler) {
	target := fmt.Sprintf(consulChecksUrl, r.config.ConsulAddress)
	var allChecks *map[string]*checks.Check
	if cachedChecks, ok := r.cache.Get(target); ok {
		allChecks = cachedChecks.(*map[string]*checks.Check)
	} else {
		resp, err := r.httpClient.Get(target)
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			r.abortWithStatusJSON(context, r.config.ConsulUnavailableStatusCode,
				checks.Result{Status: checks.Failed, Detail: err.Error()})
			return
		}
		err = r.jsonApi.NewDecoder(resp.Body).Decode(&allChecks)
		if err != nil {
			r.abortWithStatusJSON(context, r.config.UnprocessableStatusCode,
				checks.Result{Status: checks.Failed, Detail: err.Error()})
			return
		}
		r.cache.Set(target, allChecks)
	}
	handler(allChecks)
}

func (r *server) getStatus(context *gin.Context) checks.HealthStatus {
	status, statusSpecified := context.GetQuery(statusQueryStringKey)
	if !statusSpecified {
		return checks.HealthPassing
	}
	parsedStatus, parsed := checks.ParseHealthStatus(status)
	if !parsed {
		r.abortWithStatusJSON(context, r.config.BadRequestStatusCode,
			checks.Result{
				Status: checks.Failed,
				Detail: fmt.Sprintf("Unsupported status: %v", context.Query(statusQueryStringKey))})
	}
	return parsedStatus
}
