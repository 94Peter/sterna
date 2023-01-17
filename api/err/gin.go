package err

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GinOutputErr(c *gin.Context, service string, err error) {
	if err == nil {
		return
	}
	if apiErr, ok := err.(ApiError); ok {
		c.AbortWithStatusJSON(apiErr.GetStatus(),
			map[string]interface{}{
				"status":   apiErr.GetStatus(),
				"title":    apiErr.GetErrorMsg(),
				"service":  service,
				"errorKey": apiErr.GetErrorKey(),
			})
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			map[string]interface{}{
				"status":   http.StatusInternalServerError,
				"title":    err.Error(),
				"service":  service,
				"errorKey": "",
			})
	}
	return
}
