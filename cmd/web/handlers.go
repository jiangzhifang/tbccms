package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jiangzhifang/tbccms/pkg/formfiles"
	"github.com/jiangzhifang/tbccms/pkg/forms"
	"github.com/jiangzhifang/tbccms/pkg/models"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	c, err := app.coursewares.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Coursewares: c,
	})
}

func (app *application) showCourseware(w http.ResponseWriter, r *http.Request) {
	// courseCode := r.URL.Query().Get("coursecode")
	courseCode := r.URL.Query().Get(":coursecode")

	c, err := app.coursewares.Get(courseCode)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		Courseware: c,
	})

}

func (app *application) createCoursewareForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		Form:     forms.New(nil),
		FormFile: formfiles.New(nil),
	})
}

func (app *application) createCourseware(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	form := forms.New(r.PostForm)
	form.Required("coursetitle", "coursecode")
	form.MaxLength("coursetitle", 100)
	form.MaxLength("coursecode", 100)

	formfile := formfiles.New(r.MultipartForm)
	formfile.Required("uploadcf")

	if !form.Valid() || !formfile.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form, FormFile: formfile})
		return
	}

	ctx := context.Background()
	endpoint := "21tbminio.21tb.com"
	accessKeyID := "21tbminio"
	secretAccessKey := "21tbminio_key"
	useSSL := true

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	bucketName := "apibucket"
	location := "us-east-1"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	files := r.MultipartForm.File["uploadcf"]

	courseCode := r.MultipartForm.Value["coursecode"][0]
	courseTitle := r.MultipartForm.Value["coursetitle"][0]
	active := true

	if r.MultipartForm.Value["active"][0] == "" {
		active = false
	}

	objectName := files[0].Filename
	// contentType := "application/zip"
	contentType := "application/octet-stream"

	file, err := files[0].Open()
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	uploadinfo, err := minioClient.PutObject(ctx, bucketName, objectName, file, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}

	err = app.coursewares.Insert(courseCode, courseTitle, active)
	if err != nil {
		app.serverError(w, err)
		return
	}

	coursewareFileName := objectName

	err = app.coursewareFiles.Insert(courseCode, coursewareFileName)
	if err != nil {
		app.serverError(w, err)
		return
	}

	log.Println(uploadinfo)
	app.session.Put(r, "flash", "课件创建成功！")

	http.Redirect(w, r, fmt.Sprintf("/courseware/%s", courseCode), http.StatusSeeOther)
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}
func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Validate the form contents using the form helper we made earlier.
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)
	// If there are any errors, redisplay the signup form.
	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err == models.ErrDuplicateEmail {
		form.Errors.Add("email", "此邮箱已经被注册")
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "注册成功，请登录。")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}
func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "邮箱或密码不正确")
		app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}
	// Add the ID of the current user to the session, so that they are now 'logged
	// in'.
	app.session.Put(r, "authenticatedUserID", id)
	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/courseware/create", http.StatusSeeOther)
}
func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "authenticatedUserID")
	app.session.Put(r, "flash", "你已经成功退出！")
	http.Redirect(w, r, "/", 303)
}
