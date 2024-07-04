package util

// Constants for all supported currencies
const (
	USD = "USD"
	AZN = "AZN"
	EUR = "EUR"
)

// IsSuportedCurrency returns true if the currencyis supported
func IsSuportedCurrency(currency string) bool {
	switch currency {
	case USD, AZN, EUR:
		return true
	}
	return false
}
