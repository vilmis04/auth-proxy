package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Register adds unauthenticated probes. /healthz only says the process is
// serving, so a database blip does not get a healthy proxy restarted.
// /readyz also checks that the database answers.
func Register(router gin.IRoutes, db *sql.DB) {
	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})
	router.GET("/readyz", func(ctx *gin.Context) {
		pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), time.Second)
		defer cancel()
		if db == nil || db.PingContext(pingCtx) != nil {
			ctx.String(http.StatusServiceUnavailable, "database unavailable")
			return
		}
		ctx.String(http.StatusOK, "ok")
	})
}
