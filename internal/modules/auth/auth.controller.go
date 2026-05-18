package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-username/go-mux-backend-template/internal/core/events"
	"github.com/your-username/go-mux-backend-template/internal/middlewares"
	"github.com/your-username/go-mux-backend-template/internal/utils"
)

// Controller owns the auth subrouter and all auth HTTP handlers.
// Router is the only exported field — mount it in the route registry.
type Controller struct {
	Router *mux.Router

	svc  *Service
	auth *middlewares.AuthMiddleware
}

// NewController constructs a Controller and registers all routes.
func NewController(pool *pgxpool.Pool, bus *events.Bus) *Controller {
	c := &Controller{
		Router: mux.NewRouter(),
		svc:    NewService(pool, bus),
		auth:   middlewares.NewAuthMiddleware(),
	}
	c.registerRoutes()
	return c
}

// registerRoutes wires every auth endpoint to its handler method.
//
//	POST /register  — create a new account
//	POST /login     — authenticate, receive access + refresh tokens
//	POST /refresh   — exchange refresh token for a new access token
//	GET  /me        — current authenticated user (protected)
func (c *Controller) registerRoutes() {
	c.Router.Handle("/register",
		middlewares.Validate[RegisterInput](http.HandlerFunc(c.register)),
	).Methods(http.MethodPost)

	c.Router.Handle("/login",
		middlewares.Validate[LoginInput](http.HandlerFunc(c.login)),
	).Methods(http.MethodPost)

	c.Router.Handle("/refresh",
		middlewares.Validate[RefreshInput](http.HandlerFunc(c.refresh)),
	).Methods(http.MethodPost)

	c.Router.Handle("/me",
		c.auth.Guard(http.HandlerFunc(c.me)),
	).Methods(http.MethodGet)
}

// ---- Handlers ------------------------------------------------------------------

func (c *Controller) register(w http.ResponseWriter, r *http.Request) {
	input := middlewares.BodyFromContext[RegisterInput](r.Context())
	utils.SendResponse(w, c.svc.Register(r.Context(), input, r))
}

func (c *Controller) login(w http.ResponseWriter, r *http.Request) {
	input := middlewares.BodyFromContext[LoginInput](r.Context())
	utils.SendResponse(w, c.svc.Login(r.Context(), input, r))
}

func (c *Controller) me(w http.ResponseWriter, r *http.Request) {
	uid := middlewares.UIDFromContext(r.Context())
	utils.SendResponse(w, c.svc.Me(r.Context(), uid))
}

func (c *Controller) refresh(w http.ResponseWriter, r *http.Request) {
	input := middlewares.BodyFromContext[RefreshInput](r.Context())
	utils.SendResponse(w, c.svc.RefreshToken(r.Context(), input))
}
