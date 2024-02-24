package auth

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

const MONTH int = 30 * 24 * 3600

var BASE_URL = os.Getenv("BASE_URL")

type Controller struct {
	service   *Service
	authGroup *gin.RouterGroup
}

func NewController(apiGroup *gin.RouterGroup) *Controller {
	return &Controller{
		service:   NewService(),
		authGroup: apiGroup.Group("auth"),
	}
}

func (c *Controller) Use() {
	c.authGroup.GET("is-authorized", func(ctx *gin.Context) {
		jwtCookie, err := ctx.Request.Cookie(accessToken.ACCESS_TOKEN)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		isAuthorized := c.service.getIsAuthorized(jwtCookie.Value)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, isAuthorized)
	})

	c.authGroup.POST("sign-up", func(ctx *gin.Context) {
		token, err := c.service.signUp(ctx.Request)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctx.SetCookie(accessToken.ACCESS_TOKEN, *token, MONTH, "/", BASE_URL, true, true)
		ctx.Writer.WriteHeader(http.StatusCreated)
	})

	c.authGroup.POST("login", func(ctx *gin.Context) {
		token, err := c.service.login(ctx.Request)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctx.SetCookie(accessToken.ACCESS_TOKEN, *token, MONTH, "/", BASE_URL, true, true)
		ctx.Writer.WriteHeader(http.StatusOK)
	})

	c.authGroup.POST("logout", func(ctx *gin.Context) {
		ctx.SetCookie(accessToken.ACCESS_TOKEN, "", 0, "/", BASE_URL, true, true)
		ctx.Writer.WriteHeader(http.StatusOK)
	})
}
