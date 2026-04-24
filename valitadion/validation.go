package valitadion

import (
	"GIN/internal/files"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var validCharsRegex = regexp.MustCompile("^[\\p{L}\\p{N}\\s!\\(\\)]+$")

type StructValidator struct {
	db *pgxpool.Pool
}

func NewValidator(dbase *pgxpool.Pool) *StructValidator {
	return &StructValidator{db: dbase}
}

func (v *StructValidator) ValidateWithErrorsByte(obj interface{}) []byte {
	validate := validator.New()
	validate.RegisterValidation("exists", v.exists)
	validate.RegisterValidation("unique", v.unique)
	validate.RegisterValidation("photo", v.photo)
	validate.RegisterValidation("audio", v.audio)
	validate.RegisterValidation("filesize", v.filesize)
	validate.RegisterValidation("rueng", v.rusEngNum)

	err := validate.Struct(obj)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		errsMap := make(map[string]string, len(validationErrors))
		for _, err := range validationErrors {
			fieldKey, textError := NewValidationError(TagsValidationError{
				Field:     err.Field(),
				FailedTag: err.ActualTag(),
				FieldType: err.Kind().String(),
				TagValue:  err.Param(),
			})
			errsMap[fieldKey] = textError
		}

		respErr := ValidationResponseObject{
			Message: fmt.Sprintf("Ошибки валидации (%d шт.)", len(errsMap)),
			Errors:  errsMap,
		}

		indent, err := json.Marshal(respErr)
		if err != nil {
			log.Print("Json marshal error:", err)
			return []byte{}
		}
		return indent
	}
	return []byte{}
}

func (v *StructValidator) ValidateWithErrorsObj(obj interface{}) ValidationResponseObject {
	validate := validator.New()
	validate.RegisterValidation("exists", v.exists)
	validate.RegisterValidation("unique", v.unique)
	validate.RegisterValidation("photo", v.photo)
	validate.RegisterValidation("audio", v.audio)
	validate.RegisterValidation("filesize", v.filesize)
	validate.RegisterValidation("rueng", v.rusEngNum)
	err := validate.Struct(obj)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)

		errsMap := make(map[string]string, len(validationErrors))
		for _, err := range validationErrors {
			fieldKey, textError := NewValidationError(TagsValidationError{
				Field:     err.Field(),
				FailedTag: err.ActualTag(),
				FieldType: err.Kind().String(),
				TagValue:  err.Param(),
			})
			errsMap[fieldKey] = textError
		}

		respErr := ValidationResponseObject{
			Message: fmt.Sprintf("Ошибки валидации (%d шт.)", len(errsMap)),
			Errors:  errsMap,
		}
		return respErr

	}
	return ValidationResponseObject{}
}

func (v *StructValidator) exists(fl validator.FieldLevel) bool {
	tokens := strings.Split(fl.Param(), "_")

	table := tokens[0]
	if len(table) == 0 {
		return false
	}
	field := strings.ToLower(fmt.Sprintf("%v", fl.FieldName()))
	if len(tokens) > 1 {
		field = tokens[1]
	}

	value := strings.ToLower(fmt.Sprintf("%v", fl.Field()))

	query := fmt.Sprintf(`
	    SELECT EXISTS (
	        SELECT 1
	        FROM %s
	        WHERE %s = $1
	    )`,
		pgx.Identifier{table}.Sanitize(),
		pgx.Identifier{field}.Sanitize(),
	)
	var exists bool
	err := v.db.QueryRow(context.Background(), query, value).Scan(&exists)
	if err != nil {
		log.Println("EXIST CHECK ERROR:", err)
		return false
	}
	return exists
}

func (v *StructValidator) unique(fl validator.FieldLevel) bool {
	return !v.exists(fl)
}

func (v *StructValidator) photo(fl validator.FieldLevel) bool {
	file := fl.Field().Interface().(multipart.FileHeader)
	availableExts := files.AvailablePhotoTypes()
	fileExt := filepath.Ext(file.Filename)
	return slices.Contains(availableExts, fileExt)
}

func (v *StructValidator) audio(fl validator.FieldLevel) bool {
	file := fl.Field().Interface().(multipart.FileHeader)
	availableExts := files.AvailableSongTypes()
	fileExt := filepath.Ext(file.Filename)
	return slices.Contains(availableExts, fileExt)
}

func (v *StructValidator) filesize(fl validator.FieldLevel) bool {
	file := fl.Field().Interface().(multipart.FileHeader)

	return files.AvailableSizeCheck(&file)
}

func (v *StructValidator) rusEngNum(fl validator.FieldLevel) bool {
	return validCharsRegex.MatchString(fl.Field().String())
}
