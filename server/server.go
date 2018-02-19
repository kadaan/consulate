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
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"github.com/kadaan/consulate/checks"
	"github.com/kadaan/consulate/client"
	"github.com/kadaan/consulate/config"
	"github.com/kadaan/consulate/spi"
	"github.com/kadaan/consulate/version"
	"github.com/kadaan/go-gin-prometheus"
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
	requestDurationBuckets = []float64{0.5, 1, 2, 3, 5, 10}
	requestSizeBuckets     = []float64{128, 256, 512, 1024}
	responseSizeBuckets    = []float64{512, 2048, 8196, 32784}
)

type server struct {
	config     config.ServerConfig
	httpServer http.Server
	httpClient http.Client
	jsonApi    jsoniter.API
}

func NewServer(c *config.ServerConfig) spi.Server {
	svr := &server{config: *c}
	return svr
}

func (r *server) Start() (spi.RunningServer, error) {
	r.createJsonAPI()
	r.createServer()
	r.createClient()
	var err error = nil
	go func() {
		log.Printf("Started Consulate server on %s\n", r.config.ListenAddress)
		if err = r.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed to start Consulate server: %s\n", err)
		}
	}()
	return r, err
}

func (r *server) Stop() {
	log.Println("Shutting down Consulate server...")
	ctx, cancel := context.WithTimeout(context.Background(), r.config.ShutdownTimeout)
	defer cancel()
	if err := r.httpServer.Shutdown(ctx); err != nil {
		log.Panicf("Consulate server shutdown failed:%s", err)
	}
	log.Println("Consulate server shutdown")
}

func (r *server) createRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.UseRawPath = true
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	r.attachPrometheusMiddleware(router)
	router.GET(aboutRoute, r.about)
	router.GET(healthRoute, r.health)
	router.GET(verifyAllChecksRoute, r.verifyAllChecks)
	router.GET(verifyCheckIdRoute, r.verifyCheckId)
	router.GET(verifyCheckNameRoute, r.verifyCheckName)
	router.GET(verifyServiceIdRoute, r.verifyServiceId)
	router.GET(verifyServiceNameRoute, r.verifyServiceName)
	return router
}

func (r *server) attachPrometheusMiddleware(engine *gin.Engine) {
	prometheusMiddleware := ginprometheus.NewPrometheus("consulate", requestDurationBuckets, requestSizeBuckets, responseSizeBuckets)
	prometheusMiddleware.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
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
	}
	prometheusMiddleware.Use(engine)
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

func (r *server) json(context *gin.Context, code int, obj interface{}) {
	_, ok := context.GetQuery(prettyQueryStringKey)
	if ok {
		context.IndentedJSON(code, obj)
	} else {
		context.JSON(code, obj)
	}
}

func (r *server) abortWithStatusJSON(context *gin.Context, code int, obj interface{}) {
	context.Abort()
	r.json(context, code, obj)
}

func (r *server) about(context *gin.Context) {
	r.json(context, spi.StatusOK, version.NewInfo())
}

func (r *server) health(context *gin.Context) {
	r.processChecks(context, func(allChecks *map[string]*checks.Check) {
		r.json(context, spi.StatusOK, checks.Result{Status: checks.Ok})
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
		var failedCount = false
		var matchedChecks map[string]*checks.Check
		matchedChecks = make(map[string]*checks.Check)
		for k, v := range *allChecks {
			checkCount++
			if matcher.match(v) {
				verifiedCheckCount++
				b, e := v.MatchesStatus(s)
				if e != nil {
					r.abortWithStatusJSON(context, spi.StatusUnprocessableResponseError, checks.Result{Status: checks.Failed, Detail: e.Error()})
				}
				if b {
					matchedChecks[k] = v
					failedCount = true
				} else if isVerbose {
					matchedChecks[k] = v
				}
			}
		}
		if failedCount {
			r.abortWithStatusJSON(context, spi.StatusCheckError, checks.Result{Status: checks.Failed, Checks: matchedChecks})
		} else if checkCount == 0 || verifiedCheckCount == 0 {
			r.abortWithStatusJSON(context, spi.StatusNoChecksError, checks.Result{Status: checks.NoChecks, Detail: matcher.noChecksErrorMessage})
		} else {
			r.json(context, spi.StatusOK, checks.Result{Status: checks.Ok, Checks: matchedChecks})
		}
	})
}

func (r *server) processChecks(context *gin.Context, handler checkHandler) {
	target := fmt.Sprintf(consulChecksUrl, r.config.ConsulAddress)
	resp, err := r.httpClient.Get(target)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		r.abortWithStatusJSON(context, spi.StatusConsulUnavailableError, checks.Result{Status: checks.Failed, Detail: err.Error()})
	} else {
		var allChecks map[string]*checks.Check
		err = r.jsonApi.NewDecoder(resp.Body).Decode(&allChecks)
		if err != nil {
			r.abortWithStatusJSON(context, spi.StatusUnprocessableResponseError, checks.Result{Status: checks.Failed, Detail: err.Error()})
		} else {
			handler(&allChecks)
		}
	}
}

func (r *server) getStatus(context *gin.Context) checks.HealthStatus {
	status, statusSpecified := context.GetQuery(statusQueryStringKey)
	if !statusSpecified {
		return checks.HealthPassing
	}
	parsedStatus, parsed := checks.ParseHealthStatus(status)
	if !parsed {
		r.abortWithStatusJSON(context, spi.StatusBadRequestError, checks.Result{Status: checks.Failed, Detail: fmt.Sprintf("Unsupported status: %v", context.Query(statusQueryStringKey))})
	}
	return parsedStatus
}
