package mask

import (
	"errors"
	"strings"
)

// maskEmail masks the given email keeping the first & last N chars of the username, plus domain.
func maskEmail(arg string, value string) (string, error) {
	// Split the email address into local part and domain
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return "", errors.New("Invalid email format")
	}

	username := parts[0]
	domain := parts[1]

	// Mask the username, leaving the first  3 character and appending asterisks
	// Calculate the number of characters to keep at the beginning and end
	keepChars := len(username) / 4
	if keepChars < 1 {
		keepChars = 1 // Ensure at least one character is kept at each end
	}

	// Calculate the number of characters to replace with asterisks
	numAsterisks := len(username) - 2*keepChars
	if numAsterisks < 1 {
		return "*@" + domain, nil
	}

	maskedUsername := username[:keepChars] + strings.Repeat("*", numAsterisks) + username[len(username)-keepChars:]

	// Combine the masked username and domain back into an email address
	maskedEmail := maskedUsername + "@" + domain

	return maskedEmail, nil
}

// maskPhone masks a phone number, keeping the country code and the last four digits visible.
func maskPhone(arg string, value string) (string, error) {
	countryCodeLen := strings.Index(value[1:], "+") + 6
	if countryCodeLen < 1 {
		countryCodeLen = 1 // Assuming at least one character for the country code
	}

	// Keeping the country code and last four digits
	countryCode := value[:countryCodeLen]
	lastFour := value[len(value)-2:]

	// Mask the middle part of the phone number
	midSection := strings.Repeat("*", len(value)-countryCodeLen-1)

	return countryCode + midSection + lastFour, nil
}

// maskName masks all but the first name in a given full name string.
func maskName(arg string, value string) (string, error) {
	// Split the name into parts
	parts := strings.Fields(value)

	// Return the original name if it's just one word
	if len(parts) < 2 {
		return value, nil
	}

	// Keep the first name intact
	maskedName := parts[0]

	// Mask the remaining parts
	for _, part := range parts[1:] {
		maskedName += " " + strings.Repeat("*", len(part))
	}

	return maskedName, nil
}
