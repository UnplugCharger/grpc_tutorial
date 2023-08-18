package service

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"sync"
)

type ImageStore interface {
	Save(laptopId string, imageType string, imageData []byte) (string, error)
}

type ImageInfo struct {
	LaptopId string
	Type     string
	Path     string
}

type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*ImageInfo
}

func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		images:      make(map[string]*ImageInfo),
	}
}

func (store *DiskImageStore) Save(laptopId string, imageType string, imageData []byte) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}
	imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder, imageID, imageType)
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}
	// compare with imageData.WriteTo(file)
	_, err = file.Write(imageData)
	if err != nil {
		return "", fmt.Errorf("cannot write image data to file: %w", err)
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.images[imageID.String()] = &ImageInfo{
		LaptopId: laptopId,
		Type:     imageType,
		Path:     imagePath,
	}

	return imageID.String(), nil
}
