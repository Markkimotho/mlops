package feature

import (
	"errors"
	"fmt"
	"sync"
)

type Request struct {
	FeatureService string           `json:"feature_service"`
	Entities       []map[string]any `json:"entities"`
}

type Result struct {
	Values   map[string]any    `json:"values"`
	Statuses map[string]string `json:"statuses"`
}

type Response struct {
	Results []Result `json:"results"`
}

type Store interface {
	Get(service, entityKey string) (map[string]any, bool)
	Put(service, entityKey string, values map[string]any)
}

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]map[string]map[string]any
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]map[string]map[string]any)}
}

func (s *MemoryStore) Get(service, key string) (map[string]any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values, ok := s.data[service][key]
	if !ok {
		return nil, false
	}
	copy := make(map[string]any, len(values))
	for k, v := range values {
		copy[k] = v
	}
	return copy, true
}

func (s *MemoryStore) Put(service, key string, values map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data[service] == nil {
		s.data[service] = make(map[string]map[string]any)
	}
	s.data[service][key] = values
}

func Lookup(store Store, request Request) (Response, error) {
	if request.FeatureService == "" || len(request.Entities) == 0 {
		return Response{}, errors.New("feature_service and at least one entity are required")
	}
	response := Response{Results: make([]Result, 0, len(request.Entities))}
	for _, entity := range request.Entities {
		key, err := EntityKey(entity)
		if err != nil {
			return Response{}, err
		}
		values, found := store.Get(request.FeatureService, key)
		statuses := make(map[string]string)
		if !found {
			values = map[string]any{}
			statuses["entity"] = "NOT_FOUND"
		} else {
			for name := range values {
				statuses[name] = "PRESENT"
			}
		}
		response.Results = append(response.Results, Result{Values: values, Statuses: statuses})
	}
	return response, nil
}

func EntityKey(entity map[string]any) (string, error) {
	if len(entity) != 1 {
		return "", errors.New("each entity must contain exactly one identifier")
	}
	for key, value := range entity {
		if key == "" || value == nil {
			return "", errors.New("entity identifier cannot be empty")
		}
		return fmt.Sprintf("%s=%v", key, value), nil
	}
	return "", errors.New("entity cannot be empty")
}
