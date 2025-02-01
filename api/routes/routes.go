package routes

import (
	"to-de-olho-api/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/contracts", controllers.GetContracts)
		v1.GET("/contracts/:id", controllers.GetContractByID)
		v1.GET("/contracts/latest", controllers.GetLatestContract)
		v1.POST("/contracts", controllers.CreateContract)
		v1.POST("/contracts/batch", controllers.CreateContracts)
		v1.PUT("/contracts/:id", controllers.UpdateContract)
		v1.PATCH("/contracts", controllers.UpdateContracts)
		v1.DELETE("/contracts/:id", controllers.DeleteContract)

		v1.GET("/councilors", controllers.GetCouncilors)
		v1.GET("/councilors/:id", controllers.GetCouncilorByID)
		v1.POST("/councilors", controllers.CreateCouncilor)
		v1.POST("/councilors/batch", controllers.CreateCouncilors)
		v1.PUT("/councilors/:id", controllers.UpdateCouncilor)
		v1.DELETE("/councilors/:id", controllers.DeleteCouncilor)

		v1.GET("/execution-status", controllers.GetExecutionStatus)
		v1.POST("/execution-status", controllers.LogExecution)

		v1.GET("/frequencies", controllers.GetFrequencies)
		v1.GET("/frequencies/latest", controllers.GetLatestFrequency)
		v1.GET("/frequencies/:id", controllers.GetFrequencyByID)
		v1.POST("/frequencies", controllers.CreateFrequency)
		v1.POST("/frequencies/batch", controllers.CreateFrequencies)
		v1.PUT("/frequencies/:id", controllers.UpdateFrequency)
		v1.DELETE("/frequencies/:id", controllers.DeleteFrequency)

		v1.GET("/general-productivity", controllers.GetGeneralProductivities)
		v1.GET("/general-productivity/latest", controllers.GetLatestGeneralProductivity)
		v1.GET("/general-productivity/:id", controllers.GetGeneralProductivityByID)
		v1.POST("/general-productivity", controllers.CreateGeneralProductivity)
		v1.POST("/general-productivity/batch", controllers.CreateGeneralProductivities)
		v1.PUT("/general-productivity/:id", controllers.UpdateGeneralProductivity)
		v1.DELETE("/general-productivity/:id", controllers.DeleteGeneralProductivity)

		v1.GET("/propositions", controllers.GetPropositions)
		v1.GET("/propositions/:id", controllers.GetPropositionByID)
		v1.GET("/propositions/latest", controllers.GetLatestProposition)
		v1.POST("/propositions", controllers.CreateProposition)
		v1.POST("/propositions/batch", controllers.CreatePropositions)
		v1.PUT("/propositions/:id", controllers.UpdateProposition)
		v1.DELETE("/propositions/:id", controllers.DeleteProposition)

		v1.GET("/proposition-productivity", controllers.GetPropositionProductivities)
		v1.GET("/proposition-productivity/:id", controllers.GetPropositionProductivityByID)
		v1.POST("/proposition-productivity", controllers.CreatePropositionProductivity)
		v1.PUT("/proposition-productivity/:id", controllers.UpdatePropositionProductivity)
		v1.DELETE("/proposition-productivity/:id", controllers.DeletePropositionProductivity)

		v1.GET("/travel-expenses", controllers.GetTravelExpenses)
		v1.GET("/travel-expenses/latest", controllers.GetLatestTravelExpense)
		v1.POST("/travel-expenses", controllers.CreateTravelExpense)
		v1.POST("/travel-expenses/batch", controllers.CreateTravelExpenses)
		v1.PUT("/travel-expenses/:id", controllers.UpdateTravelExpense)
		v1.DELETE("/travel-expenses/:id", controllers.DeleteTravelExpense)

		v1.GET("/health", controllers.HealthCheck)
		v1.POST("/validate-user", controllers.ValidateUser)
	}
}
