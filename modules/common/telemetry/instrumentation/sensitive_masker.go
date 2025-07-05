// Package instrumentation provides sensitive data masking functionality for logs and telemetry
package instrumentation

import (
	"encoding/json"
	"regexp"
	"strings"
)

// SensitiveDataMasker handles masking of sensitive information in logs
type SensitiveDataMasker struct {
	// Compiled regex patterns for PII detection
	ssnPattern        *regexp.Regexp
	phonePattern      *regexp.Regexp
	emailPattern      *regexp.Regexp
	aadharPattern     *regexp.Regexp
	dobPattern        *regexp.Regexp
	creditCardPattern *regexp.Regexp

	// Headers to fully mask
	sensitiveHeaders map[string]bool

	// Request body fields to fully mask
	sensitiveFields map[string]bool
}

// NewSensitiveDataMasker creates a new instance with default patterns
func NewSensitiveDataMasker() *SensitiveDataMasker {
	return &SensitiveDataMasker{
		// Regex patterns for PII detection
		ssnPattern:        regexp.MustCompile(`\b\d{3}-?\d{2}-?\d{4}\b`),                                         // SSN: XXX-XX-XXXX or XXXXXXXXX
		phonePattern:      regexp.MustCompile(`\b(\+?1[-.\s]?)?(\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4})\b`),         // Phone numbers
		emailPattern:      regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),             // Email addresses
		aadharPattern:     regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),                                 // Aadhaar: XXXX-XXXX-XXXX or XXXXXXXXXXXX
		dobPattern:        regexp.MustCompile(`\b(0?[1-9]|1[0-2])[/-](0?[1-9]|[12]\d|3[01])[/-](\d{4}|\d{2})\b`), // Date of birth: MM/DD/YYYY or MM-DD-YYYY
		creditCardPattern: regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),                      // Credit card numbers

		// Headers that should be fully masked
		sensitiveHeaders: map[string]bool{
			"authorization": true,
			"cookie":        true,
			"set-cookie":    true,
			"x-auth-token":  true,
			"x-api-key":     true,
			"bearer":        true,
			"token":         true,
		},

		// Request body fields that should be fully masked
		sensitiveFields: map[string]bool{
			"password":           true,
			"passwd":             true,
			"pwd":                true,
			"secret":             true,
			"token":              true,
			"api_key":            true,
			"apikey":             true,
			"private_key":        true,
			"privatekey":         true,
			"access_token":       true,
			"refresh_token":      true,
			"client_secret":      true,
			"authorization_code": true,
			"pin":                true,
			"otp":                true,
			"cvv":                true,
			"cvc":                true,
			"security_code":      true,
		},
	}
}

// MaskHeaders masks sensitive headers
func (m *SensitiveDataMasker) MaskHeaders(headers map[string][]string) map[string][]string {
	if headers == nil {
		return nil
	}

	masked := make(map[string][]string)
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if m.sensitiveHeaders[lowerKey] {
			// Fully mask sensitive headers
			masked[key] = []string{"[MASKED]"}
		} else {
			// Apply PII masking to other headers
			maskedValues := make([]string, len(values))
			for i, value := range values {
				maskedValues[i] = m.MaskPII(value)
			}
			masked[key] = maskedValues
		}
	}
	return masked
}

// MaskCookies masks all cookie values
func (m *SensitiveDataMasker) MaskCookies(cookieHeader string) string {
	if cookieHeader == "" {
		return ""
	}

	// Split cookies by semicolon
	cookies := strings.Split(cookieHeader, ";")
	var maskedCookies []string

	for _, cookie := range cookies {
		// Find the = sign to separate name and value
		if idx := strings.Index(cookie, "="); idx != -1 {
			name := strings.TrimSpace(cookie[:idx])
			maskedCookies = append(maskedCookies, name+"=[MASKED]")
		} else {
			maskedCookies = append(maskedCookies, strings.TrimSpace(cookie))
		}
	}

	return strings.Join(maskedCookies, "; ")
}

// MaskRequestBody masks sensitive fields in JSON request body
func (m *SensitiveDataMasker) MaskRequestBody(body string) string {
	if body == "" {
		return ""
	}

	// Try to parse as JSON
	var data interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// If not JSON, apply PII masking to the raw string
		return m.MaskPII(body)
	}

	// Recursively mask the JSON data
	masked := m.maskJSONData(data)

	// Convert back to JSON string
	if maskedBytes, err := json.Marshal(masked); err == nil {
		return string(maskedBytes)
	}

	// If marshal fails, return PII-masked original
	return m.MaskPII(body)
}

// maskJSONData recursively masks sensitive data in JSON structures
func (m *SensitiveDataMasker) maskJSONData(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		masked := make(map[string]interface{})
		for key, value := range v {
			lowerKey := strings.ToLower(key)
			if m.sensitiveFields[lowerKey] {
				masked[key] = "[MASKED]"
			} else {
				masked[key] = m.maskJSONData(value)
			}
		}
		return masked

	case []interface{}:
		masked := make([]interface{}, len(v))
		for i, item := range v {
			masked[i] = m.maskJSONData(item)
		}
		return masked

	case string:
		return m.MaskPII(v)

	default:
		return v
	}
}

// MaskPII masks personally identifiable information in a string
func (m *SensitiveDataMasker) MaskPII(text string) string {
	if text == "" {
		return ""
	}

	// Mask SSN (show only last 4 digits)
	text = m.ssnPattern.ReplaceAllStringFunc(text, func(match string) string {
		cleaned := strings.ReplaceAll(strings.ReplaceAll(match, "-", ""), " ", "")
		if len(cleaned) == 9 {
			return "XXX-XX-" + cleaned[5:]
		}
		return "XXX-XX-XXXX"
	})

	// Mask phone numbers (show only last 4 digits)
	text = m.phonePattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract just the digits
		digits := regexp.MustCompile(`\d`).FindAllString(match, -1)
		if len(digits) >= 10 {
			// Show last 4 digits for US numbers
			lastFour := strings.Join(digits[len(digits)-4:], "")
			return "XXX-XXX-" + lastFour
		}
		return "XXX-XXX-XXXX"
	})

	// Mask email addresses (show only domain)
	text = m.emailPattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := strings.Split(match, "@")
		if len(parts) == 2 {
			return "****@" + parts[1]
		}
		return "****@****.com"
	})

	// Mask Aadhaar numbers (show only last 4 digits)
	text = m.aadharPattern.ReplaceAllStringFunc(text, func(match string) string {
		cleaned := strings.ReplaceAll(strings.ReplaceAll(match, "-", ""), " ", "")
		if len(cleaned) == 12 {
			return "XXXX-XXXX-" + cleaned[8:]
		}
		return "XXXX-XXXX-XXXX"
	})

	// Mask dates of birth (show only year)
	text = m.dobPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Try to extract year from various formats
		if strings.Contains(match, "/") {
			parts := strings.Split(match, "/")
			if len(parts) == 3 {
				year := parts[2]
				if len(year) == 2 {
					year = "20" + year // Assume 20xx for 2-digit years
				}
				return "XX/XX/" + year
			}
		} else if strings.Contains(match, "-") {
			parts := strings.Split(match, "-")
			if len(parts) == 3 {
				year := parts[2]
				if len(year) == 2 {
					year = "20" + year
				}
				return "XX-XX-" + year
			}
		}
		return "XX/XX/XXXX"
	})

	// Mask credit card numbers (show only last 4 digits)
	text = m.creditCardPattern.ReplaceAllStringFunc(text, func(match string) string {
		cleaned := strings.ReplaceAll(strings.ReplaceAll(match, "-", ""), " ", "")
		if len(cleaned) >= 13 && len(cleaned) <= 19 {
			return "XXXX-XXXX-XXXX-" + cleaned[len(cleaned)-4:]
		}
		return "XXXX-XXXX-XXXX-XXXX"
	})

	return text
}

// MaskQueryParams masks sensitive query parameters
func (m *SensitiveDataMasker) MaskQueryParams(queryString string) string {
	if queryString == "" {
		return ""
	}

	// Split query parameters
	params := strings.Split(queryString, "&")
	var maskedParams []string

	for _, param := range params {
		if idx := strings.Index(param, "="); idx != -1 {
			key := strings.ToLower(param[:idx])
			value := param[idx+1:]

			if m.sensitiveFields[key] {
				maskedParams = append(maskedParams, param[:idx+1]+"[MASKED]")
			} else {
				maskedValue := m.MaskPII(value)
				maskedParams = append(maskedParams, param[:idx+1]+maskedValue)
			}
		} else {
			maskedParams = append(maskedParams, param)
		}
	}

	return strings.Join(maskedParams, "&")
}

// AddSensitiveField adds a custom field name to be fully masked
func (m *SensitiveDataMasker) AddSensitiveField(fieldName string) {
	m.sensitiveFields[strings.ToLower(fieldName)] = true
}

// AddSensitiveHeader adds a custom header name to be fully masked
func (m *SensitiveDataMasker) AddSensitiveHeader(headerName string) {
	m.sensitiveHeaders[strings.ToLower(headerName)] = true
}

// GetDefaultMasker returns a singleton instance of the default masker
var defaultMasker *SensitiveDataMasker

func GetDefaultMasker() *SensitiveDataMasker {
	if defaultMasker == nil {
		defaultMasker = NewSensitiveDataMasker()
	}
	return defaultMasker
}
