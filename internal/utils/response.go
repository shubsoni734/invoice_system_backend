package utils

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

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	requestID, _ := c.Get("request_id")
	rid := ""
	if requestID != nil {
		rid = requestID.(string)
	}
	c.JSON(statusCode, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: rid,
	})
}

func SuccessResponseWithMeta(c *gin.Context, statusCode int, message string, data interface{}, meta *Meta) {
	requestID, _ := c.Get("request_id")
	rid := ""
	if requestID != nil {
		rid = requestID.(string)
	}
	c.JSON(statusCode, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		RequestID: rid,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	requestID, _ := c.Get("request_id")
	rid := ""
	if requestID != nil {
		rid = requestID.(string)
	}
	c.JSON(statusCode, Response{
		Success:   false,
		Message:   message,
		RequestID: rid,
	})
}

func ValidationErrorResponse(c *gin.Context, statusCode int, errors []FieldError) {
	requestID, _ := c.Get("request_id")
	rid := ""
	if requestID != nil {
		rid = requestID.(string)
	}
	c.JSON(statusCode, Response{
		Success:   false,
		Message:   "Validation failed",
		Errors:    errors,
		RequestID: rid,
	})
}
