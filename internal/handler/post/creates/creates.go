package creates

import (
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/golangast/lunarr-goast/internal/db/users"
	"github.com/golangast/lunarr-goast/internal/security/cookies"
	"github.com/golangast/lunarr-goast/internal/security/crypt"
	"github.com/labstack/echo/v4"
)

func Creates(c echo.Context) error {

	// Create a new user
	user := new(users.Users)
	cookie, err := cookies.ReadCookie(c, "lannarr")
	if err != nil {
		return err
	}
	user.SessionName = cookie.Name
	user.SessionKey = cookie.Value
	user.Isdisabled = "false"
	user.PasswordHash, err = crypt.HashPassword(user.PasswordHash)
	if err != nil {
		return err
	}

	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	// Validate the user
	valid, err := govalidator.ValidateStruct(user)
	if err != nil || !valid {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := user.Validate(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := user.Exists(); err != nil {
		return c.Render(http.StatusOK, "home.html", map[string]interface{}{
			"exists": "exists",
		})
	}
	if err := user.JWT(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err := user.Create(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Render(http.StatusOK, "base.html", map[string]interface{}{
		"M":  "",
		"U":  user,
		"ST": user.SiteToken,
	})

}
