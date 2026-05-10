package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"oolio-backend-challenge/internal/errs"
)

const headerRequestID = "X-Request-ID"

type contextKey string

const requestIDKey contextKey = "request_id"

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// RequestID ensures every request has a traceable ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader(headerRequestID))
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(string(requestIDKey), id)
		c.Writer.Header().Set(headerRequestID, id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDKey, id))
		c.Next()
	}
}

// SecurityHeaders adds a baseline set of safe default headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; style-src 'self' 'unsafe-inline'; script-src 'self'; font-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		c.Next()
	}
}

// MaxBodyBytes caps request size before handlers decode JSON.
func MaxBodyBytes(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limit > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}

// JSONNotFound returns the standard API envelope for unknown API routes.
func JSONNotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		errs.JSON(c, http.StatusNotFound, errs.CodeNotFound, "resource not found")
	}
}
