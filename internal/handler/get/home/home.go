package home

import (
	"net/http"

	"github.com/golangast/lunarr-goast/internal/security/cookies"
	"github.com/golangast/lunarr-goast/internal/security/jwt"
	"github.com/golangast/lunarr-goast/internal/security/tokens"
	"github.com/labstack/echo/v4"
)

func Home(c echo.Context) error {

	sitetokens, err := jwt.CreateJWT("lannarr", tokens.Timername())
	if err != nil {
		return err
	}

	err = cookies.WriteCookie(c, "lannarr", sitetokens)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"t": sitetokens,
	})

}
