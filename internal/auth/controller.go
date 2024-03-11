package auth

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

const MONTH int = 30 * 24 * 3600

var BASE_URL = cmp.Or(os.Getenv("BASE_URL"), "localhost")

type Controller struct {
	service   *Service
	authGroup *gin.RouterGroup
}

type TokenResponse struct {
	Token string `json:"token"`
}

func NewController(apiGroup *gin.RouterGroup) *Controller {
	return &Controller{
		service:   NewService(),
		authGroup: apiGroup.Group("auth"),
	}
}

func (c *Controller) Use() {
	c.authGroup.GET("is-authenticated", func(ctx *gin.Context) {
		jwtCookie, err := ctx.Request.Cookie(accessToken.ACCESS_TOKEN)
		if err != nil {
			log.Printf("[Controller] /is-authenticated ERR: %v", err)
			ctx.Writer.WriteHeader(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		isAuthenticated := c.service.getIsAuthenticated(jwtCookie.Value)
		var status int
		if isAuthenticated {
			status = http.StatusOK
		} else {
			status = http.StatusUnauthorized
		}

		ctx.Writer.WriteHeader(status)
	})

	c.authGroup.POST("sign-up", func(ctx *gin.Context) {
		token, err := c.service.signUp(ctx.Request)
		if err != nil {
			log.Printf("[Controller] /sign-up ERR: %v", err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.SetCookie(accessToken.ACCESS_TOKEN, *token, MONTH, "/", BASE_URL, true, true)
		ctx.JSON(http.StatusCreated, TokenResponse{Token: *token})
	})

	c.authGroup.POST("login", func(ctx *gin.Context) {
		token, err := c.service.login(ctx.Request)
		if err != nil {
			log.Printf("[Controller] /login ERR: %v", err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			ctx.Abort()
			return
		}

		ctx.SetCookie(accessToken.ACCESS_TOKEN, *token, MONTH, "/", BASE_URL, true, true)
		ctx.JSON(http.StatusOK, TokenResponse{Token: *token})
	})

	c.authGroup.POST("logout", func(ctx *gin.Context) {
		ctx.SetCookie(accessToken.ACCESS_TOKEN, "", 0, "/", BASE_URL, true, true)
		ctx.Writer.WriteHeader(http.StatusOK)
	})
}
