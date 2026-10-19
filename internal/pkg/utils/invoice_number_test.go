package utils

import "testing"

func TestFormatInvoiceNumber(t *testing.T) {
	got := FormatInvoiceNumber("INV", 2026, 7)
	if got != "INV-2026-0007" {
		t.Fatalf("got %s", got)
	}
}
