package middleware

import (
	config "ovmsa-be/configs"
	cryptographyHelper "ovmsa-be/pkg/cryptography"
	"ovmsa-be/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenContext represents the token-related data stored in the context
type TokenContext struct {
	Token       string
	RequesterID string
	Requester   config.Requester
}

// ExtRequesterHandler is a middleware that validates external requester tokens
func ExtRequesterHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header from the request
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// If no authorization header, continue (some routes might be public)
			c.Next()
			return
		}

		// Split the header into parts to extract the token
		// Expected format: "Bearer token:requesterID:tokenData:encodedData"
		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			response.UnauthorizedResponse(c, nil, "Invalid authorization format. Expected 'Bearer <token>'")
			c.Abort()
			return
		}

		// Extract and parse the token components
		token := authParts[1]
		tokenParts := strings.Split(token, ":")
		if len(tokenParts) != 3 {
			response.UnauthorizedResponse(c, nil, "Invalid token format")
			c.Abort()
			return
		}

		// Parse token components
		requesterID := tokenParts[0]
		tokenData := tokenParts[1]
		encodedData := tokenParts[2]

		// Check if the requester exists
		requester, exists := config.ValidRequesters[requesterID]
		if !exists {
			response.UnauthorizedResponse(c, nil, "Unauthorized: Unknown requester ID")
			c.Abort()
			return
		}

		// Verify the token's cryptographic signature
		expectedCode := cryptographyHelper.Base64HMAC(tokenData, requester.SecretKey)
		if expectedCode != encodedData {
			response.UnauthorizedResponse(c, nil, "Unauthorized: Invalid token signature")
			c.Abort()
			return
		}

		// Store the validated token information in the request context
		c.Set("tokenContext", TokenContext{
			Token:       token,
			RequesterID: requesterID,
			Requester:   requester,
		})

		c.Next()
	}
}
