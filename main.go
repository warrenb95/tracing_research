package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"

	"github.com/uber/jaeger-lib/metrics"
	"github.com/warrenb95/tracing_research/internal/users"
)

func main() {
	cfg := jaegercfg.Configuration{
		ServiceName: "your_service_name",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Fatalf("can't create new tracer, %v", err)
	}
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	go func() {
		// run the other servers
		users.RunServer()
	}()

	// Use default router
	r := gin.Default()
	r.GET("/", createUser)
	r.Run(":10000")

}

func createUser(c *gin.Context) {
	// send request to endpoint
	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("GET", "http://localhost:10001/", nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		log.Fatalf("can't send post request to users api, %v", err)
	}

	// Create a span for this request
	span := opentracing.GlobalTracer().StartSpan("get_user_request")
	defer span.Finish()

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, "http://localhost:10001/")
	ext.HTTPMethod.Set(span, "GET")

	opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header))

	_, err = httpClient.Do(httpReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		log.Fatalf("error doing request, %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
