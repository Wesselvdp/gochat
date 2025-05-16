package auth

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gochat/internal/models"
	"gochat/internal/services"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Todo make a config object with checks and fallbacks
var (
	scope = "User.Read offline_access openid email profile"
)

var protocol = func() string {
	if os.Getenv("ENV") == "production" {
		return "https"
	}
	return "http"
}()

func getRedirectURI() string {
	domain := os.Getenv("DOMAIN")
	return protocol + "://" + domain + "/oauth/redirect/azure"
}

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
		3600*100,
		"/",
		domain, // Change this to your domain
		true,   // Set to true if using HTTPS
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

var jwtKey = []byte(os.Getenv("JWT_SECURITY_TOKEN"))

// Add the account name to the context
func AccountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user")

		if !exists {
			c.Next()
		}

		userService := services.NewUserService()
		userDto, err := userService.Get(c, userID.(string))

		if err != nil {
			c.Next()
		}


		c.Set("account_name", userDto.Account.Name)
		c.Next()
	}
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenString, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Return the key we used to sign the token
			return jwtKey, nil
		})

		if err != nil {
			fmt.Printf("Token validation error: %v\n", err)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if !token.Valid {
			fmt.Println("Token is invalid")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if err != nil || !token.Valid {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Make sure user's can't go back in browser history
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// Get the user ID from the token claims
		claims := token.Claims.(jwt.MapClaims)


		c.Set("user", claims["sub"])
		c.Next()
	}
}

func CreateToken(userID string) (string, error) {
	expirationTime := time.Now().Add(86 * time.Hour)
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
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/token")
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", getRedirectURI())
	data.Set("scope", scope) // Add this line

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", fmt.Errorf("error making token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading token response body: %w", err)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing token response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("error getting token: %s (%s)", result.Error, result.ErrorDesc)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("token not found in response")
	}

	return result.AccessToken, nil
}

func LoginHandler() gin.HandlerFunc {
	clientID := os.Getenv("AZURE_CLIENT_ID")

	return func(c *gin.Context) {
		authURL := fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/authorize?"+
			"client_id=%s&response_type=code&redirect_uri=%s&response_mode=query&scope=%s",
			clientID, url.QueryEscape(getRedirectURI()), url.QueryEscape(scope))
		c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}
