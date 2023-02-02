package healthcheck

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init(addr string) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	apiv1 := r.Group("/api/v1")
	{
		apiv1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success.",
				"data":    "pong",
			})
		})
	}

	r.Run(addr)
}
