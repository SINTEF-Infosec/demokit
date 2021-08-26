package core

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"time"
)

func NewNodeRouter(logger *log.Entry) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		logger.WithField("component", "router").WithFields(log.Fields{
			"status_code": statusCode,
			"latency":     latencyTime,
			"client_ip":   clientIP,
			"method":      reqMethod,
			"uri":         reqUri,
		}).Info()
	})
	return r
}
