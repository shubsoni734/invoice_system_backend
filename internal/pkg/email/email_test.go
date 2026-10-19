package email

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/invoice-backend/internal/config"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient(config.EmailConfig{
		APIKey:      "test-key",
		APIURL:      server.URL,
		SenderEmail: "billing@example.com",
		SenderName:  "Acme Books",
	})
}

func TestSendWelcomeEmailEscapesHTMLAndUsesBrevoKey(t *testing.T) {
	var message brevoMessage
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("api-key") != "test-key" {
			t.Errorf("api-key = %q", r.Header.Get("api-key"))
		}
		if r.Header.Get("content-type") != "application/json" {
			t.Errorf("content-type = %q", r.Header.Get("content-type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
	})

	err := client.SendWelcomeEmail("owner@example.com", `Acme <script>`, `p@ss"word`)
	if err != nil {
		t.Fatal(err)
	}
	if message.Sender.Email != "billing@example.com" || message.Sender.Name != "Acme Books" {
		t.Fatalf("sender = %+v", message.Sender)
	}
	if len(message.To) != 1 || message.To[0].Email != "owner@example.com" {
		t.Fatalf("to = %+v", message.To)
	}
	if strings.Contains(message.HTMLContent, "<script>") {
		t.Fatal("organisation name was not escaped")
	}
	if !strings.Contains(message.HTMLContent, "Acme &lt;script&gt;") {
		t.Fatalf("escaped name missing from body: %s", message.HTMLContent)
	}
	if !strings.Contains(message.HTMLContent, "p@ss&#34;word") {
		t.Fatalf("escaped password missing from body: %s", message.HTMLContent)
	}
}

func TestSendForgotPasswordEmailIncludesResetLink(t *testing.T) {
	var message brevoMessage
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&message)
		w.WriteHeader(http.StatusCreated)
	})

	err := client.SendForgotPasswordEmail("owner@example.com", "reset-token", "https://app.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(message.HTMLContent, "https://app.example.com/new-password?refreshtoken=reset-token") {
		t.Fatalf("reset link missing: %s", message.HTMLContent)
	}
}

func TestSendInvoiceEmailIncludesLinesAndReplyTo(t *testing.T) {
	var message brevoMessage
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&message)
		w.WriteHeader(http.StatusCreated)
	})

	err := client.SendInvoiceEmail("client@example.com", InvoiceEmailData{
		BusinessName:  "Acme & Co",
		BusinessEmail: "accounts@acme.test",
		InvoiceNumber: "INV-2026-0007",
		IssuedDate:    "01 Jan 2026",
		DueDate:       "15 Jan 2026",
		Currency:      "INR",
		CustomerName:  "Ravi <b>Kumar</b>",
		Subtotal:      "100.00",
		TaxAmount:     "18.00",
		GrandTotal:    "118.00",
		Notes:         "Pay <now>",
		Lines: []InvoiceLine{{
			Description: "Design <work>",
			Quantity:    "1.00",
			UnitPrice:   "INR 100.00",
			Amount:      "INR 100.00",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if message.Subject != "Invoice INV-2026-0007 from Acme & Co" {
		t.Fatalf("subject = %s", message.Subject)
	}
	if message.ReplyTo == nil || message.ReplyTo.Email != "accounts@acme.test" {
		t.Fatalf("replyTo = %+v", message.ReplyTo)
	}
	for _, raw := range []string{"<b>Kumar</b>", "Design <work>", "Pay <now>"} {
		if strings.Contains(message.HTMLContent, raw) {
			t.Fatalf("unescaped content %q in %s", raw, message.HTMLContent)
		}
	}
	if !strings.Contains(message.HTMLContent, "INV-2026-0007") || !strings.Contains(message.HTMLContent, "118.00") {
		t.Fatalf("invoice totals missing: %s", message.HTMLContent)
	}
}

func TestSendPaymentEmailIncludesReceipt(t *testing.T) {
	var message brevoMessage
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&message)
		w.WriteHeader(http.StatusCreated)
	})

	err := client.SendPaymentEmail("client@example.com", PaymentEmailData{
		BusinessName:  "Acme & Co",
		BusinessEmail: "accounts@acme.test",
		CustomerName:  "Ravi <b>Kumar</b>",
		InvoiceNumber: "INV-2026-0001",
		Currency:      "INR",
		Amount:        "800.00",
		Method:        "UPI",
		PaymentDate:   "20 Oct 2026",
		Reference:     "UTR <1>",
		BalanceDue:    "0.00",
		FullyPaid:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if message.Subject != "Payment received for INV-2026-0001" {
		t.Fatalf("subject = %s", message.Subject)
	}
	if message.To[0].Email != "client@example.com" {
		t.Fatalf("to = %+v", message.To)
	}
	if strings.Contains(message.HTMLContent, "<b>Kumar</b>") || strings.Contains(message.HTMLContent, "UTR <1>") {
		t.Fatalf("unescaped content in %s", message.HTMLContent)
	}
	if !strings.Contains(message.HTMLContent, "800.00") || !strings.Contains(message.HTMLContent, "fully paid") {
		t.Fatalf("receipt details missing: %s", message.HTMLContent)
	}
}

func TestSendRejectsMissingBrevoConfiguration(t *testing.T) {
	client := NewClient(config.EmailConfig{})
	if client.Configured() {
		t.Fatal("empty client should not be configured")
	}
	err := client.SendTestEmail("owner@example.com")
	if err == nil || !strings.Contains(err.Error(), "BREVO_API_KEY") {
		t.Fatalf("err = %v", err)
	}
}

func TestSendRejectsInvalidRecipient(t *testing.T) {
	client := NewClient(config.EmailConfig{APIKey: "key", SenderEmail: "billing@example.com"})
	if err := client.SendTestEmail("not-an-email"); err == nil {
		t.Fatal("expected invalid recipient error")
	}
}

func TestSendReturnsBrevoAPIError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"message":"Key not found"}`)
	})
	err := client.SendTestEmail("owner@example.com")
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Key not found") {
		t.Fatalf("err = %v", err)
	}
}

func TestSenderNameDefaultsWhenBlank(t *testing.T) {
	var message brevoMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&message)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	client := NewClient(config.EmailConfig{
		APIKey:      "key",
		APIURL:      server.URL,
		SenderEmail: "billing@example.com",
	})
	if err := client.SendTestEmail("owner@example.com"); err != nil {
		t.Fatal(err)
	}
	if message.Sender.Name != "Invoice Flow" {
		t.Fatalf("sender name = %s", message.Sender.Name)
	}
}
