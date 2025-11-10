package domain

import (
	"errors"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Email represents a validated email address value object.
// It enforces format validation and normalizes casing.
type Email struct {
	value string
}

// NewEmail validates and constructs an Email.
func NewEmail(address string) (*Email, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, errors.New("email address cannot be empty")
	}
	if !emailRegex.MatchString(address) {
		return nil, errors.New("invalid email format")
	}
	return &Email{value: strings.ToLower(address)}, nil
}

// String returns the normalized email.
func (e Email) String() string { return e.value }

// Value returns the underlying value.
func (e Email) Value() string { return e.value }
