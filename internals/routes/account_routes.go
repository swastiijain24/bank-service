package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/swastiijain24/bank/internals/handlers"
)

func RegisterAccountRoutes(r *gin.Engine, accountHandler *handlers.AccountHandler) {

	accountRoutes := r.Group("/account")
	{
		accountRoutes.GET("/discover", accountHandler.DiscoverAccounts)
		accountRoutes.POST("/mpin", accountHandler.SetMpin)
		accountRoutes.PUT("/mpin", accountHandler.ChangeMpin)
		accountRoutes.GET("/balance", accountHandler.GetBalance)
	}
}