package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

// Initialize client (use Env variables in production!)
var client = razorpay.NewClient("YOUR_KEY_ID", "YOUR_KEY_SECRET")

func CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Create an order for ₹499 (Amount is in Paise, so 49900)
	data := map[string]interface{}{
		"amount":   49900,
		"currency": "INR",
		"receipt":  "receipt_id_" + fmt.Sprint(userID),
	}

	body, err := client.Order.Create(data, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "Payment gateway error"})
		return
	}

	// Send the Order ID to frontend (React/Mobile) to open the Payment Popup
	c.JSON(200, gin.H{
		"order_id": body["id"],
		"amount":   body["amount"],
		"key_id":   "YOUR_KEY_ID",
	})
}

// Webhook: Razorpay calls THIS URL when payment succeeds
func HandlePaymentWebhook(c *gin.Context) {
	// 1. Verify signature (Security step to ensure request is from Razorpay)
	// 2. If valid, update database:

	// UPDATE users SET is_premium = true WHERE ...

	c.JSON(200, gin.H{"status": "ok"})
}
