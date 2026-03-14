package constants

// Invoice statuses
const (
	InvoiceStatusDraft     = "draft"
	InvoiceStatusSent      = "sent"
	InvoiceStatusViewed    = "viewed"
	InvoiceStatusPaid      = "paid"
	InvoiceStatusOverdue   = "overdue"
	InvoiceStatusCancelled = "cancelled"
)

// Payment methods
const (
	PaymentCash         = "cash"
	PaymentBankTransfer = "bank_transfer"
	PaymentCard         = "card"
	PaymentCheque       = "cheque"
	PaymentOnline       = "online"
	PaymentOther        = "other"
)

// Subscription statuses
const (
	SubActive    = "active"
	SubTrialing  = "trialing"
	SubPastDue   = "past_due"
	SubCancelled = "cancelled"
	SubSuspended = "suspended"
)

// Organisation statuses
const (
	OrgActive    = "active"
	OrgSuspended = "suspended"
	OrgCancelled = "cancelled"
	OrgTrial     = "trial"
)

// WhatsApp log statuses
const (
	WAPending   = "pending"
	WASent      = "sent"
	WADelivered = "delivered"
	WAFailed    = "failed"
)
