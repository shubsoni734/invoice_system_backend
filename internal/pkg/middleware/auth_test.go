package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/invoice-backend/internal/pkg/utils"
	"github.com/your-org/invoice-backend/internal/shared/constants"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager, err := utils.NewJWTManager("correct-secret-value-32-characters", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := manager.GenerateToken(&utils.Claims{UserID: "user-1", OrgID: "org-1", Role: "admin"})
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/private", Auth(manager), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user": c.GetString(constants.CtxUserID),
			"org":  c.GetString(constants.CtxOrgID),
			"role": c.GetString(constants.CtxUserRole),
		})
	})

	missing := httptest.NewRequest(http.MethodGet, "/private", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, missing)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing header status = %d", rec.Code)
	}

	bad := httptest.NewRequest(http.MethodGet, "/private", nil)
	bad.Header.Set("Authorization", "Token "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, bad)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad scheme status = %d", rec.Code)
	}

	ok := httptest.NewRequest(http.MethodGet, "/private", nil)
	ok.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, ok)
	if rec.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["user"] != "user-1" || body["org"] != "org-1" || body["role"] != "admin" {
		t.Fatalf("body = %+v", body)
	}
}

func TestRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		c.Set(constants.CtxUserRole, c.GetHeader("X-Role"))
		c.Next()
	}, RBAC("admin"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	staff := httptest.NewRequest(http.MethodGet, "/admin", nil)
	staff.Header.Set("X-Role", "staff")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, staff)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("staff status = %d", rec.Code)
	}

	admin := httptest.NewRequest(http.MethodGet, "/admin", nil)
	admin.Header.Set("X-Role", "admin")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, admin)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("admin status = %d", rec.Code)
	}
}
