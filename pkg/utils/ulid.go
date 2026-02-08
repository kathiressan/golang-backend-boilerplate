package utils

import (
	"crypto/rand"
	"io"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropy io.Reader
	once    sync.Once
)

// NewULID generates a new, cryptographically secure, sortable ULID.
func NewULID() string {
	once.Do(func() {
		entropy = rand.Reader
	})

	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return id.String()
}
