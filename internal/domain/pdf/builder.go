package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
)

// colour palette
const (
	headerR, headerG, headerB = 30, 64, 175   // indigo-800
	accentR, accentG, accentB = 99, 102, 241  // indigo-500
	lightR, lightG, lightB    = 238, 242, 255 // indigo-50
	textR, textG, textB       = 31, 41, 55    // gray-800
	mutedR, mutedG, mutedB    = 107, 114, 128 // gray-500
)

func numericStr(n interface{ Int64Value() (int64, bool) }) string {
	// pgtype.Numeric → string via fmt
	return fmt.Sprintf("%v", n)
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func buildInvoicePDF(
	inv invoicesdb.Invoice,
	items []invoicesdb.InvoiceItem,
	customer customersdb.Customer,
	settings settingsdb.Setting,
) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pageW, _ := pdf.GetPageSize()
	contentW := pageW - 30 // left+right margins

	// ── Header bar ──────────────────────────────────────────────────────────
	pdf.SetFillColor(headerR, headerG, headerB)
	pdf.Rect(0, 0, pageW, 28, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetXY(15, 7)
	pdf.CellFormat(contentW/2, 10, "INVOICE", "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetXY(pageW/2, 7)
	pdf.CellFormat(contentW/2-15, 5, fmt.Sprintf("# %s", inv.InvoiceNumber), "", 0, "R", false, 0, "")
	pdf.SetXY(pageW/2, 13)
	pdf.CellFormat(contentW/2-15, 5,
		fmt.Sprintf("Status: %s", strings.ToUpper(inv.Status)), "", 0, "R", false, 0, "")
	pdf.SetXY(pageW/2, 19)
	pdf.CellFormat(contentW/2-15, 5,
		fmt.Sprintf("Currency: %s", inv.Currency), "", 0, "R", false, 0, "")

	// ── Business info (left) & Invoice dates (right) ─────────────────────
	pdf.SetTextColor(textR, textG, textB)
	pdf.SetY(35)

	// Business block
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetX(15)
	bizName := safeStr(settings.BusinessName)
	if bizName == "" {
		bizName = "Your Business"
	}
	pdf.CellFormat(contentW/2, 6, bizName, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	for _, line := range []string{
		safeStr(settings.BusinessAddress),
		safeStr(settings.BusinessPhone),
		safeStr(settings.BusinessEmail),
	} {
		if line != "" {
			pdf.SetX(15)
			pdf.CellFormat(contentW/2, 5, line, "", 1, "L", false, 0, "")
		}
	}

	// Dates block (right column, same Y start)
	pdf.SetXY(pageW/2, 35)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(textR, textG, textB)
	dateRows := [][]string{
		{"Issue Date:", inv.IssuedDate.Time.Format("02 Jan 2006")},
		{"Due Date:", inv.DueDate.Time.Format("02 Jan 2006")},
	}
	for _, row := range dateRows {
		pdf.SetX(pageW / 2)
		pdf.CellFormat(30, 5, row[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(contentW/2-15, 5, row[1], "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
	}

	// ── Divider ──────────────────────────────────────────────────────────────
	pdf.SetDrawColor(accentR, accentG, accentB)
	pdf.SetLineWidth(0.4)
	pdf.Line(15, pdf.GetY()+4, pageW-15, pdf.GetY()+4)
	pdf.SetY(pdf.GetY() + 8)

	// ── Bill To ──────────────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(accentR, accentG, accentB)
	pdf.SetX(15)
	pdf.CellFormat(contentW, 5, "BILL TO", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(textR, textG, textB)
	pdf.SetX(15)
	pdf.CellFormat(contentW, 5, customer.Name, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	for _, line := range []string{
		safeStr(customer.Email),
		safeStr(customer.Phone),
		safeStr(customer.Address),
		safeStr(customer.TaxNumber),
	} {
		if line != "" {
			pdf.SetX(15)
			pdf.CellFormat(contentW, 5, line, "", 1, "L", false, 0, "")
		}
	}

	pdf.SetY(pdf.GetY() + 6)

	// ── Items table header ────────────────────────────────────────────────────
	colW := [5]float64{70, 20, 30, 25, 30}
	headers := [5]string{"Description", "Qty", "Unit Price", "Tax %", "Total"}

	pdf.SetFillColor(lightR, lightG, lightB)
	pdf.SetTextColor(headerR, headerG, headerB)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetX(15)
	for i, h := range headers {
		align := "L"
		if i > 0 {
			align = "R"
		}
		pdf.CellFormat(colW[i], 7, h, "TB", 0, align, true, 0, "")
	}
	pdf.Ln(-1)

	// ── Items rows ────────────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(textR, textG, textB)
	fill := false
	for _, item := range items {
		if fill {
			pdf.SetFillColor(248, 249, 255)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetX(15)

		qty := fmt.Sprintf("%v", item.Quantity)
		unitPrice := fmt.Sprintf("%v", item.UnitPrice)
		taxRate := fmt.Sprintf("%v%%", item.TaxRate)
		lineTotal := fmt.Sprintf("%v", item.LineTotal)

		pdf.CellFormat(colW[0], 6, item.Description, "", 0, "L", true, 0, "")
		pdf.CellFormat(colW[1], 6, qty, "", 0, "R", true, 0, "")
		pdf.CellFormat(colW[2], 6, unitPrice, "", 0, "R", true, 0, "")
		pdf.CellFormat(colW[3], 6, taxRate, "", 0, "R", true, 0, "")
		pdf.CellFormat(colW[4], 6, lineTotal, "", 0, "R", true, 0, "")
		pdf.Ln(-1)
		fill = !fill
	}

	// ── Totals block ─────────────────────────────────────────────────────────
	pdf.SetY(pdf.GetY() + 4)
	totalsX := pageW - 15 - 70.0
	labelW := 40.0
	valW := 30.0

	totalsRows := [][]string{
		{"Subtotal:", fmt.Sprintf("%v", inv.Subtotal)},
		{"Tax Amount:", fmt.Sprintf("%v", inv.TaxAmount)},
		{"Discount:", fmt.Sprintf("%v", inv.DiscountAmount)},
	}

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	for _, row := range totalsRows {
		pdf.SetX(totalsX)
		pdf.CellFormat(labelW, 5, row[0], "", 0, "L", false, 0, "")
		pdf.CellFormat(valW, 5, row[1], "", 1, "R", false, 0, "")
	}

	// Total line
	pdf.SetDrawColor(accentR, accentG, accentB)
	pdf.Line(totalsX, pdf.GetY()+1, pageW-15, pdf.GetY()+1)
	pdf.SetY(pdf.GetY() + 3)
	pdf.SetX(totalsX)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(headerR, headerG, headerB)
	pdf.CellFormat(labelW, 7, "TOTAL:", "", 0, "L", false, 0, "")
	pdf.CellFormat(valW, 7,
		fmt.Sprintf("%s %v", inv.Currency, inv.Total), "", 1, "R", false, 0, "")

	// ── Notes & Terms ─────────────────────────────────────────────────────────
	if inv.Notes != nil && *inv.Notes != "" {
		pdf.SetY(pdf.GetY() + 8)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(accentR, accentG, accentB)
		pdf.SetX(15)
		pdf.CellFormat(contentW, 5, "NOTES", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(textR, textG, textB)
		pdf.SetX(15)
		pdf.MultiCell(contentW, 5, *inv.Notes, "", "L", false)
	}

	if inv.Terms != nil && *inv.Terms != "" {
		pdf.SetY(pdf.GetY() + 4)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(accentR, accentG, accentB)
		pdf.SetX(15)
		pdf.CellFormat(contentW, 5, "TERMS & CONDITIONS", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(textR, textG, textB)
		pdf.SetX(15)
		pdf.MultiCell(contentW, 5, *inv.Terms, "", "L", false)
	}

	// ── Footer ────────────────────────────────────────────────────────────────
	_, pageH := pdf.GetPageSize()
	pdf.SetY(pageH - 18)
	pdf.SetDrawColor(lightR, lightG, lightB)
	pdf.SetLineWidth(0.3)
	pdf.Line(15, pdf.GetY(), pageW-15, pdf.GetY())
	pdf.SetY(pdf.GetY() + 2)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(mutedR, mutedG, mutedB)
	pdf.SetX(15)
	pdf.CellFormat(contentW, 5, "Thank you for your business.", "", 0, "C", false, 0, "")

	// ── Output to bytes ───────────────────────────────────────────────────────
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
