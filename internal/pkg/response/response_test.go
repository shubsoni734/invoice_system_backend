package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSuccessAndErrorShapes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("request_id", "req-1")
	Success(c, http.StatusOK, "ok", gin.H{"id": "1"})

	var body Response
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || body.Message != "ok" || body.RequestID != "req-1" {
		t.Fatalf("body = %+v", body)
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	Error(c, http.StatusBadRequest, "nope")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	ValidationError(c, http.StatusUnprocessableEntity, []FieldError{{Field: "email", Message: "required"}})
	if w.Code != http.StatusUnprocessableEntity || len(w.Body.Bytes()) == 0 {
		t.Fatalf("validation response missing: %d %s", w.Code, w.Body.String())
	}
}
