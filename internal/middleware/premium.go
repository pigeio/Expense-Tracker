package middleware

import (
	"expense-tracker/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PremiumOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		var user models.User
		// Check the database to see if the user is premium
		if err := models.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if !user.IsPremium {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "This is a Premium feature. Please upgrade!",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
