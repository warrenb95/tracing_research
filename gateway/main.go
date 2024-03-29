package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/warrenb95/tracing_research/tracer"
)

func main() {
	// Use default router
	r := gin.Default()
	r.GET("/", createUser)
	r.Run(":10000")

}

func createUser(c *gin.Context) {
	// Create a span for this request
	tracer, close, err := tracer.Create("gateway_service")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		log.Fatalf("error creating tracer, %v", err)
	}
	defer close.Close()

	// send request to endpoint
	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("GET", "http://localhost:10001/", nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		log.Fatalf("can't send post request to users api, %v", err)
	}

	span := tracer.StartSpan("get_user_request")
	defer span.Finish()

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, "http://localhost:10001/")
	ext.HTTPMethod.Set(span, "GET")

	tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header))

	time.Sleep(25 * time.Millisecond)

	_, err = httpClient.Do(httpReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		log.Fatalf("error doing request, %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
