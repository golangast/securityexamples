package player

import (
	"net/http"

	"github.com/golangast/lunarr-goast/internal/db/movies"
	"github.com/golangast/lunarr-goast/internal/db/users"
	"github.com/golangast/lunarr-goast/internal/security/cookies"
	"github.com/golangast/lunarr-goast/internal/security/tokens"
	"github.com/golangast/lunarr-goast/internal/security/validate"
	"github.com/labstack/echo/v4"
)

func Player(c echo.Context) error {
	userid := c.Param("userid")
	movieid := c.Param("movieid")
	idkey := c.Param("idkey")
	// Create a new user
	movie := new(movies.Movies)
	user := new(users.Users)
	cookie, err := cookies.ReadCookie(c, "lannarr")
	if err != nil {
		return err
	}

	if userid == "" && movieid == "" && idkey == "" && cookie.Name != "lannarr" && tokens.Checktoken(cookie.Value) && tokens.Checktokencontext(idkey) {
		return echo.ErrUnauthorized
	}

	if err := c.Bind(movie); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := validate.ValidateRequest(c, &movie); err != nil {
		return err
	}
	if err := movie.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := movie.ExistID(userid, movieid); err != nil {
		return c.Render(http.StatusOK, "base.html", map[string]interface{}{
			"exists": "exists",
		})
	}
	userfill, err := user.GetUser(userid, idkey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	m, err := movie.GetMovie(movieid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Render(http.StatusOK, "player.html", map[string]interface{}{
		"M":  m,
		"U":  userfill,
		"ST": userfill.SiteToken,
	})

}
