package ginlogrus

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/almonteb/logrus"
	"github.com/gin-gonic/gin"
)

// 2016-09-27 09:38:21.541541811 +0200 CEST
// 127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700]
// "GET /apache_pb.gif HTTP/1.0" 200 2326
// "http://www.example.com/start.html"
// "Mozilla/4.08 [en] (Win98; I ;Nav)"

type fieldKey string

type FieldMap map[fieldKey]string

const (
	FieldKeyHostname   = "hostname"
	FieldKeyStatusCode = "statusCode"
	FieldKeyLatency    = "latency"
	FieldKeyClientIp   = "ClientIP"
	FieldKeyMethod     = "method"
	FieldKeyPath       = "path"
	FieldKeyReferrer   = "referrer"
	FieldKeyDataLength = "dataLength"
	FieldKeyUserAgent  = "userAgent"
)

type GinLogrusConfig struct {
	TimeFormat string
	FieldMap   FieldMap
}

var timeFormat = "02/Jan/2006:15:04:05 -0700"

var DefaultFieldMap = FieldMap{
	FieldKeyHostname:   "hostname",
	FieldKeyStatusCode: "statusCode",
	FieldKeyLatency:    "latency",
	FieldKeyClientIp:   "ClientIP",
	FieldKeyMethod:     "method",
	FieldKeyPath:       "path",
	FieldKeyReferrer:   "referer",
	FieldKeyDataLength: "dataLength",
	FieldKeyUserAgent:  "userAgent",
}

var DefaultGinLogrusConfig = GinLogrusConfig{
	TimeFormat: timeFormat,
	FieldMap:   DefaultFieldMap,
}

func Logger(log *logrus.Logger) gin.HandlerFunc {
	return LoggerWithConfig(log, DefaultGinLogrusConfig)
}

// Logger is the logrus logger handler
func LoggerWithConfig(log *logrus.Logger, cfg GinLogrusConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		entry := logrus.NewEntry(log).WithFields(logrus.Fields{
			cfg.FieldMap[FieldKeyHostname]:   hostname,
			cfg.FieldMap[FieldKeyStatusCode]: statusCode,
			cfg.FieldMap[FieldKeyLatency]:    latency, // time to process
			cfg.FieldMap[FieldKeyClientIp]:   clientIP,
			cfg.FieldMap[FieldKeyMethod]:     c.Request.Method,
			cfg.FieldMap[FieldKeyPath]:       path,
			cfg.FieldMap[FieldKeyReferrer]:   referer,
			cfg.FieldMap[FieldKeyDataLength]: dataLength,
			cfg.FieldMap[FieldKeyUserAgent]:  clientUserAgent,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" \"%s\" (%dms)", clientIP, hostname, time.Now().Format(cfg.TimeFormat), c.Request.Method, path, statusCode, dataLength, referer, clientUserAgent, latency)
			if statusCode > 499 {
				entry.Error(msg)
			} else if statusCode > 399 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
