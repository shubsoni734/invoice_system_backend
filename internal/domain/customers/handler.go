package customers

import (
	"context"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	customersdb "github.com/your-org/invoice-backend/internal/domain/customers/sqlc"
	"github.com/your-org/invoice-backend/internal/domain/notify"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

var (
	gstinRegex = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)
	panRegex   = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)
	digitsOnly = regexp.MustCompile(`^\d+$`)
)

var validGSTTreatments = map[string]bool{
	"Registered Business - Regular":     true,
	"Registered Business - Composition": true,
	"Unregistered Business":             true,
	"Consumer":                          true,
	"Overseas":                          true,
	"Special Economic Zone (SEZ)":       true,
	"Deemed Export":                     true,
}

var gstTreatmentsRequiringTaxNumber = map[string]bool{
	"Registered Business - Regular":     true,
	"Registered Business - Composition": true,
	"Special Economic Zone (SEZ)":       true,
	"Deemed Export":                     true,
}

type Handler struct {
	q *customersdb.Queries
}

func NewHandler(q *customersdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateCustomerRequest struct {
	CustomerType        string  `json:"customer_type"`
	Name                string  `json:"name"`
	CompanyName         *string `json:"company_name"`
	Email               *string `json:"email"`
	Phone               *string `json:"phone"`
	GSTTreatment        *string `json:"gst_treatment"`
	TaxNumber           *string `json:"tax_number"`
	BusinessLegalName   *string `json:"business_legal_name"`
	BusinessTradeName   *string `json:"business_trade_name"`
	PlaceOfSupply       *string `json:"place_of_supply"`
	PAN                 *string `json:"pan"`
	TaxPreference       *string `json:"tax_preference"`
	ExemptionReason     *string `json:"exemption_reason"`
	GSTTds              *bool   `json:"gst_tds"`
	PaymentTerm         *string `json:"payment_term"`
	Document            *string `json:"document"`
	Address             *string `json:"address"`
	ContactPersonName   *string `json:"contact_person_name"`
	ContactPersonEmail  *string `json:"contact_person_email"`
	ContactPersonMobile *string `json:"contact_person_mobile"`
	Remark              *string `json:"remark"`
	IsActive            *bool   `json:"is_active"`
}

type UpdateCustomerRequest struct {
	CustomerType        *string `json:"customer_type"`
	Name                *string `json:"name"`
	CompanyName         *string `json:"company_name"`
	Email               *string `json:"email"`
	Phone               *string `json:"phone"`
	GSTTreatment        *string `json:"gst_treatment"`
	TaxNumber           *string `json:"tax_number"`
	BusinessLegalName   *string `json:"business_legal_name"`
	BusinessTradeName   *string `json:"business_trade_name"`
	PlaceOfSupply       *string `json:"place_of_supply"`
	PAN                 *string `json:"pan"`
	TaxPreference       *string `json:"tax_preference"`
	ExemptionReason     *string `json:"exemption_reason"`
	GSTTds              *bool   `json:"gst_tds"`
	PaymentTerm         *string `json:"payment_term"`
	Document            *string `json:"document"`
	Address             *string `json:"address"`
	ContactPersonName   *string `json:"contact_person_name"`
	ContactPersonEmail  *string `json:"contact_person_email"`
	ContactPersonMobile *string `json:"contact_person_mobile"`
	Remark              *string `json:"remark"`
	IsActive            *bool   `json:"is_active"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	return uuid.Parse(orgIDRaw)
}

func isValidEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func countDigits(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n++
		}
	}
	return n
}

func validateGSTIN(taxNumber string) bool {
	return gstinRegex.MatchString(taxNumber)
}

func validateCreateCustomer(req *CreateCustomerRequest) string {
	if strings.TrimSpace(req.Name) == "" {
		return "name is required"
	}

	ct := strings.TrimSpace(req.CustomerType)
	if ct == "" {
		ct = "business"
	}
	if ct != "business" && ct != "individual" {
		return "customer_type must be 'business' or 'individual'"
	}

	if req.GSTTreatment == nil || strings.TrimSpace(*req.GSTTreatment) == "" {
		return "gst_treatment is required"
	}
	if !validGSTTreatments[*req.GSTTreatment] {
		return "invalid gst_treatment value"
	}

	if req.PlaceOfSupply == nil || strings.TrimSpace(*req.PlaceOfSupply) == "" {
		return "place_of_supply is required"
	}

	tp := "taxable"
	if req.TaxPreference != nil && strings.TrimSpace(*req.TaxPreference) != "" {
		tp = *req.TaxPreference
	}
	if tp != "taxable" && tp != "tax_exempt" {
		return "tax_preference must be 'taxable' or 'tax_exempt'"
	}

	if gstTreatmentsRequiringTaxNumber[*req.GSTTreatment] {
		if req.TaxNumber == nil || strings.TrimSpace(*req.TaxNumber) == "" {
			return "tax_number (GSTIN/UIN) is required for this GST treatment"
		}
		if !validateGSTIN(*req.TaxNumber) {
			return "invalid GSTIN/UIN format"
		}
	} else if req.TaxNumber != nil && *req.TaxNumber != "" {
		if !validateGSTIN(*req.TaxNumber) {
			return "invalid GSTIN/UIN format"
		}
	}

	if req.PAN != nil && *req.PAN != "" {
		if !panRegex.MatchString(*req.PAN) {
			return "invalid PAN format (must be AAAAA9999A)"
		}
	}

	if req.Email != nil && *req.Email != "" {
		if !isValidEmail(*req.Email) {
			return "invalid email address"
		}
	}

	if req.ContactPersonEmail != nil && *req.ContactPersonEmail != "" {
		if !isValidEmail(*req.ContactPersonEmail) {
			return "invalid contact person email address"
		}
	}

	if req.Phone != nil && *req.Phone != "" {
		if n := countDigits(*req.Phone); n < 7 || n > 12 {
			return "phone digits portion must be between 7 and 12 digits"
		}
	}

	if req.ContactPersonMobile != nil && *req.ContactPersonMobile != "" {
		if !digitsOnly.MatchString(*req.ContactPersonMobile) {
			return "contact_person_mobile must contain digits only"
		}
		if n := len(*req.ContactPersonMobile); n < 7 || n > 12 {
			return "contact_person_mobile must be 7-12 digits"
		}
	}

	if tp == "tax_exempt" {
		if req.ExemptionReason == nil || strings.TrimSpace(*req.ExemptionReason) == "" {
			return "exemption_reason is required when tax_preference is 'tax_exempt'"
		}
	}

	return ""
}

func validateUpdateCustomer(req *UpdateCustomerRequest) string {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return "name cannot be empty"
	}

	if req.CustomerType != nil && *req.CustomerType != "business" && *req.CustomerType != "individual" {
		return "customer_type must be 'business' or 'individual'"
	}

	if req.GSTTreatment != nil && !validGSTTreatments[*req.GSTTreatment] {
		return "invalid gst_treatment value"
	}

	if req.PlaceOfSupply != nil && strings.TrimSpace(*req.PlaceOfSupply) == "" {
		return "place_of_supply cannot be empty"
	}

	if req.TaxPreference != nil && *req.TaxPreference != "taxable" && *req.TaxPreference != "tax_exempt" {
		return "tax_preference must be 'taxable' or 'tax_exempt'"
	}

	if req.TaxNumber != nil && *req.TaxNumber != "" && !validateGSTIN(*req.TaxNumber) {
		return "invalid GSTIN/UIN format"
	}

	// if gst_treatment is being updated to a requiring type and tax_number is explicitly cleared
	if req.GSTTreatment != nil && gstTreatmentsRequiringTaxNumber[*req.GSTTreatment] &&
		req.TaxNumber != nil && strings.TrimSpace(*req.TaxNumber) == "" {
		return "tax_number (GSTIN/UIN) is required for this GST treatment"
	}

	if req.PAN != nil && *req.PAN != "" && !panRegex.MatchString(*req.PAN) {
		return "invalid PAN format (must be AAAAA9999A)"
	}

	if req.Email != nil && *req.Email != "" && !isValidEmail(*req.Email) {
		return "invalid email address"
	}

	if req.ContactPersonEmail != nil && *req.ContactPersonEmail != "" && !isValidEmail(*req.ContactPersonEmail) {
		return "invalid contact person email address"
	}

	if req.Phone != nil && *req.Phone != "" {
		if n := countDigits(*req.Phone); n < 7 || n > 12 {
			return "phone digits portion must be between 7 and 12 digits"
		}
	}

	if req.ContactPersonMobile != nil && *req.ContactPersonMobile != "" {
		if !digitsOnly.MatchString(*req.ContactPersonMobile) {
			return "contact_person_mobile must contain digits only"
		}
		if n := len(*req.ContactPersonMobile); n < 7 || n > 12 {
			return "contact_person_mobile must be 7-12 digits"
		}
	}

	// cross-field: only check if both fields are present in the request
	if req.TaxPreference != nil && *req.TaxPreference == "tax_exempt" &&
		req.ExemptionReason != nil && strings.TrimSpace(*req.ExemptionReason) == "" {
		return "exemption_reason is required when tax_preference is 'tax_exempt'"
	}

	return ""
}

// GetCustomers GET /api/v1/customers
func (h *Handler) GetCustomers(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	search := c.Query("search")
	p := utils.GetPaginationParams(c)

	total, err := h.q.CountCustomers(context.Background(), customersdb.CountCustomersParams{
		OrganisationID: orgID,
		Search:         search,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to count customers")
		return
	}

	res, err := h.q.GetCustomers(context.Background(), customersdb.GetCustomersParams{
		OrganisationID: orgID,
		Search:         search,
		Offset:         int32(p.Offset),
		PerPage:        int32(p.PerPage),
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch customers")
		return
	}

	if res == nil {
		res = []customersdb.Customer{}
	}

	response.SuccessWithMeta(c, http.StatusOK, "Customers retrieved successfully", res, &response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: utils.CalculateTotalPages(total, p.PerPage),
	})
}

// GetCustomerByID GET /api/v1/customers/:id
func (h *Handler) GetCustomerByID(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	res, err := h.q.GetCustomerByID(context.Background(), customersdb.GetCustomerByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Customer not found")
		return
	}

	response.Success(c, http.StatusOK, "Customer retrieved successfully", res)
}

// CreateCustomer POST /api/v1/customers
func (h *Handler) CreateCustomer(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply defaults before validation
	if strings.TrimSpace(req.CustomerType) == "" {
		req.CustomerType = "business"
	}

	if errMsg := validateCreateCustomer(&req); errMsg != "" {
		response.Error(c, http.StatusBadRequest, errMsg)
		return
	}

	// Resolve defaults for DB
	taxPreference := "taxable"
	if req.TaxPreference != nil && *req.TaxPreference != "" {
		taxPreference = *req.TaxPreference
	}
	gstTds := false
	if req.GSTTds != nil {
		gstTds = *req.GSTTds
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	res, err := h.q.CreateCustomer(context.Background(), customersdb.CreateCustomerParams{
		OrganisationID:      orgID,
		CustomerType:        req.CustomerType,
		Name:                strings.TrimSpace(req.Name),
		CompanyName:         req.CompanyName,
		Email:               req.Email,
		Phone:               req.Phone,
		GstTreatment:        req.GSTTreatment,
		TaxNumber:           req.TaxNumber,
		BusinessLegalName:   req.BusinessLegalName,
		BusinessTradeName:   req.BusinessTradeName,
		PlaceOfSupply:       req.PlaceOfSupply,
		Pan:                 req.PAN,
		TaxPreference:       taxPreference,
		ExemptionReason:     req.ExemptionReason,
		GstTds:              gstTds,
		PaymentTerm:         req.PaymentTerm,
		Document:            req.Document,
		Address:             req.Address,
		ContactPersonName:   req.ContactPersonName,
		ContactPersonEmail:  req.ContactPersonEmail,
		ContactPersonMobile: req.ContactPersonMobile,
		Remark:              req.Remark,
		IsActive:            isActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create customer")
		return
	}

	notify.Publish(orgID, "Customer added", strings.TrimSpace(req.Name)+" was added.", "info")
	response.Success(c, http.StatusCreated, "Customer created successfully", res)
}

// UpdateCustomer PUT /api/v1/customers/:id
func (h *Handler) UpdateCustomer(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if errMsg := validateUpdateCustomer(&req); errMsg != "" {
		response.Error(c, http.StatusBadRequest, errMsg)
		return
	}

	res, err := h.q.UpdateCustomer(context.Background(), customersdb.UpdateCustomerParams{
		ID:                  id,
		OrganisationID:      orgID,
		CustomerType:        req.CustomerType,
		Name:                req.Name,
		CompanyName:         req.CompanyName,
		Email:               req.Email,
		Phone:               req.Phone,
		GstTreatment:        req.GSTTreatment,
		TaxNumber:           req.TaxNumber,
		BusinessLegalName:   req.BusinessLegalName,
		BusinessTradeName:   req.BusinessTradeName,
		PlaceOfSupply:       req.PlaceOfSupply,
		Pan:                 req.PAN,
		TaxPreference:       req.TaxPreference,
		ExemptionReason:     req.ExemptionReason,
		GstTds:              req.GSTTds,
		PaymentTerm:         req.PaymentTerm,
		Document:            req.Document,
		Address:             req.Address,
		ContactPersonName:   req.ContactPersonName,
		ContactPersonEmail:  req.ContactPersonEmail,
		ContactPersonMobile: req.ContactPersonMobile,
		Remark:              req.Remark,
		IsActive:            req.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update customer")
		return
	}

	response.Success(c, http.StatusOK, "Customer updated successfully", res)
}

// DeleteCustomer DELETE /api/v1/customers/:id
func (h *Handler) DeleteCustomer(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	if err := h.q.DeleteCustomer(context.Background(), customersdb.DeleteCustomerParams{
		ID:             id,
		OrganisationID: orgID,
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete customer")
		return
	}

	response.Success(c, http.StatusOK, "Customer deleted successfully", nil)
}
