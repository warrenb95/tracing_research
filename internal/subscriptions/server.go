package subscriptions

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/warrenb95/tracing_research/internal/tracer"
)

func RunServer() {
	// Use default router
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		// Create a span for this request
		tracer, close, err := tracer.Create("users_service")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			log.Fatalf("error creating tracer, %v", err)
		}
		defer close.Close()

		wireContext, err := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		}

		requestSpan := tracer.StartSpan("subscription_request", ext.RPCServerOption(wireContext))
		defer requestSpan.Finish()

		time.Sleep(50 * time.Millisecond)

		c.Status(http.StatusOK)
	})

	r.Run(":10002")
}
