package proxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

const (
	// USER carries the authenticated username to the upstream service.
	USER = "user"
	// InternalTokenHeader proves to the upstream that the request came through this proxy.
	InternalTokenHeader = "X-Internal-Token"

	userContextKey = "proxy.user"
)

type ctxKey struct{}

// New builds the reverse proxy once at startup.
// Peers inside trustedProxies (CIDRs or single IPs) may supply X-Forwarded-*
// headers, as Coolify's Traefik does; for everyone else they are rebuilt from
// the connection.
func New(target *url.URL, internalToken string, trustedProxies []string) (*httputil.ReverseProxy, error) {
	var trusted []netip.Prefix
	for _, entry := range trustedProxies {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			addr, addrErr := netip.ParseAddr(entry)
			if addrErr != nil {
				return nil, err
			}
			prefix = netip.PrefixFrom(addr, addr.BitLen())
		}
		trusted = append(trusted, prefix)
	}

	isTrusted := func(remoteAddr string) bool {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			return false
		}
		addr, err := netip.ParseAddr(host)
		if err != nil {
			return false
		}
		for _, prefix := range trusted {
			if prefix.Contains(addr.Unmap()) {
				return true
			}
		}
		return false
	}

	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)

			// Whatever the client sent is dropped; only the verified identity is set.
			pr.Out.Header.Del(USER)
			pr.Out.Header.Set(USER, pr.In.Context().Value(ctxKey{}).(string))
			pr.Out.Header.Set(InternalTokenHeader, internalToken)

			// The upstream gets the identity, not the JWT.
			pr.Out.Header.Del("Cookie")
			for _, cookie := range pr.In.Cookies() {
				if cookie.Name != accessToken.ACCESS_TOKEN {
					pr.Out.AddCookie(cookie)
				}
			}

			if isTrusted(pr.In.RemoteAddr) {
				pr.Out.Header["X-Forwarded-For"] = pr.In.Header["X-Forwarded-For"]
				pr.SetXForwarded()
				for _, name := range []string{"X-Forwarded-Proto", "X-Forwarded-Host"} {
					if value := pr.In.Header.Get(name); value != "" {
						pr.Out.Header.Set(name, value)
					}
				}
			} else {
				pr.SetXForwarded()
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[Proxy] upstream ERR for %v: %v", r.URL.Path, err)
			w.WriteHeader(http.StatusBadGateway)
		},
	}, nil
}

// AuthMiddleware validates the access token cookie and stores the username
// in the Gin context for Handler.
func AuthMiddleware(signer *accessToken.Signer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie(accessToken.ACCESS_TOKEN)
		if err != nil {
			log.Printf("[AuthMiddleware] ERR %v \n", err)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		user, err := signer.Validate(cookie.Value)
		if err != nil {
			log.Printf("[AuthMiddleware] ERR %v \n", err)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set(userContextKey, *user)
		log.Printf("[AuthMiddleware] authorized for: %v", *user)
	}
}

// Handler forwards the request to the upstream for the authenticated user.
func Handler(proxy *httputil.ReverseProxy) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := ctx.GetString(userContextKey)
		if user == "" {
			log.Printf("[Proxy] request without an authenticated user")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		req := ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxKey{}, user))
		proxy.ServeHTTP(ctx.Writer, req)
	}
}
