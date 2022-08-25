package common

import (
	"encoding/json"
	"gcron/common/zap"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func BuildResposne(c *gin.Context, data any) (err error) {
	return BuildBaseResposne(c, 0, "Success", data)

}

// if err is not nil, write log and return response.
func ChkApiErr(c *gin.Context, err error) bool {
	if err != nil {
		zap.Logf(zap.ERROR, "error:%+v", err)
		BuildBaseResposne(c, -1, "Failed", err)
		return true
	}

	return false
}

// build base response
func BuildBaseResposne(c *gin.Context, code int, msg string, data any) (err error) {
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
