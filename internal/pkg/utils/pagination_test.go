package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		query   string
		page    int
		perPage int
		offset  int
	}{
		{"?page=2&per_page=10", 2, 10, 10},
		{"?page=0&per_page=500", 1, 20, 0},
		{"?page=abc", 1, 20, 0},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/"+tc.query, nil)
		got := GetPaginationParams(c)
		if got.Page != tc.page || got.PerPage != tc.perPage || got.Offset != tc.offset {
			t.Errorf("%s => %+v", tc.query, got)
		}
	}
}

func TestCalculateTotalPages(t *testing.T) {
	if CalculateTotalPages(21, 10) != 3 {
		t.Fatal("expected 3 pages")
	}
	if CalculateTotalPages(20, 10) != 2 {
		t.Fatal("expected 2 pages")
	}
	if CalculateTotalPages(5, 0) != 0 {
		t.Fatal("expected 0 pages when per page is 0")
	}
}
