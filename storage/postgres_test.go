package storage

import (
	"context"
	"os"
	"post-comment-app/graph/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStorage(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	assert.NoError(t, err)
	defer pool.Close()

	_, err = pool.Exec(context.Background(), "TRUNCATE TABLE comments, posts RESTART IDENTITY CASCADE")
	assert.NoError(t, err)

	store, err := NewPostgresStorage(dsn)
	assert.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	post := &model.Post{
		ID:            uuid.NewString(),
		Title:         "Test Post",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	assert.NoError(t, store.CreatePost(ctx, post))

	retrieved, err := store.GetPost(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, post.Title, retrieved.Title)

	comment := &model.Comment{
		PostID:    post.ID,
		Author:    "User",
		Text:      "Test comment",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	createdComment, err := store.CreateComment(ctx, comment)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdComment.ID)

	comments, err := store.GetCommentsByPostID(ctx, post.ID, 1, 0)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "Test comment", comments[0].Text)
}
