package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/jinzhu/copier"
	"log"
	"sync"
)

var ErrAlreadyExists = errors.New("This laptop Id already exists ")

type LaptopStore interface {
	// Save saves the laptop to the store
	Save(laptop *pb.Laptop) error
	// Find finds a laptop by ID
	Find(id string) (*pb.Laptop, error)
	// Search searches for laptops with filter, returns one by one via the found function
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

type InmemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InmemoryLaptopStore {
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
	other, err := doDeepCopy(laptop)
	if err != nil {
		return fmt.Errorf("can not copy laptop data: %w", err)
	}

	store.data[other.Id] = other

	return nil

}

func (store *InmemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		log.Printf("laptop with id %s not found", id)
		return nil, nil
	}

	return doDeepCopy(laptop)
}

func (store *InmemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {

		//// simulate some heavy processing
		//time.Sleep(time.Second)

		// check if the context is canceled
		if errors.Is(context.Canceled, ctx.Err()) {
			log.Print("request is canceled")
			return errors.New("request is canceled")
		}

		// check if the context is DeadLineExceeded
		if errors.Is(context.DeadlineExceeded, ctx.Err()) {
			log.Print("deadline is exceeded")
			return errors.New("deadline is exceeded")
		}

		if isQualified(filter, laptop) {
			other, err := doDeepCopy(laptop)

			if err != nil {
				return errors.New("can not process laptop")
			}

			err = found(other)
			if err != nil {
				return errors.New("can not process laptop")
			}
		}

	}

	return nil
}

func doDeepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	// do a deep copy
	other := &pb.Laptop{}

	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("can not copy laptop data: %w", err)
	}
	return other, nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPrice() > filter.GetMaxPrice() {
		return false
	}

	if laptop.GetPrice() < filter.GetMinPrice() {
		return false
	}

	if laptop.GetCpu().GetNumberOfCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetMemory()) < toBit(filter.GetMinMemory()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()
	switch memory.GetUnit() {
	case pb.Memory_BYTES:
		return uint64(value)
	case pb.Memory_KB:
		return uint64(value) * 1e3
	case pb.Memory_MB:
		return uint64(value) * 1e6
	case pb.Memory_GB:
		return uint64(value) * 1e9
	case pb.Memory_TB:
		return uint64(value) * 1e12
	default:
		return 0
	}

}
