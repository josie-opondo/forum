package routes

import (
	"net/http"
)

// Register routes
func (r *Routes) RegisterRoutes(mux *http.ServeMux) http.Handler {
	// ===== Protected RESTFUL API Endpoints ===== //

	// === Posts ===
	mux.Handle("/api/posts/create", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.CreatePost)))
	mux.Handle("/api/posts/delete", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.DeletePost)))
	mux.Handle("/api/posts/like", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.PostAddLike)))
	mux.Handle("/api/posts/dislike", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.PostDislike)))
	mux.Handle("/api/posts/image", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.UploadPostImage)))
	// === End Posts ===

	// === Comments ===
	mux.Handle("/api/posts/comments/create", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.CreatePostComment)))
	mux.Handle("/api/comments/delete", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.DeleteComment)))
	mux.Handle("/api/comments/update", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.UpdateComment)))
	mux.Handle("/api/comments/like", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.CommentAddLike)))
	mux.Handle("/api/comments/dislike", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.CommentAddDisLike)))
	// === End Comments ===

	// === Replies ===
	mux.Handle("/api/comments/reply/create", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.CreatePostReply)))
	mux.Handle("/api/comments/reply/like", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.LikeReply)))
	// === End Replies

	// === Auth ===
	mux.Handle("/api/auth/uploadProfilePic", r.auth.AuthMiddleware(http.HandlerFunc(r.authRepo.UploadProfilePic)))
	mux.Handle("/api/user/dashboard", r.auth.AuthMiddleware(http.HandlerFunc(r.authRepo.UserDashboard)))
	mux.Handle("/api/user/editBio", r.auth.AuthMiddleware(http.HandlerFunc(r.authRepo.EditBio)))
	// === End Auth ===

	// === Notifications ===
	mux.Handle("/api/notifications/check", r.auth.AuthMiddleware(http.HandlerFunc(r.postsRepo.CheckNotifications)))
	// === End Notifications ===

	// ===== End Protected RESTFUL API Endpoints ===== //

	// ==== Static files server ====
	fs := r.app.Tmpls.GetProjectRoute("/static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(fs))))

	// ===== Unprotected RESTFUL API Endpoints ===== //

	// === Posts ===
	mux.HandleFunc("/api/posts", r.postsRepo.AllPosts)
	// === End Posts ===

	// Unprotected Auth RESTFUL API Endpoints
	mux.HandleFunc("/api/auth/check", r.authRepo.CheckAuth)
	mux.HandleFunc("/api/auth/register", r.authRepo.RegisterHandler)
	mux.HandleFunc("/api/auth/logout", r.authRepo.LogoutHandler)
	mux.HandleFunc("/api/auth/login", r.authRepo.LoginHandler)
	// ===== End Unprotected RESTFUL API Endpoints =====

	// Page routes
	mux.HandleFunc("/", r.rendersRepo.HomePageHandler)
	mux.HandleFunc("/auth", r.rendersRepo.LoginPageHandler)
	mux.HandleFunc("/dashboard", r.rendersRepo.ProfilePageHandler)
	mux.HandleFunc("/auth-sign-up", r.rendersRepo.SignUpPageHandler)
	mux.HandleFunc("/moderator", r.rendersRepo.ModeratorPageHandler)
	mux.HandleFunc("/admin", r.rendersRepo.AdminPageHandler)

	// CORS middleware
	handler := r.auth.CorsMiddleware(
		r.auth.UserContextMiddleware(
			r.auth.AllowedRoutes(mux),
		),
	)
	return handler
}
