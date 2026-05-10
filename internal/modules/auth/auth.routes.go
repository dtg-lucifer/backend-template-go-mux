package auth

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-username/go-mux-backend-template/internal/core/events"
	"github.com/your-username/go-mux-backend-template/pkg"
)

// Deprecated: use NewController(pool, bus, logger) and mount ctrl.Router directly.
// This shim exists only for callers that have not yet been updated.
func RegisterRoutes(r *mux.Router, pool *pgxpool.Pool, bus *events.Bus, logger *pkg.Logger) {
	ctrl := NewController(pool, bus, logger)
	r.PathPrefix("/auth").Handler(ctrl.Router)
}
