package notify

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/invoice-backend/internal/pkg/email"
	"github.com/your-org/invoice-backend/internal/pkg/response"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

type item struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Kind      string    `json:"kind"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type readRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

type subscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func List(c *gin.Context) {
	orgID, userID, ok := actor(c)
	if !ok || pool == nil {
		return
	}
	rows, err := pool.Query(c.Request.Context(), `
		SELECT n.id, n.title, n.message, n.kind, n.created_at, (r.user_id IS NOT NULL)
		FROM notifications n
		LEFT JOIN notification_reads r ON r.notification_id = n.id AND r.user_id = $2
		WHERE n.organisation_id = $1
		ORDER BY n.created_at DESC
		LIMIT 30
	`, orgID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load notifications")
		return
	}
	defer rows.Close()
	items := []item{}
	for rows.Next() {
		var row item
		if err := rows.Scan(&row.ID, &row.Title, &row.Message, &row.Kind, &row.CreatedAt, &row.Read); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to read notifications")
			return
		}
		items = append(items, row)
	}
	response.Success(c, http.StatusOK, "Notifications", items)
}

func MarkRead(c *gin.Context) {
	orgID, userID, ok := actor(c)
	if !ok || pool == nil {
		return
	}
	var req readRequest
	_ = c.ShouldBindJSON(&req)
	ctx := c.Request.Context()
	query := `
		INSERT INTO notification_reads (notification_id, user_id)
		SELECT n.id, $2 FROM notifications n
		WHERE n.organisation_id = $1
		ON CONFLICT DO NOTHING
	`
	args := []any{orgID, userID}
	if len(req.IDs) > 0 {
		query = `
			INSERT INTO notification_reads (notification_id, user_id)
			SELECT n.id, $2 FROM notifications n
			WHERE n.organisation_id = $1 AND n.id = ANY($3)
			ON CONFLICT DO NOTHING
		`
		args = append(args, req.IDs)
	}
	if _, err := pool.Exec(ctx, query, args...); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update notifications")
		return
	}
	response.Success(c, http.StatusOK, "Notifications updated", nil)
}

func SaveSubscription(c *gin.Context) {
	orgID, userID, ok := actor(c)
	if !ok || pool == nil {
		return
	}
	var req subscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Endpoint) == "" || req.Keys.P256dh == "" || req.Keys.Auth == "" {
		response.Error(c, http.StatusBadRequest, "Invalid push subscription")
		return
	}
	_, err := pool.Exec(c.Request.Context(), `
		INSERT INTO push_subscriptions (organisation_id, user_id, endpoint, p256dh, auth_secret)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, endpoint) DO UPDATE SET p256dh = EXCLUDED.p256dh, auth_secret = EXCLUDED.auth_secret
	`, orgID, userID, strings.TrimSpace(req.Endpoint), req.Keys.P256dh, req.Keys.Auth)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save push subscription")
		return
	}
	response.Success(c, http.StatusOK, "Push subscription saved", nil)
}

func RemoveSubscription(c *gin.Context) {
	_, userID, ok := actor(c)
	if !ok || pool == nil {
		return
	}
	var req subscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Endpoint) == "" {
		response.Error(c, http.StatusBadRequest, "Invalid push subscription")
		return
	}
	_, err := pool.Exec(c.Request.Context(), `DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`, userID, strings.TrimSpace(req.Endpoint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove push subscription")
		return
	}
	response.Success(c, http.StatusOK, "Push subscription removed", nil)
}

// MailClientForOrg returns the Brevo client for this organisation, or the server default.
func MailClientForOrg(c *gin.Context) *email.Client {
	orgID, err := uuid.Parse(c.GetString(constants.CtxOrgID))
	if err != nil || pool == nil {
		return email.NewClient(baseMail)
	}
	settings, err := loadOrgMail(c.Request.Context(), orgID)
	if err != nil {
		return email.NewClient(baseMail)
	}
	return email.NewClient(resolveMail(settings))
}

func actor(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	orgID, err := uuid.Parse(c.GetString(constants.CtxOrgID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid organization context")
		return uuid.Nil, uuid.Nil, false
	}
	userID, err := uuid.Parse(c.GetString(constants.CtxUserID))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user context")
		return uuid.Nil, uuid.Nil, false
	}
	return orgID, userID, true
}
