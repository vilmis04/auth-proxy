package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

const USER = "user"

// https://pkg.go.dev/net/http/httputil#NewSingleHostReverseProxy
func NewReverseProxy() (*httputil.ReverseProxy, error) {
	url, err := url.Parse(os.Getenv("SERVICE_URL"))
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie(accessToken.ACCESS_TOKEN)
		if err != nil {
			log.Printf("[AuthMiddleware] ERR %v \n", err)
			ctx.Writer.WriteHeader(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		user, err := accessToken.Validate(cookie.Value)
		if err != nil {
			log.Printf("[AuthMiddleware] ERR %v \n", err)
			ctx.Writer.WriteHeader(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		ctx.Request.Header.Set(USER, *user)
		log.Printf("[AuthMiddleware] authorized for: %v", *user)
		ctx.Next()
	}
}

func ProxyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := ctx.Request.Header.Get(USER)
		if user == "" {
			log.Printf("[ProxyMiddleware] ERR user name in header: %v \n", user)
			ctx.Writer.WriteHeader(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		proxy, err := NewReverseProxy()
		if err != nil {
			log.Printf("[ProxyMiddleware] ERR %v \n", err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			ctx.Abort()
			return
		}

		log.Printf("[ProxyMiddleware] proxied request %v", ctx.Request.URL)
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
		ctx.Next()
	}
}
