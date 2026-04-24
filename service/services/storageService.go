package service

import (
	"GIN/internal/files"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type StorageService struct {
	baseService
}

func NewStorageService(BService *baseService) *StorageService {
	return &StorageService{
		baseService: *BService,
	}
}

func (s *StorageService) SaveFile(file *multipart.FileHeader, ctx *gin.Context) error {
	dst := files.GenerateFilePath(file)
	err := s.baseService.repositories.StorageRepository.UploadFile(file, dst, ctx)
	if err != nil {
		return err
	}
	return nil

}
