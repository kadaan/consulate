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

package cmd

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/json-iterator/go"
	"github.com/kadaan/consulate/checks"
	"github.com/kadaan/consulate/version"
	"github.com/kadaan/go-gin-prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	verboseKey                     = "verbose"
	listenAddressKey               = "listen-address"
	consulAddressKey               = "consul-address"
	readTimeoutKey                 = "read-timeout"
	writeTimeoutKey                = "write-timeout"
	queryTimeoutKey                = "query-timeout"
	queryMaxIdleConnectionCountKey = "query-max-idle-connection-count"
	queryIdleConnectionTimeoutKey  = "query-idle-connection-timeout"
	shutdownTimeoutKey             = "shutdown-timeout"
	consulChecksUrl                = "http://%s/v1/agent/checks"
	verifyCheckParamKey            = "check"
	verifyCheckParamTag            = ":" + verifyCheckParamKey
	verifyServiceParamKey          = "service"
	verifyServiceParamTag          = ":" + verifyServiceParamKey
	prettyQueryStringKey           = "pretty"
	verboseQueryStringKey          = "verbose"
	aboutRoute                     = "/about"
	healthRoute                    = "/health"
	verifyAllChecksRoute           = "/verify/checks"
	verifySpecificCheckRoute       = verifyAllChecksRoute + "/" + verifyCheckParamTag
	verifySpecificServiceRoute     = "/verify/service/" + verifyServiceParamTag
)

var (
	verbose                     bool
	listenAddress               string
	consulAddress               string
	readTimeout                 time.Duration
	writeTimeout                time.Duration
	queryTimeout                time.Duration
	queryMaxIdleConnectionCount int
	queryIdleConnectionTimeout  time.Duration
	shutdownTimeout             time.Duration

	requestDurationBuckets = []float64{0.5, 1, 2, 3, 5, 10}
	requestSizeBuckets     = []float64{128, 256, 512, 1024}
	responseSizeBuckets    = []float64{512, 2048, 8196, 32784}

	httpClient = createClient()
	jsonApi    = createJsonAPI()
	serverCmd  = &cobra.Command{
		Use:   "server",
		Short: "Runs the Consulate server",
		Long:  `Starts the Consulate server and runs until an interrupt is received.`,
		Run: func(cmd *cobra.Command, args []string) {
			serve()
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().BoolVarP(&verbose, verboseKey, "v", false, "enable verbose logging")
	viper.BindPFlag(verboseKey, serverCmd.Flags().Lookup(verboseKey))
	serverCmd.Flags().StringVarP(&listenAddress, listenAddressKey, "l", ":8080", "the listen address")
	viper.BindPFlag(listenAddressKey, serverCmd.Flags().Lookup(listenAddressKey))
	serverCmd.Flags().StringVarP(&consulAddress, consulAddressKey, "c", "localhost:8500", "the Consul HTTP API address to query against")
	viper.BindPFlag(consulAddressKey, serverCmd.Flags().Lookup(consulAddressKey))
	serverCmd.Flags().DurationVar(&readTimeout, readTimeoutKey, 10*time.Second, "the maximum duration for reading the entire request")
	viper.BindPFlag(readTimeoutKey, serverCmd.Flags().Lookup(readTimeoutKey))
	serverCmd.Flags().DurationVar(&writeTimeout, writeTimeoutKey, 10*time.Second, "the maximum duration before timing out writes of the response")
	viper.BindPFlag(writeTimeoutKey, serverCmd.Flags().Lookup(writeTimeoutKey))
	serverCmd.Flags().DurationVar(&queryTimeout, queryTimeoutKey, 5*time.Second, "the maximum duration before timing out the Consul HTTP API query")
	viper.BindPFlag(queryTimeoutKey, serverCmd.Flags().Lookup(queryTimeoutKey))
	serverCmd.Flags().IntVar(&queryMaxIdleConnectionCount, queryMaxIdleConnectionCountKey, 100, "the maximum number of idle (keep-alive) Consul HTTP API query connections")
	viper.BindPFlag(queryMaxIdleConnectionCountKey, serverCmd.Flags().Lookup(queryMaxIdleConnectionCountKey))
	serverCmd.Flags().DurationVar(&queryIdleConnectionTimeout, queryIdleConnectionTimeoutKey, 90*time.Second, "is the maximum amount of time an idle (keep-alive) Consul HTTP API query connection will remain idle before closing itself")
	viper.BindPFlag(queryIdleConnectionTimeoutKey, serverCmd.Flags().Lookup(queryIdleConnectionTimeoutKey))
	serverCmd.Flags().DurationVar(&shutdownTimeout, shutdownTimeoutKey, 15*time.Second, "the maximum duration before timing out the shutdown of the server")
	viper.BindPFlag(shutdownTimeoutKey, serverCmd.Flags().Lookup(shutdownTimeoutKey))
}

func serve() {
	router := createRouter()

	log.Println("Starting Consulate server...")
	server := startServer(router)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down Consulate server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Consulate server shutdown failed\n    %s", err)
	}
	log.Println("Consulate server shutdown")
}

func createRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.UseRawPath = true
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	attachPrometheusMiddleware(router)
	router.GET(aboutRoute, about)
	router.GET(healthRoute, health)
	router.GET(verifyAllChecksRoute, verifyAllChecks)
	router.GET(verifySpecificCheckRoute, verifyCheck)
	router.GET(verifySpecificServiceRoute, verifyService)
	return router
}

func attachPrometheusMiddleware(engine *gin.Engine) {
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

func createClient() *http.Client {
	transport := cleanhttp.DefaultPooledTransport()
	transport.MaxIdleConns = queryMaxIdleConnectionCount
	transport.IdleConnTimeout = queryIdleConnectionTimeout

	client := &http.Client{
		Transport: transport,
		Timeout:   queryTimeout,
	}
	return client
}

func createJsonAPI() jsoniter.API {
	return jsoniter.ConfigCompatibleWithStandardLibrary
}

func startServer(router *gin.Engine) *http.Server {
	srv := &http.Server{
		Addr:         listenAddress,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
	go func() {
		log.Printf("Started Consulate server on %s\n", listenAddress)
		log.Println("Press Ctrl-C to shutdown server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start Consulate server\n    %s", err)
		}
	}()
	return srv
}

func json(context *gin.Context, code int, obj interface{}) {
	_, ok := context.GetQuery(prettyQueryStringKey)
	if ok {
		context.IndentedJSON(code, obj)
	} else {
		context.JSON(code, obj)
	}
}

func abortWithStatusJSON(context *gin.Context, code int, obj interface{}) {
	context.Abort()
	json(context, code, obj)
}

func about(context *gin.Context) {
	json(context, http.StatusOK, version.NewInfo())
}

func health(context *gin.Context) {
	processChecks(context, func(allChecks *map[string]*checks.Check) {
		json(context, http.StatusOK, checks.Result{Status: checks.Ok})
	})
}

type checkHandler func(allChecks *map[string]*checks.Check)

type checkMatcher func(check *checks.Check) bool

func verifyAllChecks(context *gin.Context) {
	matcher := func(check *checks.Check) bool { return true }
	verifyChecks(context, matcher)
}

func verifyCheck(context *gin.Context) {
	checkName := context.Param(verifyCheckParamKey)
	matcher := func(check *checks.Check) bool { return check.IsCheck(checkName) }
	verifyChecks(context, matcher)
}

func verifyService(context *gin.Context) {
	service := context.Param(verifyServiceParamKey)
	matcher := func(check *checks.Check) bool { return check.IsService(service) }
	verifyChecks(context, matcher)
}

func verifyChecks(context *gin.Context, match checkMatcher) {
	processChecks(context, func(allChecks *map[string]*checks.Check) {
		_, isVerbose := context.GetQuery(verboseQueryStringKey)
		var checkCount = 0
		var failedCount = false
		var matchedChecks map[string]*checks.Check
		matchedChecks = make(map[string]*checks.Check)
		for k, v := range *allChecks {
			if match(v) {
				checkCount++
				if !v.IsHealthy() {
					matchedChecks[k] = v
					failedCount = true
				} else if isVerbose {
					matchedChecks[k] = v
				}
			}
		}
		if failedCount {
			abortWithStatusJSON(context, http.StatusInternalServerError, checks.Result{Status: checks.Failed, Checks: matchedChecks})
		} else if checkCount == 0 {
			abortWithStatusJSON(context, http.StatusNotFound, checks.Result{Status: checks.NoChecks})
		} else {
			json(context, http.StatusOK, checks.Result{Status: checks.Ok, Checks: matchedChecks})
		}
	})
}

func processChecks(context *gin.Context, handler checkHandler) {
	target := fmt.Sprintf(consulChecksUrl, consulAddress)
	resp, err := httpClient.Get(target)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		abortWithStatusJSON(context, http.StatusServiceUnavailable, checks.Result{Status: checks.Failed, Detail: err.Error()})
	} else {
		var allChecks map[string]*checks.Check
		err = jsonApi.NewDecoder(resp.Body).Decode(&allChecks)
		if err != nil {
			abortWithStatusJSON(context, http.StatusUnprocessableEntity, checks.Result{Status: checks.Failed, Detail: err.Error()})
		} else {
			handler(&allChecks)
		}
	}
}
