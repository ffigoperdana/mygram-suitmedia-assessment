package helpers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetContentType(c *gin.Context) string {
	return c.Request.Header.Get("Content-Type")
}

func BindRequest(c *gin.Context, destination interface{}) error {
	if strings.Contains(strings.ToLower(GetContentType(c)), "application/json") {
		return c.ShouldBindJSON(destination)
	}

	return c.ShouldBind(destination)
}

func ParseUintParam(c *gin.Context, key string) (uint, error) {
	value, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || value == 0 {
		return 0, errors.New("invalid parameter")
	}

	return uint(value), nil
}
