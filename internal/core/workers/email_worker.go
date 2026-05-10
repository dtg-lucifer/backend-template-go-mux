// Package workers contains job-processor functions consumed by the queue manager.
// Each file handles one job type. Workers are pure functions: they receive a typed
// job struct and return error. They are passed as callbacks to queue.Manager.ConsumeXxx.
package workers

import (
	"github.com/your-username/go-mux-backend-template/internal/core/queue"
	"github.com/your-username/go-mux-backend-template/pkg"
)

// WelcomeEmailHandler returns a handler for WelcomeEmailJob messages.
// Replace the body with a real email send in production.
func WelcomeEmailHandler(logger *pkg.Logger) func(queue.WelcomeEmailJob) error {
	return func(job queue.WelcomeEmailJob) error {
		// TODO: call your email service here, e.g. mailer.SendWelcome(job.Email)
		logger.Info("[WORKER] Sending welcome email",
			"user_id", job.UserID,
			"email", job.Email,
		)
		return nil
	}
}
