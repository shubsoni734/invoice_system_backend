package common

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-org/invoice-backend/internal/pkg/email"
	"github.com/your-org/invoice-backend/internal/pkg/response"
)

type Handler struct {
	emailClient *email.Client
	testToken   string
}

func NewHandler(emailClient *email.Client) *Handler {
	return &Handler{emailClient: emailClient}
}

type TestEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// TestEmail POST /api/v1/common/test-email
func (h *Handler) TestEmail(c *gin.Context) {
	var req TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	err := h.emailClient.SendTestEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send test email: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Test email sent successfully to "+req.Email, nil)
}

// SendSample POST /api/v1/email-check/:token
func (h *Handler) SendSample(c *gin.Context) {
	if !tokenMatches(h.testToken, c.Param("token")) {
		response.Error(c, http.StatusNotFound, "Not found")
		return
	}

	var req TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Enter a valid email address")
		return
	}

	if err := h.emailClient.SendTestEmail(req.Email); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to send test email: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Sample email sent to "+req.Email, nil)
}

func tokenMatches(secret, provided string) bool {
	secret = strings.TrimSpace(secret)
	provided = strings.TrimSpace(provided)
	if secret == "" || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(secret), []byte(provided)) == 1
}
