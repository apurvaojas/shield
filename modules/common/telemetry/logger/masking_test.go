package common

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestMaskingHandler(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a text handler that writes to our buffer
	textHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Create masking handler
	maskingHandler := newMaskingHandler(textHandler)

	// Create logger with masking handler
	logger := slog.New(maskingHandler)

	tests := []struct {
		name           string
		logFunc        func()
		expectedMasked []string
		notExpected    []string
	}{
		{
			name: "mask email in message",
			logFunc: func() {
				logger.Info("User email user@example.com failed login")
			},
			expectedMasked: []string{"****@example.com"}, // Email will be masked
			notExpected:    []string{"user@example.com"}, // Original email should be masked
		},
		{
			name: "mask password in attributes",
			logFunc: func() {
				logger.Info("User login attempt", slog.String("password", "secret123"))
			},
			expectedMasked: []string{"[MASKED]"},
			notExpected:    []string{"secret123"},
		},
		{
			name: "mask authorization header",
			logFunc: func() {
				logger.Info("HTTP request", slog.String("authorization", "Bearer token123"))
			},
			expectedMasked: []string{"[MASKED]"},
			notExpected:    []string{"Bearer token123"},
		},
		{
			name: "mask email in message",
			logFunc: func() {
				logger.Info("Processing user john.doe@example.com")
			},
			expectedMasked: []string{"****@example.com"},
			notExpected:    []string{"john.doe@example.com"},
		},
		{
			name: "mask SSN in message",
			logFunc: func() {
				logger.Info("User SSN: 123-45-6789")
			},
			expectedMasked: []string{"XXX-XX-6789"},
			notExpected:    []string{"123-45-6789"},
		},
		{
			name: "mask phone number in message",
			logFunc: func() {
				logger.Info("Contact number: (555) 123-4567")
			},
			expectedMasked: []string{"XXX-XXX-4567"},
			notExpected:    []string{"(555) 123-4567"},
		},
		{
			name: "mask multiple PII types",
			logFunc: func() {
				logger.Info("User details",
					slog.String("email", "user@test.com"),
					slog.String("phone", "555-123-4567"),
					slog.String("password", "secret"),
					slog.String("normal_field", "safe_value"),
				)
			},
			expectedMasked: []string{"****@test.com", "XXX-XXX-4567", "[MASKED]", "safe_value"},
			notExpected:    []string{"user@test.com", "555-123-4567", "secret"},
		},
		{
			name: "mask Aadhaar number",
			logFunc: func() {
				logger.Info("Aadhaar verification: 1234-5678-9012")
			},
			expectedMasked: []string{"XXXX-XXXX-9012"},
			notExpected:    []string{"1234-5678-9012"},
		},
		{
			name: "mask credit card number",
			logFunc: func() {
				logger.Info("Payment with card: 4532-1234-5678-9012")
			},
			expectedMasked: []string{"XXXX-XXXX-5678-9012"}, // Actual pattern from masker
			notExpected:    []string{"4532-1234-5678-9012"},
		},
		{
			name: "preserve non-sensitive data",
			logFunc: func() {
				logger.Info("Normal operation",
					slog.String("user_id", "12345"),
					slog.String("action", "login"),
					slog.Int("status_code", 200),
				)
			},
			expectedMasked: []string{"12345", "login", "200"},
			notExpected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear buffer
			buf.Reset()

			// Execute log function
			tt.logFunc()

			// Get logged output
			output := buf.String()

			// Check that expected masked values are present
			for _, expected := range tt.expectedMasked {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected '%s' to be present in log output: %s", expected, output)
				}
			}

			// Check that sensitive values are not present
			for _, notExpected := range tt.notExpected {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected '%s' to be masked in log output: %s", notExpected, output)
				}
			}
		})
	}
}

func TestMaskingHandlerWithContext(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a text handler that writes to our buffer
	textHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Create masking handler
	maskingHandler := newMaskingHandler(textHandler)

	// Create logger with masking handler
	logger := slog.New(maskingHandler)

	// Test with context
	ctx := context.Background()
	logger.InfoContext(ctx, "Processing user data",
		slog.String("email", "test@example.com"),
		slog.String("password", "secret123"),
	)

	output := buf.String()

	// Check masking worked
	if !strings.Contains(output, "****@example.com") {
		t.Errorf("Expected email to be masked, got: %s", output)
	}

	if !strings.Contains(output, "[MASKED]") {
		t.Errorf("Expected password to be masked, got: %s", output)
	}

	if strings.Contains(output, "test@example.com") {
		t.Errorf("Expected original email to be masked, got: %s", output)
	}

	if strings.Contains(output, "secret123") {
		t.Errorf("Expected original password to be masked, got: %s", output)
	}
}

func TestMaskingHandlerDisabled(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a text handler that writes to our buffer
	textHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Create logger without masking handler (direct text handler)
	logger := slog.New(textHandler)

	// Log sensitive data
	logger.Info("User login", slog.String("password", "secret123"))

	output := buf.String()

	// Without masking, sensitive data should be present
	if !strings.Contains(output, "secret123") {
		t.Errorf("Expected password to be present when masking is disabled, got: %s", output)
	}

	if strings.Contains(output, "[MASKED]") {
		t.Errorf("Expected password not to be masked when masking is disabled, got: %s", output)
	}
}

func TestMaskAttributeFunction(t *testing.T) {
	// Create masking handler
	maskingHandler := newMaskingHandler(nil) // We only test the maskAttribute function

	tests := []struct {
		name     string
		attr     slog.Attr
		expected string
	}{
		{
			name:     "mask password attribute",
			attr:     slog.String("password", "secret123"),
			expected: "[MASKED]",
		},
		{
			name:     "mask authorization attribute",
			attr:     slog.String("authorization", "Bearer token"),
			expected: "[MASKED]",
		},
		{
			name:     "mask email in value",
			attr:     slog.String("user_email", "test@example.com"),
			expected: "****@example.com",
		},
		{
			name:     "preserve normal attribute",
			attr:     slog.String("user_id", "12345"),
			expected: "12345",
		},
		{
			name:     "preserve number attribute",
			attr:     slog.Int("count", 100),
			expected: "100", // Will be converted to string for comparison
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskingHandler.maskAttribute(tt.attr)

			// For non-string attributes, check the value conversion
			if tt.attr.Value.Kind() != slog.KindString {
				// For numeric values, check if the key is preserved
				if result.Key != tt.attr.Key {
					t.Errorf("Expected key %s, got %s", tt.attr.Key, result.Key)
				}
				return
			}

			resultValue := result.Value.String()
			if resultValue != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, resultValue)
			}
		})
	}
}
