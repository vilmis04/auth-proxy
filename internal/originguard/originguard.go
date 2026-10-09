// Package originguard rejects cross-origin state-changing requests.
package originguard

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/gin-gonic/gin"
)

// Middleware blocks POST/PUT/PATCH/DELETE requests whose Origin header is
// neither one of allowedOrigins nor the host being served. Requests without
// an Origin header (non-browser clients) pass. SameSite=Lax alone does not
// stop cross-origin requests from sibling subdomains, so this complements it.
func Middleware(allowedOrigins []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		switch ctx.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			return
		}

		origin := ctx.GetHeader("Origin")
		if origin == "" || slices.Contains(allowedOrigins, origin) {
			return
		}
		if u, err := url.Parse(origin); err == nil && u.Host == ctx.Request.Host {
			return
		}

		ctx.String(http.StatusForbidden, "origin not allowed")
		ctx.Abort()
	}
}
