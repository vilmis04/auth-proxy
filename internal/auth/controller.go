package auth

import (
	"cmp"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

// maxBodyBytes caps JSON request bodies on the auth endpoints.
const maxBodyBytes = 4 << 10

var BASE_URL = cmp.Or(os.Getenv("BASE_URL"), "localhost")
var PATH = "/"

// Limits are optional middlewares for the credential endpoints.
type Limits struct {
	Login  gin.HandlerFunc
	SignUp gin.HandlerFunc
}

type Controller struct {
	service   *Service
	authGroup *gin.RouterGroup
	limits    Limits
}

func noop(*gin.Context) {}

func NewController(apiGroup *gin.RouterGroup, service *Service, limits Limits) *Controller {
	if limits.Login == nil {
		limits.Login = noop
	}
	if limits.SignUp == nil {
		limits.SignUp = noop
	}

	return &Controller{
		service:   service,
		authGroup: apiGroup.Group("auth"),
		limits:    limits,
	}
}

func (c *Controller) maxAge() int {
	return int(c.service.signer.TTL().Seconds())
}

func decodeBody(ctx *gin.Context, dst any) error {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBodyBytes)
	return json.NewDecoder(ctx.Request.Body).Decode(dst)
}

// fail writes the response for an error returned by the service.
func fail(ctx *gin.Context, route string, err error) {
	var ce *ClientError
	if errors.As(err, &ce) {
		log.Printf("[Controller] %v client ERR: %v", route, err)
		ctx.String(ce.Status, ce.Msg)
	} else {
		log.Printf("[Controller] %v ERR: %v", route, err)
		ctx.Status(http.StatusInternalServerError)
	}
	ctx.Abort()
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

	c.authGroup.POST("sign-up", c.limits.SignUp, func(ctx *gin.Context) {
		var body signUpRequest
		if err := decodeBody(ctx, &body); err != nil {
			fail(ctx, "/sign-up", clientErr(http.StatusBadRequest, "invalid request body"))
			return
		}

		token, err := c.service.signUp(body)
		if err != nil {
			fail(ctx, "/sign-up", err)
			return
		}

		ctx.SetCookie(accessToken.ACCESS_TOKEN, *token, c.maxAge(), PATH, BASE_URL, true, true)
		ctx.JSON(http.StatusCreated, UserResponse{Username: body.Username})
	})

	c.authGroup.POST("login", c.limits.Login, func(ctx *gin.Context) {
		var body loginRequest
		if err := decodeBody(ctx, &body); err != nil {
			fail(ctx, "/login", clientErr(http.StatusBadRequest, "invalid request body"))
			return
		}

		token, err := c.service.login(body)
		if err != nil {
			fail(ctx, "/login", err)
			return
		}

		ctx.SetCookie(accessToken.ACCESS_TOKEN, *token, c.maxAge(), PATH, BASE_URL, true, true)
		ctx.JSON(http.StatusOK, UserResponse{Username: body.Username})
	})

	c.authGroup.POST("logout", func(ctx *gin.Context) {
		ctx.SetCookie(accessToken.ACCESS_TOKEN, "", -1, PATH, BASE_URL, true, true)
		ctx.Writer.WriteHeader(http.StatusOK)
	})
}
