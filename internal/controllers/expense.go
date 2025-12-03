package controllers

import (
	"encoding/csv"
	"expense-tracker/internal/models"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

// --- Helper Functions ---

// isValidCategory checks if the category matches the allowed list
func isValidCategory(cat string) bool {
	switch cat {
	case "Groceries", "Leisure", "Electronics", "Utilities", "Clothing", "Health", "Others":
		return true
	}
	return false
}

// --- Basic CRUD Operations ---

func CreateExpense(c *gin.Context) {
	var input models.Expense
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidCategory(input.Category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category. Allowed: Groceries, Leisure, Electronics, Utilities, Clothing, Health, Others"})
		return
	}

	// Get User ID from JWT middleware context
	userID, _ := c.Get("user_id")
	input.UserID = userID.(uint)

	// If date is not provided, default to now
	if input.Date.IsZero() {
		input.Date = time.Now()
	}

	models.DB.Create(&input)
	c.JSON(http.StatusOK, gin.H{"data": input})
}

func GetExpenses(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Filtering Logic
	filter := c.Query("filter") // "week", "month", "3months", "custom"

	db := models.DB.Where("user_id = ?", userID)
	now := time.Now()

	switch filter {
	case "week":
		lastWeek := now.AddDate(0, 0, -7)
		db = db.Where("date >= ?", lastWeek)
	case "month":
		lastMonth := now.AddDate(0, -1, 0)
		db = db.Where("date >= ?", lastMonth)
	case "3months":
		last3Months := now.AddDate(0, -3, 0)
		db = db.Where("date >= ?", last3Months)
	case "custom":
		startDate := c.Query("start_date") // Format: YYYY-MM-DD
		endDate := c.Query("end_date")
		if startDate != "" && endDate != "" {
			db = db.Where("date BETWEEN ? AND ?", startDate, endDate)
		}
	}

	var expenses []models.Expense
	db.Find(&expenses)

	c.JSON(http.StatusOK, gin.H{"data": expenses})
}

func UpdateExpense(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var expense models.Expense
	// Check if expense exists AND belongs to the logged-in user
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&expense).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found or unauthorized"})
		return
	}

	var input models.Expense
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Category != "" && !isValidCategory(input.Category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	// Prevent user from changing the Owner ID via update
	input.UserID = expense.UserID

	models.DB.Model(&expense).Updates(input)
	c.JSON(http.StatusOK, gin.H{"data": expense})
}

func DeleteExpense(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var expense models.Expense
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&expense).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found or unauthorized"})
		return
	}

	models.DB.Delete(&expense)
	c.JSON(http.StatusOK, gin.H{"data": "Expense deleted"})
}

// --- Premium Features ---

// 1. Expense Statistics
type CategoryStat struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

func GetExpenseStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var stats []CategoryStat

	// SQL: SELECT category, sum(amount) as total FROM expenses WHERE user_id = ? GROUP BY category
	// Note: We use .Scan(&stats) because the result is not a pure Expense model
	err := models.DB.Model(&models.Expense{}).
		Select("category, sum(amount) as total").
		Where("user_id = ?", userID).
		Group("category").
		Scan(&stats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// 2. Export to CSV
func ExportExpensesCSV(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var expenses []models.Expense

	// Fetch all expenses for the user (ignoring pagination for export)
	if err := models.DB.Where("user_id = ?", userID).Find(&expenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	// Set headers so the browser treats this response as a file download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=expenses.csv")

	writer := csv.NewWriter(c.Writer)

	// Write the Header Row
	if err := writer.Write([]string{"ID", "Date", "Category", "Title", "Amount"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
		return
	}

	// Loop through data and write rows
	for _, e := range expenses {
		writer.Write([]string{
			fmt.Sprintf("%d", e.ID),
			e.Date.Format("2006-01-02"),
			e.Category,
			e.Title,
			fmt.Sprintf("%.2f", e.Amount),
		})
	}
	writer.Flush()
}

// --- Payment Integration (Razorpay) ---

func CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Load keys from .env file
	key := os.Getenv("RAZORPAY_KEY")
	secret := os.Getenv("RAZORPAY_SECRET")

	if key == "" || secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment gateway not configured (Missing keys in .env)"})
		return
	}

	client := razorpay.NewClient(key, secret)

	// Amount is in Paise (1 Rupee = 100 Paise). So ₹499 = 49900
	amountInPaise := 49900

	data := map[string]interface{}{
		"amount":   amountInPaise,
		"currency": "INR",
		"receipt":  fmt.Sprintf("receipt_user_%v", userID),
	}

	body, err := client.Order.Create(data, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Razorpay Error: " + err.Error()})
		return
	}

	// Return the Order ID to the client (Frontend will use this to open the payment popup)
	c.JSON(http.StatusOK, gin.H{
		"order_id": body["id"],
		"amount":   body["amount"],
		"currency": body["currency"],
		"key_id":   key, // Sending public key to frontend is safe
	})
}
