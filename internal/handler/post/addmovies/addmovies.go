package addmovies

import (
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/golangast/lunarr-goast/internal/db/movies"
	"github.com/golangast/lunarr-goast/internal/db/users"
	"github.com/golangast/lunarr-goast/internal/security/cookies"
	"github.com/golangast/lunarr-goast/internal/security/crypt"
	"github.com/golangast/lunarr-goast/internal/security/tokens"
	"github.com/labstack/echo/v4"
)

func Addmovies(c echo.Context) error {
	// Create a new user
	movie := new(movies.Movies)
	user := new(users.Users)
	cookie, err := cookies.ReadCookie(c, "lannarr")
	if err != nil {
		return err
	}

	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if user.Email == "" && user.SiteToken == "" && cookie.Name != "lannarr" && tokens.Checktoken(cookie.Value) && tokens.Checktokencontext(user.SiteToken) {
		return echo.ErrUnauthorized
	}
	user.PasswordRaw, err = crypt.HashPassword(user.PasswordRaw)
	if err != nil {
		return err
	}
	if err := c.Bind(movie); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	movie.Userid = user.ID
	// Validate the user
	valid, err := govalidator.ValidateStruct(movie)
	if err != nil || !valid {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err, exist := movie.ExistVideo(user.ID, movie.Video)
	if err != nil || !exist {
		if err := movie.Create(user.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

	}

	userfill, err := user.GetUserByEmail(user.Email, user.SiteToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// header
	err, exists := user.CheckLogin(c, user.Email, user.SiteToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	m, err := movie.GetAllMovie()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	err = movies.UploadImage(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	err = movies.UploadVideo(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	mm, err := movies.GetFiles("assets/video")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	fmt.Println(mm)

	return c.Render(http.StatusOK, "base.html", map[string]interface{}{
		"EX": exists,
		"M":  m,
		"U":  userfill,
		"ST": userfill.SiteToken,
	})

}
