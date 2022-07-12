package middleware

import (
	"bytes"
	"crontab/common/zap"
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LoggerWithWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start).Seconds()

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		header := c.Request.Header
		post := c.Request.PostForm.Encode()
		get := c.Request.URL.Query().Encode()

		Layout := "2006-01-02 15:04:05"
		dataTimeStr := time.Unix(start.Unix(), 0).Format(Layout)

		headerByte, _ := json.Marshal(header)
		headerString := string(headerByte)
		headerString = strings.Replace(headerString, " ", "", -1)
		if raw != "" {
			path = path + "?" + raw
		}

		zap.Zlogger.Infow("request info",
			"date_time", dataTimeStr,
			"client_ip", clientIP,
			"status_code", statusCode,
			"request_method", method,
			"uri", path,
			"header", headerString,
			"response_body", blw.body.String(),
			"server_ip", localIP(),
			"post data", post,
			"get data", get,
			"latency", latency,
			"comment", comment,
		)
	}
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
