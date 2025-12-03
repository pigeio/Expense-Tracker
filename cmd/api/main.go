package main

import (
	"expense-tracker/internal/controllers"
	"expense-tracker/internal/middleware"
	"expense-tracker/internal/models"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Connect to Database
	models.ConnectDB()

	// 3. Init Router
	r := gin.Default()

	// 4. Public Routes (No Token required)
	public := r.Group("/api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		// If you have a Webhook for Razorpay, it goes here (Public)
		// public.POST("/pay/webhook", controllers.HandlePaymentWebhook)
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
		// Anyone logged in can try to pay
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
