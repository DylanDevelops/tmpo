package currency

import (
	"fmt"
	"strings"
)

// DefaultCurrency is the currency code used when none is specified in configuration.
const DefaultCurrency = "USD"

// currencySymbols maps ISO 4217 currency codes to their display symbols.
// This map includes the most commonly used global currencies.
var currencySymbols = map[string]string{
	// Americas
	"USD": "$",   // United States Dollar
	"CAD": "CA$", // Canadian Dollar
	"BRL": "R$",  // Brazilian Real
	"MXN": "MX$", // Mexican Peso
	"ARS": "AR$", // Argentine Peso

	// Europe
	"EUR": "€",   // Euro
	"GBP": "£",   // British Pound Sterling
	"CHF": "Fr",  // Swiss Franc
	"SEK": "kr",  // Swedish Krona
	"NOK": "kr",  // Norwegian Krone
	"DKK": "kr",  // Danish Krone
	"PLN": "zł",  // Polish Zloty
	"CZK": "Kč",  // Czech Koruna

	// Asia
	"JPY": "¥",   // Japanese Yen
	"CNY": "¥",   // Chinese Yuan
	"INR": "₹",   // Indian Rupee
	"KRW": "₩",   // South Korean Won
	"SGD": "S$",  // Singapore Dollar
	"HKD": "HK$", // Hong Kong Dollar
	"THB": "฿",   // Thai Baht
	"IDR": "Rp",  // Indonesian Rupiah
	"MYR": "RM",  // Malaysian Ringgit
	"PHP": "₱",   // Philippine Peso
	"VND": "₫",   // Vietnamese Dong

	// Oceania
	"AUD": "A$",  // Australian Dollar
	"NZD": "NZ$", // New Zealand Dollar

	// Middle East & Africa
	"AED": "د.إ", // UAE Dirham
	"SAR": "﷼",   // Saudi Riyal
	"ILS": "₪",   // Israeli Shekel
	"ZAR": "R",   // South African Rand
	"EGP": "E£",  // Egyptian Pound
	"TRY": "₺",   // Turkish Lira
}

// FormatCurrency formats an amount with the appropriate currency symbol.
// The currencyCode is normalized to uppercase and looked up in the symbol map.
// If the currency code is empty or unknown, it defaults to USD ($).
//
// Examples:
//   FormatCurrency(150.00, "USD") returns "$150.00"
//   FormatCurrency(99.99, "EUR") returns "€99.99"
//   FormatCurrency(1234.56, "GBP") returns "£1234.56"
//   FormatCurrency(100.00, "") returns "$100.00"
//   FormatCurrency(100.00, "UNKNOWN") returns "$100.00"
func FormatCurrency(amount float64, currencyCode string) string {
	// Normalize currency code to uppercase
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))

	// Default to USD if empty or unknown
	if currencyCode == "" || !IsSupported(currencyCode) {
		currencyCode = DefaultCurrency
	}

	symbol := GetSymbol(currencyCode)
	return fmt.Sprintf("%s%.2f", symbol, amount)
}

// GetSymbol returns the display symbol for the given currency code.
// The currency code is normalized to uppercase before lookup.
// If the currency code is not found, it returns the code itself.
//
// Examples:
//   GetSymbol("USD") returns "$"
//   GetSymbol("eur") returns "€"
//   GetSymbol("UNKNOWN") returns "UNKNOWN"
func GetSymbol(currencyCode string) string {
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))

	if symbol, exists := currencySymbols[currencyCode]; exists {
		return symbol
	}

	return currencyCode
}

// IsSupported returns true if the given currency code is in the supported currencies map.
// The currency code is normalized to uppercase before checking.
func IsSupported(currencyCode string) bool {
	currencyCode = strings.ToUpper(strings.TrimSpace(currencyCode))
	_, exists := currencySymbols[currencyCode]
	return exists
}

// GetSupportedCurrencies returns a sorted list of all supported currency codes.
// This is useful for documentation or validation purposes.
func GetSupportedCurrencies() []string {
	currencies := make([]string, 0, len(currencySymbols))
	for code := range currencySymbols {
		currencies = append(currencies, code)
	}
	return currencies
}
