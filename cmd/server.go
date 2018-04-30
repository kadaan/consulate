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
	"github.com/kadaan/consulate/config"
	"github.com/kadaan/consulate/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
)

const (
	listenAddressKey               = "listen-address"
	consulAddressKey               = "consul-address"
	consulCacheDurationKey         = "consul-cache-duration"
	readTimeoutKey                 = "read-timeout"
	writeTimeoutKey                = "write-timeout"
	queryTimeoutKey                = "query-timeout"
	queryMaxIdleConnectionCountKey = "query-max-idle-connection-count"
	queryIdleConnectionTimeoutKey  = "query-idle-connection-timeout"
	shutdownTimeoutKey             = "shutdown-timeout"
	okStatusCodeKey                = "success-status-code"
	partialSuccessStatusCodeKey    = "partial-success-status-code"
	warningStatusCodeKey           = "warning-status-code"
	errorStatusCodeKey             = "error-status-code"
	badRequestStatusCodeKey        = "bad-request-status-code"
	noChecksStatusCodeKey          = "no-checks-status-code"
	unprocessableStatusCodeKey     = "unprocessable-status-code"
	consulUnavailableStatusCodeKey = "consul-navailable-status-code"
)

var (
	serverConfig config.ServerConfig

	quit      = make(chan os.Signal)
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Runs the Consulate server",
		Long:  `Starts the Consulate server and runs until an interrupt is received.`,
		Run: func(cmd *cobra.Command, args []string) {
			server, err := server.NewServer(&serverConfig).Start()
			if server != nil {
				defer server.Stop()
			}
			if err != nil {
				log.Println("Shutting down Consulate server...")
			} else {
				log.Println("Press Ctrl-C to shutdown server")
				signal.Notify(quit, os.Interrupt)
				<-quit
				server.Stop()
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&serverConfig.ListenAddress, listenAddressKey, "l", config.DefaultListenAddress, "the listen address")
	viper.BindPFlag(listenAddressKey, serverCmd.Flags().Lookup(listenAddressKey))
	serverCmd.Flags().StringVarP(&serverConfig.ConsulAddress, consulAddressKey, "c", config.DefaultConsulAddress, "the Consul HTTP API address to query against")
	viper.BindPFlag(consulAddressKey, serverCmd.Flags().Lookup(consulAddressKey))
	serverCmd.Flags().DurationVar(&serverConfig.CacheConfig.ConsulCacheDuration, consulCacheDurationKey, config.DefaultConsulCacheDuration, "the duration that Consul results will be cached")
	viper.BindPFlag(consulAddressKey, serverCmd.Flags().Lookup(consulCacheDurationKey))
	serverCmd.Flags().DurationVar(&serverConfig.ReadTimeout, readTimeoutKey, config.DefaultReadTimeout, "the maximum duration for reading the entire request")
	viper.BindPFlag(readTimeoutKey, serverCmd.Flags().Lookup(readTimeoutKey))
	serverCmd.Flags().DurationVar(&serverConfig.WriteTimeout, writeTimeoutKey, config.DefaultWriteTimeout, "the maximum duration before timing out writes of the response")
	viper.BindPFlag(writeTimeoutKey, serverCmd.Flags().Lookup(writeTimeoutKey))
	serverCmd.Flags().DurationVar(&serverConfig.ClientConfig.QueryTimeout, queryTimeoutKey, config.DefaultQueryTimeout, "the maximum duration before timing out the Consul HTTP API query")
	viper.BindPFlag(queryTimeoutKey, serverCmd.Flags().Lookup(queryTimeoutKey))
	serverCmd.Flags().IntVar(&serverConfig.ClientConfig.QueryMaxIdleConnectionCount, queryMaxIdleConnectionCountKey, config.DefaultQueryMaxIdleConnectionCount, "the maximum number of idle (keep-alive) Consul HTTP API query connections")
	viper.BindPFlag(queryMaxIdleConnectionCountKey, serverCmd.Flags().Lookup(queryMaxIdleConnectionCountKey))
	serverCmd.Flags().DurationVar(&serverConfig.ClientConfig.QueryIdleConnectionTimeout, queryIdleConnectionTimeoutKey, config.DefaultQueryIdleConnectionTimeout, "is the maximum amount of time an idle (keep-alive) Consul HTTP API query connection will remain idle before closing itself")
	viper.BindPFlag(queryIdleConnectionTimeoutKey, serverCmd.Flags().Lookup(queryIdleConnectionTimeoutKey))
	serverCmd.Flags().DurationVar(&serverConfig.ShutdownTimeout, shutdownTimeoutKey, config.DefaultShutdownTimeout, "the maximum duration before timing out the shutdown of the server")
	viper.BindPFlag(shutdownTimeoutKey, serverCmd.Flags().Lookup(shutdownTimeoutKey))
	serverCmd.Flags().IntVar(&serverConfig.SuccessStatusCode, okStatusCodeKey, config.DefaultSuccessStatusCode, "the status code returned when there are 1+ passing health checks, 0 warning health checks, and 0 failing health checks")
	viper.BindPFlag(okStatusCodeKey, serverCmd.Flags().Lookup(okStatusCodeKey))
	serverCmd.Flags().IntVar(&serverConfig.PartialSuccessStatusCode, partialSuccessStatusCodeKey, config.DefaultPartialSuccessStatusCode, "the status code returned when there are 1+ passing health checks and 1+ warning health checks")
	viper.BindPFlag(partialSuccessStatusCodeKey, serverCmd.Flags().Lookup(partialSuccessStatusCodeKey))
	serverCmd.Flags().IntVar(&serverConfig.WarningStatusCode, warningStatusCodeKey, config.DefaultWarningStatusCode, "the status code returned when there are 0 passing health checks and 1+ warning health checks")
	viper.BindPFlag(warningStatusCodeKey, serverCmd.Flags().Lookup(warningStatusCodeKey))
	serverCmd.Flags().IntVar(&serverConfig.ErrorStatusCode, errorStatusCodeKey, config.DefaultErrorStatusCode, "the status code returned when there are 1+ failing health checks")
	viper.BindPFlag(errorStatusCodeKey, serverCmd.Flags().Lookup(errorStatusCodeKey))
	serverCmd.Flags().IntVar(&serverConfig.BadRequestStatusCode, badRequestStatusCodeKey, config.DefaultBadRequestStatusCode, "the status code returned when a request to Consulate could not be understood")
	viper.BindPFlag(badRequestStatusCodeKey, serverCmd.Flags().Lookup(badRequestStatusCodeKey))
	serverCmd.Flags().IntVar(&serverConfig.NoCheckStatusCode, noChecksStatusCodeKey, config.DefaultNoCheckStatusCode, "the status code returned when no Consul checks exist")
	viper.BindPFlag(noChecksStatusCodeKey, serverCmd.Flags().Lookup(noChecksStatusCodeKey))
	serverCmd.Flags().IntVar(&serverConfig.UnprocessableStatusCode, unprocessableStatusCodeKey, config.DefaultUnprocessableStatusCode, "the status code returned when Consulate could not parse the response from Consul")
	viper.BindPFlag(unprocessableStatusCodeKey, serverCmd.Flags().Lookup(unprocessableStatusCodeKey))
	serverCmd.Flags().IntVar(&serverConfig.ConsulUnavailableStatusCode, consulUnavailableStatusCodeKey, config.DefaultConsulUnavailableStatusCode, "the status code returned when Consul did not respond promptly")
	viper.BindPFlag(consulUnavailableStatusCodeKey, serverCmd.Flags().Lookup(consulUnavailableStatusCodeKey))
}
