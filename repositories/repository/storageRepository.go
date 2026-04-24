package repository

import (
	errs "GIN/errors"
	"log"
	"mime/multipart"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

type StorageRepository struct {
	mtx *sync.RWMutex
}

func NewStorageRepository() *StorageRepository {
	return &StorageRepository{
		mtx: &sync.RWMutex{},
	}
}

func (r *StorageRepository) UploadFile(file *multipart.FileHeader, dst string, c *gin.Context) error {
	err := c.SaveUploadedFile(file, dst)
	if err != nil {
		log.Println("SaveUploadedFile error:", err)
		return errs.ServerError()
	}
	return nil
}
func (r *StorageRepository) DeleteFile(name string) error {

	r.mtx.RLock()
	defer r.mtx.RUnlock()

	err := os.Remove(name)
	if err != nil {
		log.Println("STORAGE REPOSITORY DELETE FILE ERROR:", err)
		return errs.ServerError()
	}

	return nil

}
