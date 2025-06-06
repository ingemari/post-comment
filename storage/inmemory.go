package storage

import (
	"context"
	"fmt"
	"post-comment-app/graph/model"
	"sync"
)

type InMemoryStorage struct {
	posts    []*model.Post
	comments []*model.Comment
	mu       sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		posts:    []*model.Post{},
		comments: []*model.Comment{},
	}
}

func (s *InMemoryStorage) CreatePost(ctx context.Context, post *model.Post) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.posts = append(s.posts, post)
	return nil
}

func (s *InMemoryStorage) GetPosts(ctx context.Context) ([]*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.posts, nil
}

func (s *InMemoryStorage) GetPost(ctx context.Context, id string) (*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.posts {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("post not found")
}

func (s *InMemoryStorage) CreateComment(ctx context.Context, comment *model.Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.comments = append(s.comments, comment)
	return nil
}

func (s *InMemoryStorage) GetComment(ctx context.Context, id string) (*model.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.comments {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("comment not found")
}

func (s *InMemoryStorage) GetCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var comments []*model.Comment
	for _, c := range s.comments {
		if c.PostID == postID && c.ParentID == nil {
			comments = append(comments, c)
		}
	}
	start := offset
	end := offset + limit
	if start > len(comments) {
		return []*model.Comment{}, nil
	}
	if end > len(comments) {
		end = len(comments)
	}
	return comments[start:end], nil
}

func (s *InMemoryStorage) GetRepliesByCommentID(ctx context.Context, commentID string, limit, offset int) ([]*model.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var replies []*model.Comment
	for _, c := range s.comments {
		if c.ParentID != nil && *c.ParentID == commentID {
			replies = append(replies, c)
		}
	}
	start := offset
	end := offset + limit
	if start > len(replies) {
		return []*model.Comment{}, nil
	}
	if end > len(replies) {
		end = len(replies)
	}
	return replies[start:end], nil
}
