package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func RunServer() {
	// Use default router
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		wireContext, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		}

		requestSpan := opentracing.GlobalTracer().StartSpan("users_get_request", ext.RPCServerOption(wireContext))
		defer requestSpan.Finish()

		childSpan := opentracing.GlobalTracer().StartSpan("server", opentracing.ChildOf(requestSpan.Context()))
		defer childSpan.Finish()

		c.Status(http.StatusOK)
	})

	r.Run(":10001")
}
