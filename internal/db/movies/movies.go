package movies

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator"
	connect "github.com/golangast/lunarr-goast/internal/db"
	"github.com/labstack/echo/v4"
)

type Movies struct {
	ID     string `param:"id"   query:"id" form:"id" json:"id" xml:"id"`
	Image  string `param:"image" valid:"type(string)"  query:"image" form:"image" json:"image" xml:"image" validate:"required,image" mod:"trim"`
	Video  string `param:"video" valid:"type(string)"  query:"video" form:"video" json:"video" xml:"video" validate:"required,video" mod:"trim"`
	Title  string `param:"title" valid:"type(string),required" query:"title" form:"title" json:"title" xml:"title" validate:"required,title" mod:"trim"`
	Year   string `param:"year" valid:"type(string),required" query:"year" form:"year" json:"year" xml:"year" validate:"required,year" mod:"trim"`
	Userid string `param:"userid" valid:"type(string)" query:"userid" form:"userid" json:"userid" xml:"userid" validate:"required,userid" mod:"trim"`
}

func (movie *Movies) GetMovie(id string) (Movies, error) {
	db, err := connect.DbConnection()
	if err != nil {
		return *movie, err
	}
	var (
		image string
		video string
		title string
		year  string
	)

	//get from database
	stmt, err := db.Prepare("SELECT * FROM mov WHERE id = ?")
	connect.ErrorCheck(err)
	err = stmt.QueryRow(image).Scan(image, video, title, year)
	m := Movies{ID: id, Image: image, Video: video, Title: title, Year: year}
	defer db.Close()
	defer stmt.Close()
	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!", image)
		// close db when not in use
		return m, nil

	case nil:
		fmt.Println("was nil !!!!!!!!!!!!!12", image)

		// close db when not in use
		return m, nil

	default:

		fmt.Println("default!!!!!!!!!!!!")

		return m, nil
	}

}
func (m *Movies) ExistID(userid, moviesid string) error {
	var exists bool
	db, err := connect.DbConnection()
	if err != nil {
		return err
	}
	stmts := db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM mov WHERE userid=? AND id=?)", userid, moviesid)
	err = stmts.Scan(&exists)
	if err != nil {
		return err
	}
	db.Close()

	return nil

}
func (m *Movies) ExistVideo(userid, video string) (error, bool) {
	var exists bool
	db, err := connect.DbConnection()
	if err != nil {
		return err, exists
	}
	stmts := db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM mov WHERE userid=? AND video=?)", userid, video)
	err = stmts.Scan(&exists)
	if err != nil {
		return err, exists
	}
	db.Close()

	return nil, exists

}

// ValidateValuer implements validator.CustomTypeFunc
func (m *Movies) Validate() error {

	// use a single instance of Validate, it caches struct info
	validate := validator.New()

	// returns InvalidValidationError for bad validation input, nil or ValidationErrors ( []FieldError )
	err := validate.Struct(m)
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
func (m *Movies) Create(userid string) error {

	db, err := connect.DbConnection()
	if err != nil {
		return err
	}

	query := "INSERT INTO `mov` (`image`, `video`, `title`, `year`, 'userid') VALUES (?, ?,?, ?, ?)"
	insertResult, err := db.ExecContext(context.Background(), query, m.Image, m.Title, m.Video, m.Year, m.Userid)
	if err != nil {
		log.Fatalf("impossible insert : %s", err)
		return err
	}
	ids, err := insertResult.LastInsertId()
	if err != nil {
		log.Fatalf("impossible to retrieve last inserted id: %s", err)
		return err

	}
	log.Printf("inserted id: %d", ids)

	db.Close()
	return nil
}

func (m Movies) GetAllMovie() ([]Movies, error) {
	var (
		id     string
		image  string
		video  string
		title  string
		year   string
		userid string
		mm     []Movies
	)
	db, err := connect.DbConnection()
	if err != nil {
		return mm, err
	}

	//get from database
	stmt, err := db.Prepare("SELECT * FROM mov")
	connect.ErrorCheck(err)
	// Execute the statement and get the results.
	rows, err := stmt.Query()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	defer db.Close()
	defer stmt.Close()
	// Iterate over the rows and print the data.
	for rows.Next() {

		err := rows.Scan(&id, &image, &video, &title, &year, &userid)
		if err != nil {
			panic(err)
		}
		m = Movies{Image: image, Title: title, Year: year}
		mm = append(mm, m)
		switch err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!", image)
			return mm, nil

		case nil:
			fmt.Println("was nil", image)
			return mm, nil

		default:
			fmt.Println("default")
			return mm, nil
		}
	}
	return mm, nil
}

func UploadImage(c echo.Context) error {
	// Read form fields

	//------------
	// Read files
	//------------

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	if file == nil && file.Filename == "" &&
		!strings.Contains(file.Filename, ".jpg") &&
		!strings.Contains(file.Filename, ".png") {
		return c.JSON(400, "no file")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create("assets/img/" + file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func UploadVideo(c echo.Context) error {
	// Read form fields

	//------------
	// Read files
	//------------

	file, err := c.FormFile("video")
	if err != nil {
		return err
	}

	if file == nil && file.Filename == "" &&
		!strings.Contains(file.Filename, ".mkv") &&
		!strings.Contains(file.Filename, ".mp4") &&
		!strings.Contains(file.Filename, ".webm") {
		return c.JSON(400, "no file")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create("assets/video/" + file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func GetFiles(path string) ([]string, error) {
	var filenames []string
	files, err := os.ReadDir(path)
	if err != nil {
		return filenames, err
	}

	for _, file := range files {
		fmt.Println(path + file.Name())
		filenames = append(filenames, path+file.Name())
	}

	return filenames, nil
}
