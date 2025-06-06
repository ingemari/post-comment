package graph

import (
	"post-comment-app/graph/model"
	"post-comment-app/storage"
	"sync"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	storage     storage.Storage
	subscribers map[string][]chan *model.Comment
	mu          sync.RWMutex
}

func NewResolver() *Resolver {
	return &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
}
