package middlewares

import (
	"context"
	"net/http"

	"forum/forumapp"
)

type contextKey string

const userIDKey contextKey = "userID"

type AuthContext struct {
	app *forumapp.ForumApp
}

func NewAuthContext(app *forumapp.ForumApp) *AuthContext {
	return &AuthContext{app: app}
}

// SetUserIDInContext adds the user ID to the request context
func (a *AuthContext) SetUserIDInContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserIDFromContext retrieves the user ID from the request context
func (a *AuthContext) GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// AuthMiddleware protects routes from unauthorized access and redirects if not logged in
func (a *AuthContext) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if session token exists
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusFound) // Redirect to login page
			return
		}

		// Verify session in global store
		var userID string
		found := false

		a.app.Sessions.Range(func(key, value interface{}) bool {
			if token, ok := value.(*http.Cookie); ok && token.Value == cookie.Value {
				userID, _ = key.(string)
				found = true
				return false // Stop iteration
			}
			return true
		})

		if !found {
			http.Redirect(w, r, "/auth", http.StatusFound) // Redirect to login page
			return
		}

		// Set user ID in request context for handlers
		ctx := a.SetUserIDInContext(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
