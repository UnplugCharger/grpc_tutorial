package service

import "sync"

type RatingStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

type Rating struct {
	Count uint32
	Sum   float64
}

type InMemoryRatingStore struct {
	mutex   sync.RWMutex
	ratings map[string]*Rating
}

func NewInMemoryRatingStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		ratings: make(map[string]*Rating),
	}
}

func (store *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	rating := store.ratings[laptopID]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.ratings[laptopID] = rating

	return rating, nil
}
