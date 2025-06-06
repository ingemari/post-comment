package graph

import (
	"context"
	"post-comment-app/graph/model"
	"post-comment-app/storage"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreatePost(t *testing.T) {
	r := &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
	ctx := context.Background()

	// Test valid post creation
	post, err := r.Mutation().CreatePost(ctx, "Test Title", "Test Content", "John Doe", true)
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "Test Title", post.Title)
	assert.Equal(t, "John Doe", post.Author)

	// Test empty title
	_, err = r.Mutation().CreatePost(ctx, "", "Content", "Author", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title cannot be empty")
}

func TestAddComment(t *testing.T) {
	r := &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
	ctx := context.Background()

	// Create a post
	post := &model.Post{
		ID:            uuid.NewString(),
		Title:         "Test Post",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	err := r.storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	// Test valid comment
	comment, err := r.Mutation().AddComment(ctx, post.ID, nil, "Jane", "Great post!")
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, post.ID, comment.PostID)
	assert.Equal(t, "Jane", comment.Author)

	// Test empty text
	_, err = r.Mutation().AddComment(ctx, post.ID, nil, "Jane", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text cannot be empty")
}
