package formfiles

import (
	"mime/multipart"
)

type FormFile struct {
	*multipart.Form
	Errors errors
}

func New(data *multipart.Form) *FormFile {
	return &FormFile{
		data,
		errors(map[string][]string{}),
	}
}

func (f *FormFile) Required(fields ...string) {
	for _, field := range fields {
		value := f.File[field]
		if len(value) == 0 {
			f.Errors.Add(field, "必须上传文件")
		}
	}

}

func (f *FormFile) Valid() bool {
	return len(f.Errors) == 0
}
