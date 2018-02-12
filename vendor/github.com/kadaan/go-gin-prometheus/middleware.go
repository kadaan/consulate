package ginprometheus

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var defaultMetricPath = "/metrics"

/*
RequestCounterURLLabelMappingFn is a function which can be supplied to the middleware to control
the cardinality of the request counter's "url" label, which might be required in some contexts.
For instance, if for a "/customer/:name" route you don't want to generate a time series for every
possible customer name, you could use this function:

func(c *gin.Context) string {
	url := c.Request.URL.String()
	for _, p := range c.Params {
		if p.Key == "name" {
			url = strings.Replace(url, p.Value, ":name", 1)
			break
		}
	}
	return url
}

which would map "/customer/alice" and "/customer/bob" to their template "/customer/:name".
*/
type RequestCounterURLLabelMappingFn func(c *gin.Context) string

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqCnt               *prometheus.CounterVec
	reqDur, reqSz, resSz *prometheus.HistogramVec
	router               *gin.Engine
	listenAddress        string

	Ppg PrometheusPushGateway

	MetricsPath string

	ReqCntURLLabelMappingFn RequestCounterURLLabelMappingFn
}

// PrometheusPushGateway contains the configuration for pushing to a Prometheus pushgateway (optional)
type PrometheusPushGateway struct {

	// Push interval in seconds
	PushIntervalSeconds time.Duration

	// Push Gateway URL in format http://domain:port
	// where JOBNAME can be any string of your choice
	PushGatewayURL string

	// Local metrics URL where metrics are fetched from, this could be ommited in the future
	// if implemented using prometheus common/expfmt instead
	MetricsURL string

	// pushgateway job name, defaults to "gin"
	Job string
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(subsystem string, requestDurationBuckets []float64, requestSizeBuckets []float64, responseSizeBuckets []float64) *Prometheus {

	p := &Prometheus{
		MetricsPath: defaultMetricPath,
		ReqCntURLLabelMappingFn: func(c *gin.Context) string {
			return c.Request.URL.String() // i.e. by default do nothing, i.e. return URL as is
		},
	}
	p.registerMetrics(subsystem, requestDurationBuckets, requestSizeBuckets, responseSizeBuckets)

	return p
}

// SetPushGateway sends metrics to a remote pushgateway exposed on pushGatewayURL
// every pushIntervalSeconds. Metrics are fetched from metricsURL
func (p *Prometheus) SetPushGateway(pushGatewayURL, metricsURL string, pushIntervalSeconds time.Duration) {
	p.Ppg.PushGatewayURL = pushGatewayURL
	p.Ppg.MetricsURL = metricsURL
	p.Ppg.PushIntervalSeconds = pushIntervalSeconds
	p.startPushTicker()
}

// SetPushGatewayJob job name, defaults to "gin"
func (p *Prometheus) SetPushGatewayJob(j string) {
	p.Ppg.Job = j
}

// SetListenAddress for exposing metrics on address. If not set, it will be exposed at the
// same address of the gin engine that is being used
func (p *Prometheus) SetListenAddress(address string) {
	p.listenAddress = address
	if p.listenAddress != "" {
		p.router = gin.Default()
	}
}

func (p *Prometheus) setMetricsPath(e *gin.Engine) {

	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, prometheusHandler())
	}
}

func (p *Prometheus) setMetricsPathWithAuth(e *gin.Engine, accounts gin.Accounts) {

	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
	}

}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go p.router.Run(p.listenAddress)
	}
}

func (p *Prometheus) getMetrics() []byte {
	response, _ := http.Get(p.Ppg.MetricsURL)

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	return body
}

func (p *Prometheus) getPushGatewayURL() string {
	h, _ := os.Hostname()
	if p.Ppg.Job == "" {
		p.Ppg.Job = "gin"
	}
	return p.Ppg.PushGatewayURL + "/metrics/job/" + p.Ppg.Job + "/instance/" + h
}

func (p *Prometheus) sendMetricsToPushGateway(metrics []byte) {
	req, err := http.NewRequest("POST", p.getPushGatewayURL(), bytes.NewBuffer(metrics))
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Printf("Error sending to push gatway: %s\n", err.Error())
	}
}

func (p *Prometheus) startPushTicker() {
	ticker := time.NewTicker(time.Second * p.Ppg.PushIntervalSeconds)
	go func() {
		for range ticker.C {
			p.sendMetricsToPushGateway(p.getMetrics())
		}
	}()
}

func (p *Prometheus) registerMetrics(subsystem string, requestDurationBuckets []float64, requestSizeBuckets []float64, responseSizeBuckets []float64) {

	p.reqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method", "url"},
	)

	if err := prometheus.Register(p.reqCnt); err != nil {
		log.Printf("requests_total counter  could not be registered: %s\n", err.Error())
	}

	p.reqDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
			Buckets:   requestDurationBuckets,
		},
		[]string{"code", "method", "url"},
	)

	if err := prometheus.Register(p.reqDur); err != nil {
		log.Printf("request_duration_seconds could not be registered: %s\n", err.Error())
	}

	p.reqSz = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
			Buckets:   requestSizeBuckets,
		},
		[]string{"code", "method", "url"},
	)

	if err := prometheus.Register(p.reqSz); err != nil {
		log.Printf("request_size_bytes could not be registered: %s\n", err.Error())
	}

	p.resSz = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
			Buckets:   responseSizeBuckets,
		},
		[]string{"code", "method", "url"},
	)

	if err := prometheus.Register(p.resSz); err != nil {
		log.Printf("response_size_bytes could not be registered: %s\n", err.Error())
	}

}

// Use adds the middleware to a gin engine.
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.handlerFunc())
	p.setMetricsPath(e)
}

// UseWithAuth adds the middleware to a gin engine with BasicAuth.
func (p *Prometheus) UseWithAuth(e *gin.Engine, accounts gin.Accounts) {
	e.Use(p.handlerFunc())
	p.setMetricsPathWithAuth(e, accounts)
}

func (p *Prometheus) handlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.String() == p.MetricsPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(c.Writer.Size())

		url := p.ReqCntURLLabelMappingFn(c)
		p.reqDur.WithLabelValues(status, c.Request.Method, url).Observe(elapsed)
		p.reqCnt.WithLabelValues(status, c.Request.Method, url).Inc()
		p.reqSz.WithLabelValues(status, c.Request.Method, url).Observe(float64(reqSz))
		p.resSz.WithLabelValues(status, c.Request.Method, url).Observe(resSz)
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
