package handlerHelpers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPaginationFromRequest(ctx *gin.Context) (start int, count int) {
	start, err := strconv.Atoi(ctx.Query("start"))
	if err != nil {
		start = 0
	}
	count, err = strconv.Atoi(ctx.Query("count"))
	if err != nil {
		count = 10
	}
	return start, count
}
