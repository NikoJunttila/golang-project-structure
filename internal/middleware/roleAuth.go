package middleware

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

// RequireRoles allows access only if the user's role is one of the allowed roles.
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	// Create a fast lookup map
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		roleSet[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())

			rawRole, ok := claims["role"]
			if !ok {
				http.Error(w, "Missing role in token", http.StatusForbidden)
				return
			}

			// JWT claims come in as interface{} (could be float64, string, etc.)
			var roleStr string
			switch v := rawRole.(type) {
			case string:
				roleStr = v
			case float64: // just in case it's numeric
				roleStr = formatRoleNumber(int(v))
			default:
				http.Error(w, "Invalid role format", http.StatusForbidden)
				return
			}

			if _, found := roleSet[roleStr]; !found {
				http.Error(w, "Insufficient role privileges", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Optional: translate numeric roles if needed (fallback)
func formatRoleNumber(n int) string {
	switch n {
	case 1:
		return "user"
	case 5:
		return "moderator"
	case 10:
		return "admin"
	default:
		return "unknown"
	}
}
