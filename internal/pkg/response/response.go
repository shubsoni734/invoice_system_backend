package response

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      interface{}  `json:"data,omitempty"`
	Meta      *Meta        `json:"meta,omitempty"`
	Errors    []FieldError `json:"errors,omitempty"`
	RequestID string       `json:"request_id"`
}

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

func SuccessWithMeta(c *gin.Context, statusCode int, message string, data interface{}, meta *Meta) {
	c.JSON(statusCode, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		RequestID: getRequestID(c),
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success:   false,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

func ValidationError(c *gin.Context, statusCode int, errors []FieldError) {
	c.JSON(statusCode, Response{
		Success:   false,
		Message:   "Validation failed",
		Errors:    errors,
		RequestID: getRequestID(c),
	})
}

func getRequestID(c *gin.Context) string {
	rid, _ := c.Get("request_id")
	if rid != nil {
		return rid.(string)
	}
	return ""
}
