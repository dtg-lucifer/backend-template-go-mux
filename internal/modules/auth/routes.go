package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-username/go-mux-backend-template/internal/core/events"
	"github.com/your-username/go-mux-backend-template/internal/middlewares"
	"github.com/your-username/go-mux-backend-template/internal/utils"
)

// RegisterRoutes mounts all auth routes onto the provided subrouter.
//
//	POST /auth/register  — create a new account
//	POST /auth/login     — authenticate, receive access + refresh tokens
//	GET  /auth/me        — current authenticated user (protected)
//	POST /auth/refresh   — exchange refresh token for a new access token
func RegisterRoutes(r *mux.Router, pool *pgxpool.Pool, bus *events.Bus) {
	svc := newService(pool, bus)
	auth := middlewares.NewAuthMiddleware()

	sub := r.PathPrefix("/auth").Subrouter()

	sub.Handle("/register",
		middlewares.Validate[RegisterInput](http.HandlerFunc(register(svc))),
	).Methods(http.MethodPost)

	sub.Handle("/login",
		middlewares.Validate[LoginInput](http.HandlerFunc(login(svc))),
	).Methods(http.MethodPost)
	
	sub.Handle("/refresh",
	middlewares.Validate[RefreshInput](http.HandlerFunc(refresh(svc))),
	).Methods(http.MethodPost)
	
	sub.Handle("/me",
		auth.Guard(http.HandlerFunc(me(svc))),
	).Methods(http.MethodGet)
}

func register(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := middlewares.BodyFromContext[RegisterInput](r.Context())
		utils.SendResponse(w, svc.Register(r.Context(), input))
	}
}

func login(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := middlewares.BodyFromContext[LoginInput](r.Context())
		utils.SendResponse(w, svc.Login(r.Context(), input))
	}
}

func me(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := middlewares.UIDFromContext(r.Context())
		utils.SendResponse(w, svc.Me(r.Context(), uid))
	}
}

func refresh(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := middlewares.BodyFromContext[RefreshInput](r.Context())
		utils.SendResponse(w, svc.RefreshToken(r.Context(), input))
	}
}
