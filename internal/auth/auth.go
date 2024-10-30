package auth

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gochat/internal/models"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Todo make a config object with checks and fallbacks
var (
	clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	tenantID     = os.Getenv("AZURE_TENANT_ID")
	clientID     = os.Getenv("AZURE_CLIENT_ID")
	scope        = "User.Read"
)

var redirectURI = func() string {
	if os.Getenv("ENV") == "production" {
		return "https://app.torgon.io/oauth/redirect/azure"
	}
	return "http://localhost:8080/oauth/redirect/azure"
}()

var domain = func() string {
	if os.Getenv("ENV") == "production" {
		return "torgon.io"
	}
	return "localhost"
}()

func SetTokenCookie(c *gin.Context, token string) {
	c.SetCookie(
		"token",
		token,
		3600*24, // 1 day
		"/",
		domain, // Change this to your domain
		false,  // Set to true if using HTTPS
		true,
	)
}

func UnsetTokenCookie(c *gin.Context) {
	c.SetCookie(
		"token",
		"",
		-1,
		"/",
		domain, // Change this to your domain
		true,   // Set to true if using HTTPS
		true,
	)
}

// Todo move to upper scope
var jwtKey = []byte("MY_S4crEt_k4y")

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func CreateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &models.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func GetUserInfo(token string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

func GetToken(code string) (string, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}

func LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?"+
			"client_id=%s&response_type=code&redirect_uri=%s&response_mode=query&scope=%s",
			tenantID, clientID, url.QueryEscape(redirectURI), url.QueryEscape(scope))
		c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}
