package service

import (
	"errors"
	"fmt"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/jinzhu/copier"
	"sync"
)

var ErrAlreadyExists = errors.New("This laptop Id already exists ")

type LaptopStore interface {
	// Save saves the laptop to the store
	Save(laptop *pb.Laptop) error
}

type InmemoryLaptopStore struct {
	mutex sync.Mutex
	data  map[string]*pb.Laptop
}

func NewMemoryLaptopStore() *InmemoryLaptopStore {
	return &InmemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (store *InmemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	// do a deep copy
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("can not copy laptop data: %w", err)
	}

	store.data[other.Id] = other
	return nil

}
