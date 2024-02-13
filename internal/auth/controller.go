package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service   *Service
	authGroup *gin.RouterGroup
}

func NewController(server *gin.Engine) *Controller {
	return &Controller{
		service:   NewService(),
		authGroup: server.Group("auth"),
	}
}

func (c *Controller) Use() {
	c.authGroup.GET("is-authorized", func(ctx *gin.Context) {
		jwtCookie, err := ctx.Request.Cookie("jwt")
		if err != nil {
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		isAuthorized := c.service.getIsAuthorized(jwtCookie.Value)
		ctx.JSON(http.StatusOK, isAuthorized)
	})

	c.authGroup.POST("sign-up", func(ctx *gin.Context) {
		token, err := c.service.signUp(ctx.Request)
		if err != nil {
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctx.SetCookie("access_token", *token, 30*24*3600, "/", "/", true, true)
		ctx.Writer.WriteHeader(http.StatusCreated)
	})

}
