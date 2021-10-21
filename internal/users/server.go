package users

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

		requestSpan := tracer.StartSpan("users_get_request", ext.RPCServerOption(wireContext))
		defer requestSpan.Finish()

		childSpan := tracer.StartSpan("server", opentracing.ChildOf(requestSpan.Context()))
		defer childSpan.Finish()

		httpClient := &http.Client{}
		httpReq, err := http.NewRequest("GET", "http://localhost:10002/", nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			log.Fatalf("can't send post request to users api, %v", err)
		}

		ext.SpanKindRPCClient.Set(childSpan)
		ext.HTTPUrl.Set(childSpan, "http://localhost:10002/")
		ext.HTTPMethod.Set(childSpan, "GET")

		tracer.Inject(childSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header))

		time.Sleep(50 * time.Millisecond)

		_, err = httpClient.Do(httpReq)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			log.Fatalf("error doing request, %v", err)
		}

		c.Status(http.StatusOK)
	})

	r.Run(":10001")
}
