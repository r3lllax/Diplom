package service

import (
	"GIN/internal/files"
	"GIN/repositories/repository"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

func uploadRequestFile(file *multipart.FileHeader, ctx *gin.Context, repo *repository.StorageRepository) (string, error) {
	dst := ""
	if file != nil {
		dst = files.GenerateFilePath(file)
		err := repo.UploadFile(file, dst, ctx)
		if err != nil {
			return "", err
		}
	}
	if len(dst) != 0 {
		dst = dst[1:]
	}
	return dst, nil
}
