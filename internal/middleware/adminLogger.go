// Package middleware contains all middleware logic
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nikojunttila/community/internal/cache"
	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// responseWriter wrapper to capture status code and response
type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
	if !w.written {
		w.statusCode = statusCode
		w.written = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	return w.ResponseWriter.Write(data)
}

// AdminAuditMiddleware logs all admin actions to the database
func AdminAuditMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Generate unique request ID for this request
			requestID := fmt.Sprintf("req_%d_%s", startTime.UnixNano(), generateShortID())

			// Add request ID to context for potential use in handlers
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			// Capture request body (if present)
			var requestBody string
			if r.Body != nil && r.ContentLength > 0 {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					requestBody = string(bodyBytes)
					// Restore the body for the actual handler
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			// Wrap response writer to capture status code
			auditWriter := &auditResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Get admin user from context (should be available after JWT middleware)
			admin, err := cache.GetUser(ctx)
			if err != nil {
				logger.Error(r.Context(), err, "Failed to get admin user for audit log")
				next.ServeHTTP(w, r)
				return
			}

			// Extract target user ID from request (this will vary based on your routes)
			targetUserID := extractTargetUserID(r, requestBody)

			// Determine action based on HTTP method
			action := mapHTTPMethodToAction(r.Method)

			// Extract resource from path
			resource := extractResourceFromPath(r.URL.Path)

			// Process the request
			next.ServeHTTP(auditWriter, r)

			// Calculate response time
			responseTime := time.Since(startTime)

			// Log to database asynchronously to avoid blocking the response
			go func() {
				err := logAdminAction(context.Background(), adminAuditParams{
					AdminUserID:    admin.ID,
					AdminEmail:     admin.Email,
					TargetUserID:   targetUserID,
					Action:         action,
					Resource:       resource,
					Method:         r.Method,
					Path:           r.URL.Path,
					QueryParams:    r.URL.RawQuery,
					RequestBody:    requestBody,
					IPAddress:      r.RemoteAddr,
					UserAgent:      r.UserAgent(),
					StatusCode:     auditWriter.statusCode,
					ResponseTimeMs: responseTime.Milliseconds(),
					Timestamp:      startTime,
					RequestID:      requestID,
				})
				if err != nil {
					logger.Error(ctx, err, "Failed to log admin audit")
				}
			}()
		})
	}
}

type adminAuditParams struct {
	AdminUserID    string
	AdminEmail     string
	TargetUserID   string
	Action         string
	Resource       string
	Method         string
	Path           string
	QueryParams    string
	RequestBody    string
	IPAddress      string
	UserAgent      string
	StatusCode     int
	ResponseTimeMs int64
	Timestamp      time.Time
	RequestID      string
}

// logAdminAction saves the audit log to database
func logAdminAction(ctx context.Context, params adminAuditParams) error {
	_, err := db.Get().CreateAuditLog(ctx, db.CreateAuditLogParams{
		AdminUserID:    params.AdminUserID,
		AdminEmail:     params.AdminEmail,
		TargetUserID:   params.TargetUserID,
		Action:         params.Action,
		Resource:       params.Resource,
		Method:         params.Method,
		Path:           params.Path,
		QueryParams:    params.QueryParams,
		RequestBody:    params.RequestBody,
		IpAddress:      params.IPAddress,
		UserAgent:      params.UserAgent,
		StatusCode:     int64(params.StatusCode),
		ResponseTimeMs: params.ResponseTimeMs,
		Timestamp:      params.Timestamp,
		RequestID:      params.RequestID,
	})
	return err
}

// mapHTTPMethodToAction converts HTTP methods to audit actions
func mapHTTPMethodToAction(method string) string {
	switch method {
	case "GET":
		return "VIEW"
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return method
	}
}

// extractResourceFromPath extracts the resource type from URL path
func extractResourceFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		// /admin/users -> "users"
		// /admin/orders -> "orders"
		return parts[1]
	}
	return "unknown"
}

// extractTargetUserID attempts to extract target user ID from various sources
func extractTargetUserID(r *http.Request, requestBody string) string {
	// Try URL parameters first (e.g., /admin/users/123)
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i, part := range pathParts {
		if part == "users" && i+1 < len(pathParts) {
			return pathParts[i+1]
		}
	}

	// Try query parameters
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		return userID
	}
	if userID := r.URL.Query().Get("target_user_id"); userID != "" {
		return userID
	}

	// Try to parse from JSON body
	if requestBody != "" {
		var bodyMap map[string]any
		if err := json.Unmarshal([]byte(requestBody), &bodyMap); err == nil {
			if userID, ok := bodyMap["user_id"].(string); ok {
				return userID
			}
			if userID, ok := bodyMap["target_user_id"].(string); ok {
				return userID
			}
			if email, ok := bodyMap["email"].(string); ok {
				return email // Use email as identifier if user_id not available
			}
		}
	}
	return "unknown"
}

// generateShortID generates a short random ID for request tracking
func generateShortID() string {
	return strconv.FormatInt(time.Now().UnixNano()%1000000, 36)
}
