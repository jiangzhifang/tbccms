package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jiangzhifang/tbccms/pkg/models"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	/*
		if r.URL.Path != "/" {
			app.notFound(w)
			return
		}
	*/

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
	app.render(w, r, "create.page.tmpl", nil)
}

func (app *application) createCourseware(w http.ResponseWriter, r *http.Request) {

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

	r.ParseMultipartForm(32 << 20)
	files := r.MultipartForm.File["uploadcf"]
	len := len(files)

	if len != 1 {
		app.serverError(w, err)
		return
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

	courseCode := r.MultipartForm.Value["coursecode"][0]
	courseTitle := r.MultipartForm.Value["coursetitle"][0]
	active := true
	if r.MultipartForm.Value["active"][0] == "" {
		active = false
	}

	/*
		errors := make(map[string]string)
		if strings.TrimSpace(courseTitle) == "" {
			errors["courseTitle"] = "This field cannot be blank"
		} else if utf8.RuneCountInString(courseTitle) > 100 {
			errors["courseTitle"] = "This field is too long (maximum is 100 characters)"
		}
	*/
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

	http.Redirect(w, r, fmt.Sprintf("/courseware/%s", courseCode), http.StatusSeeOther)
}
