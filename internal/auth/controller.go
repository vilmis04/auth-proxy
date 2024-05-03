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
var PATH = "/"

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

		var status int
		username := ""
		user, _ := c.service.getIsAuthenticated(jwtCookie.Value)
		if user != nil {
			status = http.StatusOK
			username = *user
		} else {
			status = http.StatusUnauthorized
		}

		ctx.String(status, username)
	})

	c.authGroup.POST("sign-up", func(ctx *gin.Context) {
		token, serverErr, clientErr := c.service.signUp(ctx.Request)
		if serverErr != nil {
			log.Printf("[Controller] /sign-up ERR: %v", serverErr)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			ctx.Abort()
			return
		}
		if clientErr != nil {
			log.Printf("[Controller] /sign-up ERR: %v", clientErr)
			ctx.String(http.StatusBadRequest, clientErr.Error())
			ctx.Abort()
			return
		}

		ctx.SetCookie(accessToken.ACCESS_TOKEN, *token, MONTH, PATH, BASE_URL, true, true)
		ctx.JSON(http.StatusCreated, TokenResponse{Token: *token})
	})

	c.authGroup.POST("login", func(ctx *gin.Context) {
		token, serverErr, clientErr := c.service.login(ctx.Request)
		if serverErr != nil {
			log.Printf("[Controller] /login ERR: %v\n", serverErr)
			ctx.String(http.StatusInternalServerError, serverErr.Error())
			ctx.Abort()
			return
		}
		if clientErr != nil {
			log.Printf("[Controller] /login ERR: %v\n", clientErr)
			ctx.String(http.StatusUnauthorized, clientErr.Error())
			ctx.Abort()
			return
		}

		ctx.SetCookie(accessToken.ACCESS_TOKEN, *token, MONTH, PATH, BASE_URL, true, true)
		ctx.JSON(http.StatusOK, TokenResponse{Token: *token})
	})

	c.authGroup.POST("logout", func(ctx *gin.Context) {
		ctx.SetCookie(accessToken.ACCESS_TOKEN, "", 0, PATH, BASE_URL, true, true)
		ctx.Writer.WriteHeader(http.StatusOK)
	})
}
