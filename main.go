package main

import (
	"github.com/urfave/cli"
	"github.com/gin-gonic/gin"
	"os"
	"log"
	"fmt"
	"os/signal"
	"net/http"
	"context"
	"time"
	"github.com/json-iterator/go"
	"strings"
)

const (
	HealthPassing  = "passing"
)

type AgentCheck struct {
	Node        string
	CheckID     string
	Name        string
	Status      string
	Notes       string
	Output      string
	ServiceID   string
	ServiceName string
	ServiceTags []string
	Definition  HealthCheckDefinition
	CreateIndex uint64
	ModifyIndex uint64
}

type HealthCheckDefinition struct {
	HTTP                           string
	Header                         map[string][]string
	Method                         string
	TLSSkipVerify                  bool
	TCP                            string
	Interval                       ReadableDuration
	Timeout                        ReadableDuration
	DeregisterCriticalServiceAfter ReadableDuration
}

type ReadableDuration time.Duration

func main() {
	app := cli.NewApp()
	app.Name = "consulate"
	app.HelpName = "consulate"
	app.Version = "0.0.1"
	app.Usage = "consul checks monitoring endpoint"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		{
			Name: "server",
			Aliases: []string{"s"},
			Flags: []cli.Flag{
				cli.StringFlag{Name: "address, a", Value: "0.0.0.0"},
				cli.UintFlag{Name: "port, p", Value: 80},
				cli.UintFlag{Name: "consulPort, c", Value: 8500},
			},
			Action: ServeAction,
		},
	}
	app.Run(os.Args)
}

func ServeAction(c *cli.Context) error {
	serverAddress := c.String("address")
	serverPort := c.Uint("port")
	consulPort := c.Uint("consulPort")
	consulUrl := fmt.Sprintf("http://localhost:%d/v1/agent/checks", consulPort)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.GET("/checks/:service", CheckService(consulUrl))

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", serverAddress, serverPort),
		Handler: router,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		log.Printf("Starting Consulate server (%s)\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start Consulate server\n    %s", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down Consulate server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed\n    %s", err)
	}
	log.Println("Server exiting")
	return nil
}

func CheckService(consulUrl string) func(c *gin.Context) {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return func(c *gin.Context) {
		service := c.Param("service")
		resp, err := httpClient.Get(consulUrl)
		defer resp.Body.Close()
		if err != nil {
			c.AbortWithError(http.StatusServiceUnavailable, err)
		} else {
			var checks map[string]*AgentCheck
			err = json.NewDecoder(resp.Body).Decode(&checks)
			if err != nil {
				c.AbortWithError(http.StatusUnprocessableEntity, err)
			} else {
				var checkCount = 0
				var failedChecks map[string]*AgentCheck
				failedChecks = make(map[string]*AgentCheck)
				for k, v := range checks {
					if service == v.ServiceName || service == v.ServiceID {
						checkCount++
						if strings.ToLower(v.Status) != HealthPassing {
							failedChecks[k] = v
						}
					}
				}
				if len(failedChecks) > 0 {
					c.AbortWithStatusJSON(http.StatusInternalServerError, failedChecks)
				} else if checkCount == 0 {
					c.AbortWithStatusJSON(http.StatusNotFound, "NO CHECKS")
				} else {
					c.String(http.StatusOK, "OK")
				}
			}
			resp.Body.Close()
		}
	}
}