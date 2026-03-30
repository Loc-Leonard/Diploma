package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/models"
)

func MustChangePasswordMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()

		if path == "auth/change-password" {
			c.Next()
			return
		}
		userID := UserIDFromContext(c)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var user models.User
		if err := db.Select("must_change_password").First(&user, userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		if user.MustChangePassword {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "must_change_password"})
			return
		}

		c.Next()
	}
}
