package authrepo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// RegisterHandler handles user registration.
func (h *AuthRepo) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.res.SetError(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	// Extract form values
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	imagePath, err := h.shared.SaveImage(r)
	if err != nil {
		h.res.SetError(w, err, http.StatusBadRequest)
		return
	}

	// Create the user
	user := &User{
		Email:     email,
		FirstName: username,
		LastName:  username,
		Password:  password,
		Image:     imagePath[1:],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Register the user
	err = h.user.Register(user)
	if err != nil {
		h.res.SetError(w, err, http.StatusConflict)
		return
	}

	// Generate a token (e.g., JWT)
	token := h.Sessions.GenerateToken(user.UserID)

	// Set the session cookie
	http.SetCookie(w, &token)

	// Respond with success and token
	h.res.Data = user
	h.res.Err = false
	h.res.Message = "User registered successfully"

	// Respond with JSON
	if err := h.res.WriteJSON(w, *h.res, http.StatusCreated); err != nil {
		h.res.SetError(w, err, http.StatusInternalServerError)
		return
	}
}

// LoginHandler handles user login.
func (h *AuthRepo) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.res.SetError(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.res.SetError(w, errors.New("invalid request body"), http.StatusBadRequest)
		return
	}

	user, err := h.user.Login(req.Email, req.Password)
	if err != nil {
		h.res.SetError(w, err, http.StatusUnauthorized)
		return
	}

	// Check if user already has a valid session
	if cookie, err := r.Cookie("session_token"); err == nil {
		if storedCookie, exists := h.Sessions.Sess.Load(user.UserID); exists {
			if token, ok := storedCookie.(*http.Cookie); ok && token.Value == cookie.Value {
				h.res.Err = false
				h.res.Message = "Login successful (existing session)"
				h.res.Data = user
				h.res.WriteJSON(w, *h.res, http.StatusOK)
				return
			}
		}
	}

	// Generate a token
	token := h.Sessions.GenerateToken(user.UserID)

	// Set the session cookie
	http.SetCookie(w, &token)

	// Respond with success and token
	h.res.Err = false
	h.res.Message = "Login successful"
	h.res.Data = &user

	err = h.res.WriteJSON(w, *h.res, http.StatusOK)
	if err != nil {
		h.res.SetError(w, err, http.StatusInternalServerError)
		return
	}
}

// LogoutHandler handles user logout.
func (h *AuthRepo) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.res.SetError(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	// Get the session token from the cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			h.res.SetError(w, errors.New("no session found"), http.StatusUnauthorized)
			return
		}
		h.res.SetError(w, err, http.StatusBadRequest)
		return
	}

	sessionToken := cookie.Value

	h.Sessions.Sess.Delete(sessionToken)

	// Clear the session cookie by setting an expired cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
		Path:    "/",
	})

	// Respond with success
	h.res.Err = false
	h.res.Message = "Logout successful"

	err = h.res.WriteJSON(w, *h.res, http.StatusOK)
	if err != nil {
		h.res.SetError(w, err, http.StatusInternalServerError)
		return
	}
}

// CheckAuth confirms if a user is logged in
func (h *AuthRepo) CheckAuth(w http.ResponseWriter, r *http.Request) {
	// Return a success response if this protected route is reached
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"signedIn": true})
}

// Posts handler (dummy implementation)
func (h *AuthRepo) PostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"posts": []}`)
		return
	}

	h.res.SetError(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
}
