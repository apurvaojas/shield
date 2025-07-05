package instrumentation

import (
	"testing"
)

func TestSensitiveDataMasker_MaskHeaders(t *testing.T) {
	masker := NewSensitiveDataMasker()

	tests := []struct {
		name     string
		headers  map[string][]string
		expected map[string][]string
	}{
		{
			name: "mask authorization header",
			headers: map[string][]string{
				"Authorization": {"Bearer token123"},
				"Content-Type":  {"application/json"},
			},
			expected: map[string][]string{
				"Authorization": {"[MASKED]"},
				"Content-Type":  {"application/json"},
			},
		},
		{
			name: "mask cookie header",
			headers: map[string][]string{
				"Cookie":     {"session=abc123; user=john"},
				"User-Agent": {"Mozilla/5.0"},
			},
			expected: map[string][]string{
				"Cookie":     {"[MASKED]"},
				"User-Agent": {"Mozilla/5.0"},
			},
		},
		{
			name: "mask multiple sensitive headers",
			headers: map[string][]string{
				"Authorization": {"Bearer token123"},
				"X-API-Key":     {"api_key_123"},
				"Cookie":        {"session=abc123"},
				"Content-Type":  {"application/json"},
			},
			expected: map[string][]string{
				"Authorization": {"[MASKED]"},
				"X-API-Key":     {"[MASKED]"},
				"Cookie":        {"[MASKED]"},
				"Content-Type":  {"application/json"},
			},
		},
		{
			name:     "nil headers",
			headers:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.MaskHeaders(tt.headers)

			if tt.expected == nil && result != nil {
				t.Errorf("expected nil but got %v", result)
				return
			}

			if tt.expected != nil && result == nil {
				t.Errorf("expected %v but got nil", tt.expected)
				return
			}

			for key, expectedValues := range tt.expected {
				actualValues, exists := result[key]
				if !exists {
					t.Errorf("expected header %s not found", key)
					continue
				}

				if len(actualValues) != len(expectedValues) {
					t.Errorf("header %s: expected %d values, got %d", key, len(expectedValues), len(actualValues))
					continue
				}

				for i, expectedValue := range expectedValues {
					if actualValues[i] != expectedValue {
						t.Errorf("header %s[%d]: expected %s, got %s", key, i, expectedValue, actualValues[i])
					}
				}
			}
		})
	}
}

func TestSensitiveDataMasker_MaskCookies(t *testing.T) {
	masker := NewSensitiveDataMasker()

	tests := []struct {
		name     string
		cookies  string
		expected string
	}{
		{
			name:     "single cookie",
			cookies:  "session=abc123",
			expected: "session=[MASKED]",
		},
		{
			name:     "multiple cookies",
			cookies:  "session=abc123; user=john; token=xyz789",
			expected: "session=[MASKED]; user=[MASKED]; token=[MASKED]",
		},
		{
			name:     "cookies with spaces",
			cookies:  "session=abc123 ; user=john ; token=xyz789",
			expected: "session=[MASKED]; user=[MASKED]; token=[MASKED]",
		},
		{
			name:     "empty cookie header",
			cookies:  "",
			expected: "",
		},
		{
			name:     "malformed cookie (no value)",
			cookies:  "session",
			expected: "session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.MaskCookies(tt.cookies)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSensitiveDataMasker_MaskPII(t *testing.T) {
	masker := NewSensitiveDataMasker()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mask SSN with hyphens",
			input:    "My SSN is 123-45-6789",
			expected: "My SSN is XXX-XX-6789",
		},
		{
			name:     "mask SSN without hyphens",
			input:    "SSN: 123456789",
			expected: "SSN: XXX-XX-6789",
		},
		{
			name:     "mask phone number",
			input:    "Call me at (555) 123-4567",
			expected: "Call me at (XXX-XXX-4567",
		},
		{
			name:     "mask email address",
			input:    "Contact john.doe@example.com for help",
			expected: "Contact ****@example.com for help",
		},
		{
			name:     "mask Aadhaar number",
			input:    "Aadhaar: 1234-5678-9012",
			expected: "Aadhaar: XXXX-XXXX-9012",
		},
		{
			name:     "mask date of birth",
			input:    "DOB: 01/15/1990",
			expected: "DOB: XX/XX/1990",
		},
		{
			name:     "mask credit card",
			input:    "Card: 4532-1234-5678-9012",
			expected: "Card: XXXX-XXXX-5678-9012",
		},
		{
			name:     "multiple PII types",
			input:    "SSN: 123-45-6789, Email: user@test.com, Phone: 555-123-4567",
			expected: "SSN: XXX-XX-6789, Email: ****@test.com, Phone: XXX-XXX-4567",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no PII",
			input:    "This is a normal string with no PII",
			expected: "This is a normal string with no PII",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.MaskPII(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSensitiveDataMasker_MaskRequestBody(t *testing.T) {
	masker := NewSensitiveDataMasker()

	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "JSON with password field",
			body:     `{"username": "john", "password": "secret123", "email": "john@example.com"}`,
			expected: `{"email":"****@example.com","password":"[MASKED]","username":"john"}`,
		},
		{
			name:     "JSON with token field",
			body:     `{"user": "john", "token": "abc123", "data": "normal"}`,
			expected: `{"data":"normal","token":"[MASKED]","user":"john"}`,
		},
		{
			name:     "non-JSON body with PII",
			body:     "username=john&password=secret&email=user@test.com",
			expected: "username=john&password=secret&email=****@test.com",
		},
		{
			name:     "empty body",
			body:     "",
			expected: "",
		},
		{
			name:     "nested JSON",
			body:     `{"user": {"name": "john", "credentials": {"password": "secret", "api_key": "key123"}}}`,
			expected: `{"user":{"credentials":{"api_key":"[MASKED]","password":"[MASKED]"},"name":"john"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.MaskRequestBody(tt.body)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSensitiveDataMasker_MaskQueryParams(t *testing.T) {
	masker := NewSensitiveDataMasker()

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "query with password",
			query:    "user=john&password=secret123&page=1",
			expected: "user=john&password=[MASKED]&page=1",
		},
		{
			name:     "query with token",
			query:    "action=login&token=abc123&redirect=home",
			expected: "action=login&token=[MASKED]&redirect=home",
		},
		{
			name:     "query with email PII",
			query:    "search=user@example.com&type=email",
			expected: "search=****@example.com&type=email",
		},
		{
			name:     "empty query",
			query:    "",
			expected: "",
		},
		{
			name:     "query without sensitive data",
			query:    "page=1&limit=10&sort=name",
			expected: "page=1&limit=10&sort=name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := masker.MaskQueryParams(tt.query)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSensitiveDataMasker_AddCustomFields(t *testing.T) {
	masker := NewSensitiveDataMasker()

	// Add a custom sensitive field
	masker.AddSensitiveField("custom_secret")

	// Test that the custom field is now masked
	body := `{"user": "john", "custom_secret": "secret_value", "normal_field": "normal_value"}`
	result := masker.MaskRequestBody(body)

	// The custom_secret should be masked
	if !contains(result, `"custom_secret":"[MASKED]"`) {
		t.Errorf("expected custom_secret to be masked, got %s", result)
	}

	// Normal field should not be masked
	if !contains(result, `"normal_field":"normal_value"`) {
		t.Errorf("expected normal_field to remain unmasked, got %s", result)
	}
}

func TestSensitiveDataMasker_AddCustomHeaders(t *testing.T) {
	masker := NewSensitiveDataMasker()

	// Add a custom sensitive header
	masker.AddSensitiveHeader("X-Custom-Token")

	headers := map[string][]string{
		"X-Custom-Token": {"secret_token_123"},
		"Content-Type":   {"application/json"},
	}

	result := masker.MaskHeaders(headers)

	// Custom header should be masked
	if result["X-Custom-Token"][0] != "[MASKED]" {
		t.Errorf("expected X-Custom-Token to be masked, got %s", result["X-Custom-Token"][0])
	}

	// Normal header should not be masked
	if result["Content-Type"][0] != "application/json" {
		t.Errorf("expected Content-Type to remain unmasked, got %s", result["Content-Type"][0])
	}
}

func TestGetDefaultMasker(t *testing.T) {
	// Test that GetDefaultMasker returns the same instance (singleton)
	masker1 := GetDefaultMasker()
	masker2 := GetDefaultMasker()

	if masker1 != masker2 {
		t.Error("expected GetDefaultMasker to return the same instance (singleton)")
	}

	// Test that the default masker is properly initialized
	if masker1 == nil {
		t.Error("expected GetDefaultMasker to return a non-nil masker")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
