package pp

import (
	"testing"
)

func TestNewPolarPaymentProvider(t *testing.T) {
	// Test with a dummy access token (should default to sandbox)
	provider, err := NewPolarPaymentProvider("test_token")
	if err != nil {
		t.Fatalf("Failed to create Polar payment provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider is nil")
	}

	if provider.AccessToken != "test_token" {
		t.Errorf("Expected access token 'test_token', got '%s'", provider.AccessToken)
	}

	if provider.Server != "sandbox" {
		t.Errorf("Expected server 'sandbox', got '%s'", provider.Server)
	}

	// Test sandbox methods
	if !provider.IsSandbox() {
		t.Error("Expected provider to be in sandbox mode")
	}

	if provider.GetEnvironment() != "sandbox" {
		t.Errorf("Expected environment 'sandbox', got '%s'", provider.GetEnvironment())
	}

	expectedURL := "https://sandbox-api.polar.sh"
	if provider.GetServerURL() != expectedURL {
		t.Errorf("Expected server URL '%s', got '%s'", expectedURL, provider.GetServerURL())
	}
}

func TestNewPolarPaymentProviderWithEnv(t *testing.T) {
	tests := []struct {
		name           string
		environment    string
		expectedServer string
		expectedIsProd bool
		expectedURL    string
		expectError    bool
	}{
		{
			name:           "sandbox environment",
			environment:    "sandbox",
			expectedServer: "sandbox",
			expectedIsProd: false,
			expectedURL:    "https://sandbox-api.polar.sh",
			expectError:    false,
		},
		{
			name:           "test environment",
			environment:    "test",
			expectedServer: "sandbox",
			expectedIsProd: false,
			expectedURL:    "https://sandbox-api.polar.sh",
			expectError:    false,
		},
		{
			name:           "production environment",
			environment:    "production",
			expectedServer: "production",
			expectedIsProd: true,
			expectedURL:    "https://api.polar.sh",
			expectError:    false,
		},
		{
			name:           "prod environment",
			environment:    "prod",
			expectedServer: "production",
			expectedIsProd: true,
			expectedURL:    "https://api.polar.sh",
			expectError:    false,
		},
		{
			name:        "invalid environment",
			environment: "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewPolarPaymentProviderWithEnv("test_token", tt.environment)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create Polar payment provider: %v", err)
			}

			if provider.Server != tt.expectedServer {
				t.Errorf("Expected server '%s', got '%s'", tt.expectedServer, provider.Server)
			}

			if provider.isProd != tt.expectedIsProd {
				t.Errorf("Expected isProd %v, got %v", tt.expectedIsProd, provider.isProd)
			}

			if provider.GetServerURL() != tt.expectedURL {
				t.Errorf("Expected server URL '%s', got '%s'", tt.expectedURL, provider.GetServerURL())
			}

			if provider.IsSandbox() == tt.expectedIsProd {
				t.Error("IsSandbox() returned unexpected value")
			}
		})
	}
}

func TestPolarPaymentProvider_GetResponseError(t *testing.T) {
	provider := &PolarPaymentProvider{}

	// Test with nil error
	if result := provider.GetResponseError(nil); result != "success" {
		t.Errorf("Expected 'success' for nil error, got '%s'", result)
	}

	// Test with non-nil error
	if result := provider.GetResponseError(testError{}); result != "fail" {
		t.Errorf("Expected 'fail' for non-nil error, got '%s'", result)
	}
}

func TestPolarPaymentProvider_Pay(t *testing.T) {
	// Test sandbox provider
	provider, err := NewPolarPaymentProviderWithEnv("test_token", "sandbox")
	if err != nil {
		t.Fatalf("Failed to create sandbox provider: %v", err)
	}

	// Create a test payment request
	payReq := &PayReq{
		ProviderName:       "TestProvider",
		ProductName:        "TestProduct",
		ProductDisplayName: "Test Product Display",
		PaymentName:        "test-payment-123",
		PayerName:          "Test User",
		PayerEmail:         "test@example.com",
		Price:              10.50, // $10.50
		Currency:           "usd",
		ReturnUrl:          "https://example.com/success",
		NotifyUrl:          "https://example.com/notify",
	}

	// Note: This test will fail without a real token, but we can test the structure
	// In a real scenario, you would mock the Polar client
	_, err = provider.Pay(payReq)

	// We expect this to fail with authentication error since we're using a fake token
	if err == nil {
		t.Error("Expected authentication error with fake token")
	}

	// Verify provider is in sandbox mode
	if !provider.IsSandbox() {
		t.Error("Expected provider to be in sandbox mode")
	}

	if provider.GetEnvironment() != "sandbox" {
		t.Errorf("Expected sandbox environment, got '%s'", provider.GetEnvironment())
	}
}

func TestPolarPaymentProvider_ProductionMode(t *testing.T) {
	// Test production provider
	provider, err := NewPolarPaymentProviderWithEnv("test_token", "production")
	if err != nil {
		t.Fatalf("Failed to create production provider: %v", err)
	}

	// Verify provider is in production mode
	if provider.IsSandbox() {
		t.Error("Expected provider to be in production mode")
	}

	if provider.GetEnvironment() != "production" {
		t.Errorf("Expected production environment, got '%s'", provider.GetEnvironment())
	}

	if provider.GetServerURL() != "https://api.polar.sh" {
		t.Errorf("Expected production URL, got '%s'", provider.GetServerURL())
	}
}

type testError struct{}

func (e testError) Error() string {
	return "test error"
}
