package config

// Requester represents a valid requester configuration
type Requester struct {
	ID        string
	SecretKey string
	Role      string // Default role assigned to this service
}

// ValidRequesters is a map of valid requester IDs to their configurations
var ValidRequesters = map[string]Requester{
	// Add your valid requesters here
	// Example:
	// "requester1": {
	//     ID:        "requester1",
	//     SecretKey: "your-secret-key-here",
	// },
}
