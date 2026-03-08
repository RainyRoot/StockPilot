package money

import "fmt"

// CentsToFloat converts cents (int64) to float for display only.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}

// FloatToCents converts a float price to cents.
func FloatToCents(price float64) int64 {
	return int64(price * 100)
}

// Format formats cents as a currency string (e.g., "$123.45" or "123,45 EUR").
func Format(cents int64, currency string) string {
	whole := cents / 100
	frac := cents % 100
	if frac < 0 {
		frac = -frac
	}
	switch currency {
	case "EUR":
		return fmt.Sprintf("%d,%02d EUR", whole, frac)
	default:
		return fmt.Sprintf("$%d.%02d", whole, frac)
	}
}
