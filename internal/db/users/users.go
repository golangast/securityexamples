package users

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-playground/validator"
	connect "github.com/golangast/lunarr-goast/internal/db"
	"github.com/golangast/lunarr-goast/internal/security/cookies"
	"github.com/golangast/lunarr-goast/internal/security/crypt"
	"github.com/golangast/lunarr-goast/internal/security/jwt"
	"github.com/labstack/echo/v4"
)

func (u *Users) Exists() error {
	var exists bool
	db, err := connect.DbConnection()
	if err != nil {
		return err
	}
	stmts := db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM user WHERE email=?)", u.Email)
	err = stmts.Scan(&exists)
	if err != nil {
		return err
	}
	db.Close()

	return nil

}
func Exists(email, password, sitetoken string) (bool, error) {
	var passwordhash string
	db, err := connect.DbConnection()
	if err != nil {
		return false, err
	}

	stmts := db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM user WHERE email=? AND sitetoken=?)", email, sitetoken)
	err = stmts.Scan(&passwordhash)
	if err != nil {
		return false, err
	}

	err = crypt.CheckPassword([]byte(passwordhash), []byte(password))
	if err != nil {
		return false, err
	}

	db.Close()

	return true, nil

}
func (u *Users) CheckLogin(c echo.Context, email, sitetokens string) (error, string) {
	//https://gitlab.com/zendrulat123/gocourses/-/blob/main/go/handler/user/processlogin/processlogin.go
	db, err := connect.DbConnection()
	if err != nil {
		return err, "wrong input"
	}
	var exists string
	ctx, cancel := context.WithTimeout(context.Background(), 56666*time.Millisecond)
	defer cancel()
	stmts := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM user WHERE email=? AND sitetoken=?)", email, sitetokens)
	err = stmts.Scan(&exists)
	if err != nil {
		return err, "wrong input"
	}

	if exists == "0" {
		return err, "wrong input"
	} else {
		u, err := u.GetUserByEmail(email, sitetokens)
		if err != nil {
			return err, "wrong input"
		}

		err = crypt.CheckPassword([]byte(u.PasswordHash), []byte(u.PasswordRaw))
		if err != nil {
			return err, "wrong input"
		}

		cookie, err := cookies.ReadCookie(c, "lannarr")
		if err != nil {
			return err, "wrong input"
		}

		fmt.Println(u)
		if cookie.Name != u.SessionName && cookie.Value != u.SessionKey {
			return err, "wrong input"
		}

		//header
		hkey := c.Response().Header().Get("headerkey")

		crypt.CheckPassword([]byte(hkey), []byte("lunarr"))

		//context

		db.Close()

		return nil, ""
	}

}
func (u *Users) Create() error {

	db, err := connect.DbConnection()
	if err != nil {
		return err
	}
	// Create a statement to insert data into the `users` table.
	stmt, err := db.PrepareContext(context.Background(), "INSERT INTO `user` (`email`, `passwordhash`, `isdisabled`, `sessionkey`, `sessionname`, `sitetoken`, `movieperm`) VALUES (?, ?,?, ?,?,?,?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Insert data into the `users` table.
	_, err = stmt.ExecContext(context.Background(), u.Email, u.PasswordHash, u.Isdisabled, u.SessionKey, u.SessionName, u.SessionToken, u.SiteToken, "admin")
	if err != nil {
		panic(err)
	}

	db.Close()
	return nil
}

func (u *Users) JWT() error {
	t, err := jwt.CreateJWT(u.SessionName, u.SessionKey)
	if err != nil {
		return err
	}
	u.SessionToken = t
	return nil
}
func (u *Users) SessionKeys(c echo.Context) error {
	err := cookies.WriteCookie(c, u.SessionName, u.SessionKey)
	if err != nil {
		return err
	}
	return nil
}

// ValidateValuer implements validator.CustomTypeFunc
func (users *Users) Validate(user *Users) error {

	// use a single instance of Validate, it caches struct info
	//var validate *validator.Validate

	validate := validator.New()

	// returns InvalidValidationError for bad validation input, nil or ValidationErrors ( []FieldError )
	err := validate.Struct(user)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return err
		}

		fmt.Println("------ List of tag fields with error ---------")

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println(err.StructField())
			fmt.Println(err.ActualTag())
			fmt.Println(err.Kind())
			fmt.Println(err.Value())
			fmt.Println(err.Param())
			fmt.Println("---------------")
		}

	}
	return nil
	// save user to database
}

func (user *Users) GetUser(id, idkey string) (Users, error) {
	db, err := connect.DbConnection()
	if err != nil {
		return *user, err
	}
	var (
		email        string
		passwordhash string
		isdisabled   string
		sessionkey   string
		sessionname  string
		sessiontoken string
		sitetoken    string
		movieperm    string
		u            Users
	)

	//get from database
	stmt, err := db.Prepare("SELECT * FROM user WHERE id = ? AND SiteToken = ?")
	if err != nil {
		return u, err
	}
	err = stmt.QueryRow(id, idkey).Scan(email, passwordhash, isdisabled, sessionkey, sessionname, sessiontoken, sitetoken, movieperm)
	if err != nil {
		return u, err
	}
	u = Users{ID: id, Email: email, PasswordHash: passwordhash, Isdisabled: isdisabled, SessionKey: sessionkey, SessionName: sessionname, SessionToken: sessiontoken, SiteToken: sitetoken, MoviePerm: movieperm}
	defer db.Close()
	defer stmt.Close()
	switch err {
	case sql.ErrNoRows:
		fmt.Println("was nil !!!!!!!!!!!!!1", email)
		fmt.Println("No rows were returned!")
		// close db when not in use
		return u, nil

	case nil:
		fmt.Println("was nil !!!!!!!!!!!!!12", email)

		// close db when not in use
		return u, nil

	default:
		fmt.Println("was nil !!!!!!!!!!!!!13", email)

		fmt.Println("default!!!!!!!!!!!!")

		return u, nil
	}

}

// https://golangbot.com/mysql-select-single-multiple-rows/
func (user Users) GetUserByEmail(email, idkey string) (Users, error) {
	// ctx, cancelfunc := context.WithTimeout(context.Background(), 500*time.Second)
	// defer cancelfunc()
	var (
		id           string
		passwordhash string
		passwordraw  string
		isdisabled   string
		sessionkey   string
		sessionname  string
		sessiontoken string
		sitetoken    string
		movieperm    string
		u            Users
	)
	db, err := connect.DbConnection()
	if err != nil {
		return u, err
	}

	//get from database
	stmt, err := db.Prepare("SELECT * FROM user WHERE email = ? AND sitetoken = ?")
	if err != nil {
		return u, err
	}
	err = stmt.QueryRow(email, idkey).Scan(&id, &email, &passwordhash, &passwordraw, &isdisabled, &sessionkey, &sessionname, &sessiontoken, &sitetoken, &movieperm)
	if err != nil {
		return u, err
	}
	u = Users{ID: id, Email: email, PasswordHash: passwordhash, Isdisabled: isdisabled, SessionKey: sessionkey, SessionName: sessionname, SessionToken: sessiontoken, SiteToken: sitetoken, MoviePerm: movieperm}
	defer db.Close()
	defer stmt.Close()
	switch err {
	case sql.ErrNoRows:
		fmt.Println("was nil !!!!!!!!!!!!!1", email)
		fmt.Println("No rows were returned!")
		// close db when not in use
		return u, nil

	case nil:
		fmt.Println("was nil !!!!!!!!!!!!!12", email)

		// close db when not in use
		return u, nil

	default:
		fmt.Println("was nil !!!!!!!!!!!!!13", email)

		fmt.Println("default!!!!!!!!!!!!")

		return u, nil
	}

	// stmt, err := db.PrepareContext(ctx, "SELECT * FROM user WHERE email = ? AND sitetoken = ?")
	// if err != nil {
	// 	log.Printf("Error %s when preparing SQL statement", err)
	// 	return u, err
	// }
	// defer stmt.Close()

	// rows := db.QueryRowContext(ctx, "SELECT * FROM user WHERE email = ? AND sitetoken = ?", &user.Email, &user.SiteToken)
	// if err != nil {
	// 	return u, err
	// }

	// err = rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.PasswordRaw, &user.Isdisabled, &user.SessionKey, &user.SessionName, &user.SessionToken, &user.SiteToken, &user.MoviePerm)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if err := rows.Err(); err != nil {
	// 	return u, err
	// }
	// defer db.Close()
	// return u, nil

}

func (user Users) SetUserSitetoken(sitetoken string) error {
	//opening database
	db, err := connect.DbConnection()
	if err != nil {
		return err
	}

	//prepare statement so that no sql injection
	stmt, err := db.Prepare("update user set sitetoken=?")
	if err != nil {
		return err
	}

	//execute qeury
	_, err = stmt.Exec(sitetoken)
	if err != nil {
		return err
	}
	return nil
}

type Users struct {
	ID           string `param:"id" query:"id" form:"id" json:"id" xml:"id"`
	Email        string `valid:"type(string),required" param:"email" query:"email" form:"email" json:"email" xml:"email" validate:"required,email" mod:"trim"`
	PasswordHash string `valid:"type(string),required" param:"passwordhash" query:"passwordhash" form:"passwordhash" json:"passwordhash" xml:"passwordhash"`
	PasswordRaw  string `valid:"type(string)" param:"passwordraw" query:"passwordraw" form:"password" json:"passwordraw" xml:"passwordraw" validate:"required" scrub:"password" mod:"trim"`
	Isdisabled   string `valid:"type(string)" param:"isdisabled" query:"isdisabled" form:"isdisabled" json:"isdisabled" xml:"isdisabled"`
	SessionKey   string `valid:"type(string)" param:"sessionkey" query:"sessionkey" form:"sessionkey" json:"sessionkey" xml:"sessionkey"`
	SessionName  string `valid:"type(string)" param:"sessionname" query:"sessionname" form:"sessionname" json:"sessionname" xml:"sessionname"`
	SessionToken string `valid:"type(string)" param:"sessiontoken" query:"sessiontoken" form:"sessiontoken" json:"sessiontoken" xml:"sessiontoken"`
	SiteToken    string `valid:"type(string),required" param:"sitetoken" query:"sitetoken" form:"sitetoken" json:"sitetoken" xml:"sitetoken"`
	MoviePerm    string `valid:"type(string)" param:"movieperm" query:"movieperm" form:"movieperm" json:"movieperm" xml:"movieperm"`
}
