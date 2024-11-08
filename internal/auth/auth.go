package auth

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Get the tenant ID from the token
			tenantID := token.Claims.(jwt.MapClaims)["tid"].(string)

			// Fetch the OpenID Connect metadata document for the tenant
			oidcMetadataURL := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0/.well-known/openid-configuration", tenantID)
			resp, err := http.Get(oidcMetadataURL)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			var metadata struct {
				JwksURI string `json:"jwks_uri"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
				return nil, err
			}

			// Fetch the signing keys from the JWKS endpoint
			resp, err = http.Get(metadata.JwksURI)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			var jwks struct {
				Keys []struct {
					Kid string   `json:"kid"`
					X5C []string `json:"x5c"`
				} `json:"keys"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
				return nil, err
			}

			// Find the correct signing key
			var signingKey interface{}
			for _, key := range jwks.Keys {
				if key.Kid == token.Header["kid"].(string) {
					// Decode the key from base64
					block, _ := pem.Decode([]byte(key.X5C[0]))
					if block == nil {
						return nil, fmt.Errorf("failed to parse PEM block containing the key")
					}

					cert, err := x509.ParseCertificate(block.Bytes)
					if err != nil {
						return nil, err
					}

					signingKey = cert.PublicKey
					break
				}
			}

			if signingKey == nil {
				return nil, fmt.Errorf("signing key not found")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, fmt.Errorf("invalid token claims")
			}

			// Validate the expiration time
			if float64(time.Now().Unix()) > claims["exp"].(float64) {
				return nil, fmt.Errorf("token is expired")
			}

			// Validate the issued at time
			if float64(time.Now().Unix()) < claims["iat"].(float64) {
				return nil, fmt.Errorf("token used before issued")
			}

			// Validate the not before time
			if float64(time.Now().Unix()) < claims["nbf"].(float64) {
				return nil, fmt.Errorf("token not yet valid")
			}

			// Validate the issuer
			iss := token.Claims.(jwt.MapClaims)["iss"].(string)
			if iss != fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID) {
				return nil, fmt.Errorf("invalid issuer: %s", iss)
			}

			return signingKey, nil

			//return jwtKey, nil
		})

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
		userID := claims["sub"].(string)
		c.Set("userID", userID)
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
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/token")
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

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
	return func(c *gin.Context) {
		authURL := fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/authorize?"+
			"client_id=%s&response_type=code&redirect_uri=%s&response_mode=query&scope=%s",
			clientID, url.QueryEscape(redirectURI), url.QueryEscape(scope))
		c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}
