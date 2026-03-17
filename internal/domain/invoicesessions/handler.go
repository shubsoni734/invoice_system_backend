package invoicesessions

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	invoicesessionsdb "github.com/your-org/invoice-backend/internal/domain/invoicesessions/sqlc"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type Handler struct {
	q *invoicesessionsdb.Queries
}

func NewHandler(q *invoicesessionsdb.Queries) *Handler {
	return &Handler{q: q}
}

type CreateInvoiceSessionRequest struct {
	Year            int32   `json:"year" binding:"required"`
	Prefix          string  `json:"prefix" binding:"required"`
	CurrentSequence *int32  `json:"current_sequence"`
}

func getOrgID(c *gin.Context) (uuid.UUID, error) {
	orgIDRaw := c.GetString(constants.CtxOrgID)
	return uuid.Parse(orgIDRaw)
}

// GetInvoiceSessions GET /api/v1/invoice-sessions
func (h *Handler) GetInvoiceSessions(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	res, err := h.q.GetInvoiceSessions(context.Background(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch invoice sessions")
		return
	}

	if res == nil {
		res = []invoicesessionsdb.InvoiceSession{}
	}

	response.Success(c, http.StatusOK, "Invoice sessions retrieved", res)
}

// GetInvoiceSessionByID GET /api/v1/invoice-sessions/:id
func (h *Handler) GetInvoiceSessionByID(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	res, err := h.q.GetInvoiceSessionByID(context.Background(), invoicesessionsdb.GetInvoiceSessionByIDParams{
		ID:             id,
		OrganisationID: orgID,
	})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invoice session not found")
		return
	}

	response.Success(c, http.StatusOK, "Invoice session retrieved", res)
}

// CreateInvoiceSession POST /api/v1/invoice-sessions
func (h *Handler) CreateInvoiceSession(c *gin.Context) {
	orgID, err := getOrgID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organisation ID")
		return
	}

	var req CreateInvoiceSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	var sequence int32 = 0
	if req.CurrentSequence != nil {
		sequence = *req.CurrentSequence
	}

	res, err := h.q.CreateInvoiceSession(context.Background(), invoicesessionsdb.CreateInvoiceSessionParams{
		OrganisationID:  orgID,
		Year:            req.Year,
		Prefix:          req.Prefix,
		CurrentSequence: sequence,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create invoice session")
		return
	}

	response.Success(c, http.StatusCreated, "Invoice session created", res)
}
