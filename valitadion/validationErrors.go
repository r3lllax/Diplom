package valitadion

import (
	"GIN/internal/env"
	"GIN/internal/files"
	"encoding/json"
	"fmt"
	"strings"
)

var stringTagsErrors = map[string]string{
	"min":              "Поле имеет длинну меньше разрешенной",
	"max":              "Поле имеет длинну больше разрешенной",
	"required":         "Поле обязательно",
	"alpha":            "Поле содержит недопустимые символы (разрешена только латиница)",
	"email":            "Поле должно быть формата email",
	"required_without": "Поле обязательно, если нет другого",
	"exists":           "Значение поля не существует",
	"unique":           "Значение поля уже используется",
	"eqfield":          "Значение не совпадает",
	"photo":            "Доступны только форматы",
	"filesize":         fmt.Sprintf("Превышен допустимый размер файла (%d)", env.GetMaxFileSize()>>20),
	"alphanum":         "Поле может содержать только латиницу и арабские цифры",
	"audio":            "Доступны только форматы",
	"rueng":            "Поле может содержать только латиницу, кириллицу, арабские цифры и некоторые знаки ( ( ) ! . )",
}

var intTagsErrors = map[string]string{
	"min":              "Значение меньше минимального разрешенного",
	"max":              "Значение больше максимального разрешенного",
	"required":         "Поле обязательно",
	"excludesall":      "Поле содержит запрещенные символы",
	"required_without": "Поле обязательно, если нет другого",
	"exists":           "Значение поля не существует",
	"unique":           "Значение поля уже используется",
	"eqfield":          "Значение не совпадает",
	"photo":            "Доступны только форматы",
	"filesize":         fmt.Sprintf("Превышен допустимый размер файла (%d мб.)", env.GetMaxFileSize()>>20),
	"audio":            "Доступны только форматы",
	"rueng":            "Поле может содержать только латиницу, кириллицу, арабские цифры и некоторые знаки ( ( ) ! . )",
}

type TagsValidationError struct {
	Field     string
	FailedTag string
	FieldType string
	TagValue  string
}

type HttpValidationError struct {
	Field     string
	TextError string
}

func (e HttpValidationError) MarshalJSON() ([]byte, error) {
	custom := map[string]string{
		e.Field: e.TextError,
	}
	return json.Marshal(custom)
}

type ValidationResponseObject struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

func NewValidationError(err TagsValidationError) (fieldKey, textError string) {
	errStr := "Неизвестный тип"
	if strings.ToLower(err.FieldType) == "string" {
		errStr = stringTagsErrors[err.FailedTag]
	} else {
		errStr = intTagsErrors[err.FailedTag]
	}

	if err.TagValue != "" && !strings.Contains(err.TagValue, "_") {
		errStr += fmt.Sprintf(" (%v)", err.TagValue)
	}
	if err.FailedTag == "photo" {
		errStr += fmt.Sprintf(" %s", strings.Join(files.AvailablePhotoTypes(), ", "))
	}
	if err.FailedTag == "audio" {
		errStr += fmt.Sprintf(" %s", strings.Join(files.AvailableSongTypes(), ", "))
	}
	return strings.ToLower(err.Field), errStr
}
