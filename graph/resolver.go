package graph

import (
	"post-comment-app/graph/model"
	"post-comment-app/storage"
	"sync"
)

type Resolver struct {
	storage     storage.Storage
	subscribers map[string][]chan *model.Comment
	mu          sync.RWMutex
}

func NewResolver(store storage.Storage) *Resolver {
	return &Resolver{
		storage:     store,
		subscribers: make(map[string][]chan *model.Comment),
	}
}
