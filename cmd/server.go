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
	readTimeoutKey                 = "read-timeout"
	writeTimeoutKey                = "write-timeout"
	queryTimeoutKey                = "query-timeout"
	queryMaxIdleConnectionCountKey = "query-max-idle-connection-count"
	queryIdleConnectionTimeoutKey  = "query-idle-connection-timeout"
	shutdownTimeoutKey             = "shutdown-timeout"
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
}
