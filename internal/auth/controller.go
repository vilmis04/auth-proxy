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
			ctx.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		isAuthorized := c.service.getIsAuthorized(jwtCookie.Value)
		ctx.JSON(http.StatusOK, isAuthorized)
	})

}
