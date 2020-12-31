package models

import (
	"errors"
	"time"
)

//ErrNoRecord 创建日志
var ErrNoRecord = errors.New("models: no matching record found")

/*
//Snippet 定义snippet
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}
*/

type Courseware struct {
	CourseCode  string
	CourseTitle string
	Created     time.Time
	Active      bool
}

type CoursewareFile struct {
	CourseCode         string
	CoursewareFileName string
}
