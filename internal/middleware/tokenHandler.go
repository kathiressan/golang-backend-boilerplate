package middleware

import (
	config "ovmsa-be/configs"
	"ovmsa-be/pkg/cryptography"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenContext represents the token-related data stored in the context
// This struct holds information about the token and the entity making the request
// Token: The complete token string
// RequesterID: Unique identifier for the entity making the request
// Requester: Configuration details of the requester including permissions and secret key
type TokenContext struct {
	Token       string
	RequesterID string
	Requester   config.Requester
}

// ExtRequesterHandler is a middleware that validates external requester tokens
// This middleware:
// 1. Checks for the presence of an Authorization header
// 2. Validates the token format (Bearer token:requesterID:tokenData:encodedData)
// 3. Verifies the requester exists in the valid requesters list
// 4. Validates the token's cryptographic signature
// 5. Stores the validated token information in the request context
func ExtRequesterHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header from the request
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// If no authorization header, continue without token validation
			c.Next()
			return
		}

		// Split the header into parts to extract the token
		// Expected format: "Bearer token:requesterID:tokenData:encodedData"
		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			// If the format is incorrect, continue without token validation
			c.Next()
			return
		}

		// Extract and parse the token components
		token := authParts[1]
		tokenParts := strings.Split(token, ":")
		if len(tokenParts) != 3 {
			// If token doesn't have the expected number of parts, mark as invalid
			c.Set("tokenError", "invalid token")
			c.Next()
			return
		}

		// Parse token components
		requesterID := tokenParts[0]  // The ID of the entity making the request
		tokenData := tokenParts[1]    // The data being signed
		encodedData := tokenParts[2]  // The HMAC signature of the token data

		// Check if the requester exists in our list of valid requesters
		requester, exists := config.ValidRequesters[requesterID]
		if !exists {
			// If requester is not recognized, mark as invalid
			c.Set("tokenError", "invalid requester (wrong id)")
			c.Next()
			return
		}

		// Verify the token's cryptographic signature
		// Generate the expected HMAC using the requester's secret key
		expectedCode := cryptography.Base64HMAC(tokenData, requester.SecretKey)
		if expectedCode != encodedData {
			// If the signature doesn't match, mark as invalid
			c.Set("tokenError", "invalid requester (wrong encoding)")
			c.Next()
			return
		}

		// Store the validated token information in the request context
		// This information can be used by subsequent handlers
		c.Set("tokenContext", TokenContext{
			Token:       token,
			RequesterID: requesterID,
			Requester:   requester,
		})

		c.Next()
	}
}
