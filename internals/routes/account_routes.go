package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/swastiijain24/bank/internals/handlers"
	"github.com/swastiijain24/bank/internals/middlewares"
)

func RegisterAccountRoutes(r *gin.Engine, accountHandler *handlers.AccountHandler) {

	accountRoutes := r.Group("/account")
	accountRoutes.Use(middlewares.InternalAuthMiddleware())
	{
		accountRoutes.POST("/discover", accountHandler.DiscoverAccounts)
		accountRoutes.POST("/mpin", accountHandler.SetMpin)
		accountRoutes.PUT("/mpin", accountHandler.ChangeMpin)
		accountRoutes.POST("/balance", accountHandler.GetBalance)
	}
}