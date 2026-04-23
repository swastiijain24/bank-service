package middlewares

import (
	"os"

	"github.com/gin-gonic/gin"
)

func InternalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-INTERNAL-API-KEY")

		if key != os.Getenv("INTERNAL_API_KEY") {
			c.AbortWithStatus(403)
			return
		}

		c.Next()
	}
}