package utils

import "fmt"

func FormatInvoiceNumber(prefix string, year int, sequence int) string {
	return fmt.Sprintf("%s-%d-%04d", prefix, year, sequence)
}
