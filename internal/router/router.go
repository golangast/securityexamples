package router

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golangast/lunarr-goast/internal/db/users"
	"github.com/golangast/lunarr-goast/internal/handler/get/home"
	"github.com/golangast/lunarr-goast/internal/handler/get/player"
	"github.com/golangast/lunarr-goast/internal/handler/post/addmovies"
	"github.com/golangast/lunarr-goast/internal/handler/post/creates"
	"github.com/golangast/lunarr-goast/internal/handler/post/login"
	"github.com/golangast/lunarr-goast/internal/security/crypt"
	"github.com/golangast/lunarr-goast/internal/security/tokens"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Routes is for routing
func Routes(e *echo.Echo) {

	// Restricted group
	r := e.Group("/restricted")

	// Register the validator
	//get router
	e.GET("/home", home.Home)
	r.GET("/player/:movieid/:userid/:idkey", player.Player)

	//post router
	e.POST("/userlogin", login.Login)
	e.POST("/usercreate", creates.Creates)
	e.POST("/addmovies", addmovies.Addmovies)

	// Create a new Secure middleware instance.
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	// Server header
	//r.Use(ServerHeader)
	// Set up key auth middleware
	queryAuthConfig := middleware.KeyAuthConfig{
		KeyLookup: "query:idkey,header:headkey,cookie:lannarr",
		Validator: func(key string, c echo.Context) (bool, error) {
			user := new(users.Users)
			userid := c.Param("userid")
			idkey := c.Param("idkey")

			u, err := user.GetUser(userid, idkey)
			if err != nil {
				return false, err
			}
			err, exists := u.CheckLogin(c, userid, idkey)
			if err != nil {
				fmt.Println("middleware", exists)
				return false, err
			}

			fmt.Println(key, " keylookup")
			b := tokens.Checktokencontext(key)
			return b, nil
		},
		ErrorHandler: func(error, echo.Context) error {
			var err error
			fmt.Println(err, "idkey")
			return err
		},
	}

	r.Use(middleware.KeyAuthWithConfig(queryAuthConfig))

}
func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key, err := crypt.HashPassword("lunarr")
		if err != nil {
			fmt.Println(err)
		}
		c.Response().Header().Set("headerkey", key)
		return next(c)
	}
}

// jwtCustomClaims are custom claims extending default ones.
// See https://github.com/golang-jwt/jwt for more examples
type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func Accessible(c echo.Context) error {
	return c.String(http.StatusOK, "Accessible")
}

func Restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwtCustomClaims)
	name := claims.Name
	return c.String(http.StatusOK, "Welcome "+name+"!")
}
