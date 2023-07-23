package login

import (
	"net/http"

	"github.com/golangast/lunarr-goast/internal/db/users"
	"github.com/golangast/lunarr-goast/internal/security/cookies"
	"github.com/labstack/echo/v4"
)

func Login(c echo.Context) error {

	user := new(users.Users)

	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := user.Validate(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := user.JWT(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	err := cookies.WriteCookie(c, "lannarr", user.SiteToken)
	if err != nil {
		return err
	}

	err = user.SetUserSitetoken(user.SiteToken)
	if err != nil {
		return err
	}

	err, exist := user.CheckLogin(c, user.Email, user.SiteToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Render(http.StatusOK, "base.html", map[string]interface{}{
		"EX": exist,
		"M":  "",
		"U":  user,
		"ST": user.SiteToken,
	})

}
