package pdf

import (
	"fmt"
	"strings"

	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	invoicesdb "github.com/your-org/invoice-backend/internal/domain/invoices/sqlc"
	settingsdb "github.com/your-org/invoice-backend/internal/domain/settings/sqlc"
)

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&#34;")
	return s
}

// buildInvoiceHTML renders invoice data into a printable HTML page.
// If customTemplate is non-empty it is used as the base HTML with {{placeholder}} substitution.
// Otherwise the built-in indigo-themed template is used.
func buildInvoiceHTML(
	inv invoicesdb.Invoice,
	items []invoicesdb.InvoiceItem,
	customer customersdb.Customer,
	settings settingsdb.Setting,
	customTemplate string,
) string {
	issuedDate := ""
	if inv.IssuedDate.Valid {
		issuedDate = inv.IssuedDate.Time.Format("02 Jan 2006")
	}
	dueDate := ""
	if inv.DueDate.Valid {
		dueDate = inv.DueDate.Time.Format("02 Jan 2006")
	}

	itemsTable := buildItemsTable(items, inv.IsSameState, inv.Currency)

	if customTemplate != "" {
		return applyCustomTemplate(customTemplate, inv, customer, settings, issuedDate, dueDate, itemsTable)
	}
	return buildDefaultHTML(inv, customer, settings, issuedDate, dueDate, itemsTable)
}

func buildItemsTable(items []invoicesdb.InvoiceItem, isSameState bool, currency string) string {
	var sb strings.Builder

	sb.WriteString(`<table class="items-table"><thead><tr>`)
	sb.WriteString(`<th style="text-align:left">Description</th><th>HSN/SAC</th><th>Qty</th><th>Unit Price</th>`)
	if isSameState {
		sb.WriteString(`<th>CGST</th><th>SGST</th>`)
	} else {
		sb.WriteString(`<th>IGST</th>`)
	}
	sb.WriteString(`<th>Amount</th></tr></thead><tbody>`)

	for _, item := range items {
		sb.WriteString("<tr>")
		sb.WriteString(fmt.Sprintf(`<td>%s</td>`, escapeHTML(item.Description)))
		sb.WriteString(fmt.Sprintf(`<td>%s</td>`, escapeHTML(safeStr(item.HsnSacCode))))
		sb.WriteString(fmt.Sprintf(`<td>%.2f</td>`, item.Quantity))
		sb.WriteString(fmt.Sprintf(`<td>%s %.2f</td>`, currency, item.UnitPrice))
		if isSameState {
			half := item.TaxRate / 2
			sb.WriteString(fmt.Sprintf(`<td>%.1f%%</td><td>%.1f%%</td>`, half, half))
		} else {
			sb.WriteString(fmt.Sprintf(`<td>%.1f%%</td>`, item.TaxRate))
		}
		sb.WriteString(fmt.Sprintf(`<td>%s %.2f</td></tr>`, currency, item.LineTotal))
	}
	sb.WriteString("</tbody></table>")
	return sb.String()
}

func applyCustomTemplate(
	tmpl string,
	inv invoicesdb.Invoice,
	customer customersdb.Customer,
	settings settingsdb.Setting,
	issuedDate, dueDate, itemsTable string,
) string {
	r := strings.NewReplacer(
		"{{business_name}}", escapeHTML(safeStr(settings.BusinessName)),
		"{{business_email}}", escapeHTML(safeStr(settings.BusinessEmail)),
		"{{business_phone}}", escapeHTML(safeStr(settings.BusinessPhone)),
		"{{business_address}}", escapeHTML(safeStr(settings.BusinessAddress)),
		"{{logo_url}}", safeStr(settings.LogoUrl),
		"{{invoice_number}}", escapeHTML(inv.InvoiceNumber),
		"{{status}}", strings.ToUpper(inv.Status),
		"{{issued_date}}", issuedDate,
		"{{due_date}}", dueDate,
		"{{currency}}", inv.Currency,
		"{{customer_name}}", escapeHTML(customer.Name),
		"{{customer_email}}", escapeHTML(safeStr(customer.Email)),
		"{{customer_phone}}", escapeHTML(safeStr(customer.Phone)),
		"{{customer_address}}", escapeHTML(safeStr(customer.Address)),
		"{{customer_gstin}}", escapeHTML(safeStr(customer.TaxNumber)),
		"{{items_table}}", itemsTable,
		"{{subtotal}}", fmt.Sprintf("%.2f", inv.Subtotal),
		"{{tax_amount}}", fmt.Sprintf("%.2f", inv.TaxAmount),
		"{{cgst_amount}}", fmt.Sprintf("%.2f", inv.CgstAmount),
		"{{sgst_amount}}", fmt.Sprintf("%.2f", inv.SgstAmount),
		"{{igst_amount}}", fmt.Sprintf("%.2f", inv.IgstAmount),
		"{{discount_amount}}", fmt.Sprintf("%.2f", inv.DiscountAmount),
		"{{total}}", fmt.Sprintf("%.2f", inv.Total),
		"{{grand_total}}", fmt.Sprintf("%.2f", inv.GrandTotal),
		"{{notes}}", escapeHTML(safeStr(inv.Notes)),
		"{{terms}}", escapeHTML(safeStr(inv.Terms)),
	)
	html := r.Replace(tmpl)

	// Inject auto-print script before </head>
	printScript := `<script>window.onload=function(){setTimeout(function(){window.print();},400);}</script>`
	if strings.Contains(html, "</head>") {
		return strings.Replace(html, "</head>", printScript+"</head>", 1)
	}
	return printScript + html
}

func buildDefaultHTML(
	inv invoicesdb.Invoice,
	customer customersdb.Customer,
	settings settingsdb.Setting,
	issuedDate, dueDate, itemsTable string,
) string {
	bizName := safeStr(settings.BusinessName)
	if bizName == "" {
		bizName = "Your Business"
	}

	logoBlock := ""
	if u := safeStr(settings.LogoUrl); u != "" {
		logoBlock = fmt.Sprintf(`<img src="%s" alt="Logo" class="logo" />`, u)
	}

	gstRows := ""
	if inv.IsSameState {
		gstRows = fmt.Sprintf(
			`<div class="t-row"><span>CGST</span><span>%s %.2f</span></div><div class="t-row"><span>SGST</span><span>%s %.2f</span></div>`,
			inv.Currency, inv.CgstAmount, inv.Currency, inv.SgstAmount,
		)
	} else {
		gstRows = fmt.Sprintf(
			`<div class="t-row"><span>IGST</span><span>%s %.2f</span></div>`,
			inv.Currency, inv.IgstAmount,
		)
	}

	discountRow := ""
	if inv.DiscountAmount > 0 {
		discountRow = fmt.Sprintf(
			`<div class="t-row"><span>Discount</span><span>- %s %.2f</span></div>`,
			inv.Currency, inv.DiscountAmount,
		)
	}

	gstinLine := ""
	if customer.TaxNumber != nil && *customer.TaxNumber != "" {
		gstinLine = fmt.Sprintf(`<div>GSTIN: %s</div>`, escapeHTML(*customer.TaxNumber))
	}

	notesBlock := ""
	if inv.Notes != nil && *inv.Notes != "" {
		notesBlock = fmt.Sprintf(
			`<div class="extra"><div class="lbl">Notes</div><p>%s</p></div>`,
			escapeHTML(*inv.Notes),
		)
	}
	termsBlock := ""
	if inv.Terms != nil && *inv.Terms != "" {
		termsBlock = fmt.Sprintf(
			`<div class="extra"><div class="lbl">Terms &amp; Conditions</div><p>%s</p></div>`,
			escapeHTML(*inv.Terms),
		)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>Invoice %s</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:'Helvetica Neue',Arial,sans-serif;color:#1f2937;background:#fff;font-size:13px}
.wrap{max-width:820px;margin:0 auto;padding:40px}
.header{display:flex;justify-content:space-between;align-items:flex-start;padding-bottom:24px;border-bottom:3px solid #1e40af;margin-bottom:28px}
.inv-title{font-size:34px;font-weight:800;color:#1e40af;line-height:1}
.inv-num{font-size:13px;color:#6b7280;margin-top:6px}
.badge{display:inline-block;margin-top:8px;padding:3px 12px;border-radius:20px;font-size:11px;font-weight:700;text-transform:uppercase;background:#dbeafe;color:#1e40af}
.logo{max-height:60px;max-width:180px;object-fit:contain;display:block;margin-left:auto;margin-bottom:8px}
.biz-name{font-size:16px;font-weight:700;color:#1e40af;text-align:right}
.biz-info{font-size:12px;color:#6b7280;text-align:right;line-height:1.6;margin-top:4px}
.meta{display:flex;gap:40px;margin-bottom:28px}
.bill-to{flex:1}
.lbl{font-size:11px;font-weight:700;color:#6366f1;text-transform:uppercase;letter-spacing:.5px;margin-bottom:6px}
.cust-name{font-size:15px;font-weight:700;margin-bottom:4px}
.cust-detail{font-size:12px;color:#6b7280;line-height:1.6}
.dates{text-align:right}
.date-item{font-size:12px;color:#6b7280;margin-bottom:4px}
.date-item strong{color:#1f2937}
.items-table{width:100%%;border-collapse:collapse;margin-bottom:24px}
.items-table th{background:#eef2ff;color:#1e40af;font-size:11px;font-weight:700;padding:10px 12px;text-align:right;text-transform:uppercase;letter-spacing:.3px}
.items-table th:first-child{text-align:left}
.items-table td{padding:9px 12px;border-bottom:1px solid #e5e7eb;text-align:right}
.items-table td:first-child{text-align:left}
.items-table tr:nth-child(even) td{background:#f9fafb}
.totals-wrap{display:flex;justify-content:flex-end;margin-bottom:28px}
.totals-box{min-width:270px}
.t-row{display:flex;justify-content:space-between;padding:6px 0;font-size:13px;color:#6b7280;border-bottom:1px solid #f3f4f6}
.grand{background:#1e40af;color:#fff!important;padding:12px 14px;border-radius:8px;margin-top:8px;font-size:15px;font-weight:700;border-bottom:none}
.grand span{color:#fff!important}
.extra{margin-bottom:20px}
.extra p{font-size:12px;color:#6b7280;line-height:1.6;margin-top:4px}
.footer{margin-top:36px;padding-top:14px;border-top:1px solid #e5e7eb;text-align:center;font-size:11px;color:#9ca3af}
@media print{body{background:#fff}.wrap{max-width:100%%;padding:20px}@page{margin:10mm;size:A4}}
</style>
<script>window.onload=function(){setTimeout(function(){window.print();},400);}</script>
</head>
<body>
<div class="wrap">
  <div class="header">
    <div>
      <div class="inv-title">INVOICE</div>
      <div class="inv-num"># %s</div>
      <div><span class="badge">%s</span></div>
    </div>
    <div>
      %s
      <div class="biz-name">%s</div>
      <div class="biz-info">%s<br/>%s<br/>%s</div>
    </div>
  </div>
  <div class="meta">
    <div class="bill-to">
      <div class="lbl">Bill To</div>
      <div class="cust-name">%s</div>
      <div class="cust-detail">%s<br/>%s<br/>%s%s</div>
    </div>
    <div class="dates">
      <div class="date-item"><strong>Issue Date:</strong> %s</div>
      <div class="date-item"><strong>Due Date:</strong> %s</div>
      <div class="date-item"><strong>Currency:</strong> %s</div>
    </div>
  </div>
  %s
  <div class="totals-wrap">
    <div class="totals-box">
      <div class="t-row"><span>Subtotal</span><span>%s %.2f</span></div>
      %s
      %s
      <div class="t-row" style="font-weight:600;color:#1f2937"><span>Tax Amount</span><span>%s %.2f</span></div>
      <div class="grand t-row"><span>GRAND TOTAL</span><span>%s %.2f</span></div>
    </div>
  </div>
  %s%s
  <div class="footer">Thank you for your business.</div>
</div>
</body>
</html>`,
		inv.InvoiceNumber,
		inv.InvoiceNumber,
		strings.ToUpper(inv.Status),
		logoBlock,
		escapeHTML(bizName),
		escapeHTML(safeStr(settings.BusinessAddress)),
		escapeHTML(safeStr(settings.BusinessPhone)),
		escapeHTML(safeStr(settings.BusinessEmail)),
		escapeHTML(customer.Name),
		escapeHTML(safeStr(customer.Email)),
		escapeHTML(safeStr(customer.Phone)),
		escapeHTML(safeStr(customer.Address)),
		gstinLine,
		issuedDate,
		dueDate,
		inv.Currency,
		itemsTable,
		inv.Currency, inv.Subtotal,
		discountRow,
		gstRows,
		inv.Currency, inv.TaxAmount,
		inv.Currency, inv.GrandTotal,
		notesBlock,
		termsBlock,
	)
}
