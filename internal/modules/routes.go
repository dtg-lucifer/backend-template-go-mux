// Package modules is the central route registry.
// Add a single line here to mount a new module — nothing else needs to change.
package modules

import (
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-username/go-mux-backend-template/internal/core/cache"
	"github.com/your-username/go-mux-backend-template/internal/core/events"
	"github.com/your-username/go-mux-backend-template/internal/modules/auth"
	"github.com/your-username/go-mux-backend-template/internal/modules/health"
	"github.com/your-username/go-mux-backend-template/pkg"
)

// Register mounts every module's routes onto the API subrouter.
// apiRouter is already scoped to the API prefix (e.g. /api/v1).
func Register(apiRouter *mux.Router, pool *pgxpool.Pool, redis cache.Cache, bus *events.Bus, startTime time.Time, logger *pkg.Logger) {
	health.RegisterRoutes(apiRouter, pool, redis, startTime)

	authCtrl := auth.NewController(pool, bus, logger)
	apiRouter.PathPrefix("/auth").Handler(authCtrl.Router)

	// Add new modules here:
	// userCtrl := user.NewController(pool, bus, logger)
	// apiRouter.PathPrefix("/users").Handler(userCtrl.Router)
}
