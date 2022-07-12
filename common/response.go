package common

import (
	"crontab/common/zap"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func BuildResposne(c *gin.Context, code int, msg string, data any) (err error) {
	if c.IsAborted() {
		return
	}

	var (
		retValue []byte
	)
	resp := &Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}

	if retValue, err = json.Marshal(resp); err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.Data(200, gin.MIMEJSON, retValue)
	return
}

// if err is not nil, write log and return response.
func ChkApiErr(c *gin.Context, err error) bool {
	if err != nil {
		zap.Zlogger.Infof("error:%v", err)
		BuildResposne(c, -1, "falied", err)
		return true
	}

	return false
}
