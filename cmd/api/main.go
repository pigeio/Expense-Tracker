package main

import (
	"expense-tracker/internal/controllers"
	"expense-tracker/internal/middleware"
	"expense-tracker/internal/models"
	"log"
	"time" // <--- Added for CORS config

	"github.com/gin-contrib/cors" // <--- Added for CORS
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	// We ignore the error here because on Render/Cloud, we use real Env Vars, not a .env file.
	_ = godotenv.Load() 

	// 2. Connect to Database
	models.ConnectDB()

	// 3. Init Router
	r := gin.Default()

	// --- FIX CORS ERROR (Allow Frontend to talk to Backend) ---
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // Allows localhost:5173 AND your Render URL
	config.AllowMethods = []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	r.Use(cors.New(config))
	// ---------------------------------------------------------

	// 4. Public Routes (No Token required)
	public := r.Group("/api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
	}

	// 5. Protected Routes (Token required)
	protected := r.Group("/api")
	protected.Use(middleware.JwtAuthMiddleware())
	{
		// --- Basic Features (Free) ---
		protected.POST("/expenses", controllers.CreateExpense)
		protected.GET("/expenses", controllers.GetExpenses)
		protected.PUT("/expenses/:id", controllers.UpdateExpense)
		protected.DELETE("/expenses/:id", controllers.DeleteExpense)

		// --- Payment (Initiate Order) ---
		protected.POST("/pay/order", controllers.CreateOrder)

		// --- Premium Features (Token + IsPremium=true required) ---
		premium := protected.Group("/")
		premium.Use(middleware.PremiumOnly())
		{
			premium.GET("/expenses/stats", controllers.GetExpenseStats)
			premium.GET("/expenses/export", controllers.ExportExpensesCSV)
		}
	}

	// 6. Run Server
	r.Run(":8080")
}
