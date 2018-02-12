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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zsais/go-gin-prometheus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"strings"
)

const (
	VerboseKey                     = "verbose"
	ListenAddressKey               = "listen-address"
	ConsulAddressKey               = "consul-address"
	ReadTimeoutKey                 = "read-timeout"
	WriteTimeoutKey                = "write-timeout"
	QueryTimeoutKey                = "query-timeout"
	QueryMaxIdleConnectionCountKey = "query-max-idle-connection-count"
	QueryIdleConnectionTimeoutKey  = "query-idle-connection-timeout"
	ShutdownTimeoutKey             = "shutdown-timeout"
	ConsulChecksUrl                = "http://%s/v1/agent/checks"
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


	httpClient = createClient()
	jsonApi    = createJsonAPI()
    serverCmd  = &cobra.Command{
		Use:   "server",
		Short: "Runs the Consulate server",
		Long: `Starts the Consulate server and runs until an interrupt is received.`,
		Run: func(cmd *cobra.Command, args []string) {
			serve()
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().BoolVarP(&verbose, VerboseKey, "v", false, "enable verbose logging")
	viper.BindPFlag(VerboseKey, serverCmd.Flags().Lookup(VerboseKey))
	serverCmd.Flags().StringVarP(&listenAddress, ListenAddressKey, "l", ":8080", "the listen address")
	viper.BindPFlag(ListenAddressKey, serverCmd.Flags().Lookup(ListenAddressKey))
	serverCmd.Flags().StringVarP(&consulAddress, ConsulAddressKey, "c", "localhost:8500", "the Consul HTTP API address to query against")
	viper.BindPFlag(ConsulAddressKey, serverCmd.Flags().Lookup(ConsulAddressKey))
	serverCmd.Flags().DurationVar(&readTimeout, ReadTimeoutKey, 10 * time.Second, "the maximum duration for reading the entire request")
	viper.BindPFlag(ReadTimeoutKey, serverCmd.Flags().Lookup(ReadTimeoutKey))
	serverCmd.Flags().DurationVar(&writeTimeout, WriteTimeoutKey, 10 * time.Second, "the maximum duration before timing out writes of the response")
	viper.BindPFlag(WriteTimeoutKey, serverCmd.Flags().Lookup(WriteTimeoutKey))
	serverCmd.Flags().DurationVar(&queryTimeout, QueryTimeoutKey, 5 * time.Second, "the maximum duration before timing out the Consul HTTP API query")
	viper.BindPFlag(QueryTimeoutKey, serverCmd.Flags().Lookup(QueryTimeoutKey))
	serverCmd.Flags().IntVar(&queryMaxIdleConnectionCount, QueryMaxIdleConnectionCountKey, 100, "the maximum number of idle (keep-alive) Consul HTTP API query connections")
	viper.BindPFlag(QueryMaxIdleConnectionCountKey, serverCmd.Flags().Lookup(QueryMaxIdleConnectionCountKey))
	serverCmd.Flags().DurationVar(&queryIdleConnectionTimeout, QueryIdleConnectionTimeoutKey, 90 * time.Second, "is the maximum amount of time an idle (keep-alive) Consul HTTP API query connection will remain idle before closing itself")
	viper.BindPFlag(QueryIdleConnectionTimeoutKey, serverCmd.Flags().Lookup(QueryIdleConnectionTimeoutKey))
	serverCmd.Flags().DurationVar(&shutdownTimeout, ShutdownTimeoutKey, 15 * time.Second, "the maximum duration before timing out the shutdown of the server")
	viper.BindPFlag(ShutdownTimeoutKey, serverCmd.Flags().Lookup(ShutdownTimeoutKey))
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
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	attachPrometheusMiddleware(router)
	router.GET("/about", about)
	router.GET("/verify/checks", verifyAllChecks)
	router.GET("/verify/service/:service", verifyService)
	return router
}

func attachPrometheusMiddleware(engine *gin.Engine) {
	logrus.SetLevel(logrus.WarnLevel)
	prometheusMiddleware := ginprometheus.NewPrometheus("gin")
	prometheusMiddleware.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.String()
		for _, param := range c.Params {
			if param.Key == "service" {
				url = strings.Replace(url, param.Value, ":service", 1)
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
		Timeout: queryTimeout,
	}
	return client
}

func createJsonAPI() jsoniter.API {
	return jsoniter.ConfigCompatibleWithStandardLibrary
}

func startServer(router *gin.Engine) *http.Server {
	srv := &http.Server {
		Addr: listenAddress,
		Handler: router,
		ReadTimeout: readTimeout,
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

func about(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, version.NewInfo())
}

type checkMatcher func(check *checks.Check) bool

func verifyAllChecks(context *gin.Context) {
	matcher := func(check *checks.Check) bool {return true}
	verifyChecks(context, matcher)
}

func verifyService(context *gin.Context) {
	service := context.Param("service")
	matcher := func(check *checks.Check) bool {return check.IsService(service)}
	verifyChecks(context, matcher)
}

func verifyChecks(context *gin.Context, match checkMatcher) {
	target := fmt.Sprintf(ConsulChecksUrl, consulAddress)
	resp, err := httpClient.Get(target)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		context.AbortWithStatusJSON(http.StatusServiceUnavailable, checks.Result{Status: checks.Failed, Detail: err.Error()})
	} else {
		var allChecks map[string]*checks.Check
		err = jsonApi.NewDecoder(resp.Body).Decode(&allChecks)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusUnprocessableEntity, checks.Result{Status: checks.Failed, Detail: err.Error()})
		} else {
			var checkCount = 0
			var failedChecks map[string]*checks.Check
			failedChecks = make(map[string]*checks.Check)
			for k, v := range allChecks {
				if match(v) {
					checkCount++
					if !v.IsHealthy() {
						failedChecks[k] = v
					}
				}
			}
			if len(failedChecks) > 0 {
				context.AbortWithStatusJSON(http.StatusInternalServerError, checks.Result{Status: checks.Failed, Checks: failedChecks})
			} else if checkCount == 0 {
				context.AbortWithStatusJSON(http.StatusNotFound, checks.Result{Status: checks.NoChecks})
			} else {
				context.JSON(http.StatusOK, checks.Result{Status: checks.OK})
			}
		}
		resp.Body.Close()
	}
}