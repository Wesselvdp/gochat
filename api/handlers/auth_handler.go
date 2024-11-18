package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gochat/internal/auth"
	"gochat/internal/services"
	"net/http"
)

func OAuthRedirectAzure(r *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.String(http.StatusBadRequest, "Code not provided")
			return
		}

		azureToken, err := auth.GetToken(code)

		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting token: "+err.Error())
			return
		}
		userInfo, err := auth.GetUserInfo(azureToken)

		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting user info: "+err.Error())
		}

		//==  handle local user

		// Get or create user in your database
		externalID := userInfo["id"].(string)
		email := userInfo["mail"].(string) // or userInfo["userPrincipalName"] for Azure AD

		savedUser, err := services.GetOrCreateUser(email)

		if err != nil {
			fmt.Println("Error getting user: " + err.Error())
		}

		// Store user in context for this request
		auth.SetUserCookie(c, savedUser.Email)
		// ===

		// Create a JWT
		token, err := auth.CreateToken(externalID, savedUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
			return
		}
		auth.SetTokenCookie(c, token)

		c.Redirect(http.StatusMovedPermanently, "/")

	}
}

//func OAuthRedirectGoogle(r *gin.Engine) gin.HandlerFunc {
//	return func(c *gin.Context) {
//
//		code := c.Query("code")
//		if code == "" {
//			c.String(http.StatusBadRequest, "Code not provided")
//			return
//		}
//
//		googleToken, err := auth.GetToken(code)
//
//		if err != nil {
//			c.String(http.StatusInternalServerError, "Error getting token: "+err.Error())
//			return
//		}
//		userInfo, err := auth.GetUserInfo(googleToken)
//		if err != nil {
//			c.String(http.StatusInternalServerError, "Error getting user info: "+err.Error())
//		}
//
//		// Create a JWT
//		token, err := auth.CreateToken(userInfo["id"].(string))
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
//			return
//		}
//		auth.SetTokenCookie(c, token)
//
//		c.Redirect(http.StatusMovedPermanently, "/")
//
//	}
//}
