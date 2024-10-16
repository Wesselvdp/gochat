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
	"time"
)

// Todo: move to envs
const (
	clientID     = "6cff7fb3-321c-4845-b85a-4016f8ed52b9"
	clientSecret = "Daz8Q~TCX5hgRwr-dDf0muvFHVtXO36Y8VQl0bzf"
	redirectURI  = "http://localhost:8080/oauth/redirect/azure"
	tenantID     = "5159c2f7-a08e-49de-badb-221b53d4c4ed"
	scope        = "User.Read"
)

func SetTokenCookie(c *gin.Context, token string) {
	c.SetCookie(
		"token",
		token,
		3600*24, // 1 day
		"/",
		"localhost", // Change this to your domain
		false,       // Set to true if using HTTPS
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

	fmt.Println(result)
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
