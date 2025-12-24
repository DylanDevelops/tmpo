package currency

import (
	"testing"
)

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name         string
		amount       float64
		currencyCode string
		expected     string
	}{
		// USD tests
		{
			name:         "USD with standard amount",
			amount:       150.00,
			currencyCode: "USD",
			expected:     "$150.00",
		},
		{
			name:         "USD with decimal places",
			amount:       99.99,
			currencyCode: "USD",
			expected:     "$99.99",
		},
		{
			name:         "USD with zero",
			amount:       0.00,
			currencyCode: "USD",
			expected:     "$0.00",
		},
		{
			name:         "USD with large amount",
			amount:       123456.78,
			currencyCode: "USD",
			expected:     "$123456.78",
		},

		// Euro tests
		{
			name:         "EUR with standard amount",
			amount:       100.00,
			currencyCode: "EUR",
			expected:     "€100.00",
		},
		{
			name:         "EUR lowercase",
			amount:       50.50,
			currencyCode: "eur",
			expected:     "€50.50",
		},

		// GBP tests
		{
			name:         "GBP with standard amount",
			amount:       200.00,
			currencyCode: "GBP",
			expected:     "£200.00",
		},

		// Asian currencies
		{
			name:         "JPY with standard amount",
			amount:       10000.00,
			currencyCode: "JPY",
			expected:     "¥10000.00",
		},
		{
			name:         "INR with standard amount",
			amount:       5000.00,
			currencyCode: "INR",
			expected:     "₹5000.00",
		},
		{
			name:         "KRW with standard amount",
			amount:       100000.00,
			currencyCode: "KRW",
			expected:     "₩100000.00",
		},

		// Other currencies
		{
			name:         "CAD with standard amount",
			amount:       75.00,
			currencyCode: "CAD",
			expected:     "CA$75.00",
		},
		{
			name:         "AUD with standard amount",
			amount:       150.00,
			currencyCode: "AUD",
			expected:     "A$150.00",
		},
		{
			name:         "CHF with standard amount",
			amount:       100.00,
			currencyCode: "CHF",
			expected:     "Fr100.00",
		},

		// Edge cases
		{
			name:         "Empty currency code defaults to USD",
			amount:       100.00,
			currencyCode: "",
			expected:     "$100.00",
		},
		{
			name:         "Unknown currency code defaults to USD",
			amount:       100.00,
			currencyCode: "XYZ",
			expected:     "$100.00",
		},
		{
			name:         "Whitespace in currency code",
			amount:       50.00,
			currencyCode: "  USD  ",
			expected:     "$50.00",
		},
		{
			name:         "Mixed case currency code",
			amount:       75.25,
			currencyCode: "GbP",
			expected:     "£75.25",
		},
		{
			name:         "Very small amount",
			amount:       0.01,
			currencyCode: "USD",
			expected:     "$0.01",
		},
		{
			name:         "Amount with many decimal places (should round to 2)",
			amount:       99.999,
			currencyCode: "USD",
			expected:     "$100.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCurrency(tt.amount, tt.currencyCode)
			if result != tt.expected {
				t.Errorf("FormatCurrency(%f, %q) = %q, expected %q",
					tt.amount, tt.currencyCode, result, tt.expected)
			}
		})
	}
}

func TestGetSymbol(t *testing.T) {
	tests := []struct {
		name         string
		currencyCode string
		expected     string
	}{
		{
			name:         "USD returns dollar sign",
			currencyCode: "USD",
			expected:     "$",
		},
		{
			name:         "EUR returns euro sign",
			currencyCode: "EUR",
			expected:     "€",
		},
		{
			name:         "GBP returns pound sign",
			currencyCode: "GBP",
			expected:     "£",
		},
		{
			name:         "JPY returns yen sign",
			currencyCode: "JPY",
			expected:     "¥",
		},
		{
			name:         "Lowercase currency code",
			currencyCode: "usd",
			expected:     "$",
		},
		{
			name:         "Mixed case currency code",
			currencyCode: "Eur",
			expected:     "€",
		},
		{
			name:         "Unknown currency returns code itself",
			currencyCode: "XYZ",
			expected:     "XYZ",
		},
		{
			name:         "Whitespace is trimmed",
			currencyCode: "  GBP  ",
			expected:     "£",
		},
		{
			name:         "Empty string returns empty",
			currencyCode: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSymbol(tt.currencyCode)
			if result != tt.expected {
				t.Errorf("GetSymbol(%q) = %q, expected %q",
					tt.currencyCode, result, tt.expected)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		name         string
		currencyCode string
		expected     bool
	}{
		{
			name:         "USD is supported",
			currencyCode: "USD",
			expected:     true,
		},
		{
			name:         "EUR is supported",
			currencyCode: "EUR",
			expected:     true,
		},
		{
			name:         "GBP is supported",
			currencyCode: "GBP",
			expected:     true,
		},
		{
			name:         "JPY is supported",
			currencyCode: "JPY",
			expected:     true,
		},
		{
			name:         "INR is supported",
			currencyCode: "INR",
			expected:     true,
		},
		{
			name:         "Lowercase USD is supported",
			currencyCode: "usd",
			expected:     true,
		},
		{
			name:         "Unknown currency is not supported",
			currencyCode: "XYZ",
			expected:     false,
		},
		{
			name:         "Empty string is not supported",
			currencyCode: "",
			expected:     false,
		},
		{
			name:         "Whitespace around supported code",
			currencyCode: "  EUR  ",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSupported(tt.currencyCode)
			if result != tt.expected {
				t.Errorf("IsSupported(%q) = %v, expected %v",
					tt.currencyCode, result, tt.expected)
			}
		})
	}
}

func TestGetSupportedCurrencies(t *testing.T) {
	currencies := GetSupportedCurrencies()

	// Check that we have a reasonable number of currencies
	if len(currencies) < 20 {
		t.Errorf("GetSupportedCurrencies() returned %d currencies, expected at least 20",
			len(currencies))
	}

	// Check that common currencies are included
	commonCurrencies := []string{"USD", "EUR", "GBP", "JPY", "CNY", "INR", "CAD", "AUD"}
	for _, code := range commonCurrencies {
		found := false
		for _, c := range currencies {
			if c == code {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedCurrencies() missing expected currency: %s", code)
		}
	}
}

func TestDefaultCurrency(t *testing.T) {
	if DefaultCurrency != "USD" {
		t.Errorf("DefaultCurrency = %q, expected %q", DefaultCurrency, "USD")
	}
}
