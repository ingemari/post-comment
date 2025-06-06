package graph

import (
	"post-comment-app/graph/model"
	"post-comment-app/storage"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	storage     storage.Storage
	posts       []*model.Post
	comments    []*model.Comment
	subscribers map[string][]chan *model.Comment
}

func NewResolver() *Resolver {
	return &Resolver{
		storage:     storage.NewInMemoryStorage(),
		posts:       []*model.Post{},
		comments:    []*model.Comment{},
		subscribers: make(map[string][]chan *model.Comment),
	}
}
