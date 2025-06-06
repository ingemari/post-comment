package graph

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"post-comment-app/graph/model"
	"post-comment-app/storage"
	"sync"
	"testing"
	"time"
)

func TestCreatePost(t *testing.T) {
	r := &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
	ctx := context.Background()

	// Тест создания поста с валидными данными
	post, err := r.Mutation().CreatePost(ctx, "Test Title", "Test Content", "John Doe", true)
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "Test Title", post.Title)
	assert.Equal(t, "Test Content", post.Content)
	assert.Equal(t, "John Doe", post.Author)
	assert.True(t, post.AllowComments)
	assert.NotEmpty(t, post.ID)
	assert.NotEmpty(t, post.CreatedAt)
}

func TestAddComment(t *testing.T) {
	r := &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
	ctx := context.Background()

	// Создаём пост
	postID := uuid.NewString()
	post := &model.Post{
		ID:            postID,
		Title:         "Test Post",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	err := r.storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	// Тест добавления комментария
	comment, err := r.Mutation().AddComment(ctx, postID, nil, "Jane", "Great post!")
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, postID, comment.PostID)
	assert.Nil(t, comment.ParentID)
	assert.Equal(t, "Jane", comment.Author)
	assert.Equal(t, "Great post!", comment.Text)
	assert.NotEmpty(t, comment.ID)
	assert.NotEmpty(t, comment.CreatedAt)

	// Тест слишком длинного комментария
	tooLongText := string(make([]byte, 2001)) // 2001 символ
	_, err = r.Mutation().AddComment(ctx, postID, nil, "Jane", tooLongText)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comment too long")
}

func TestPosts(t *testing.T) {
	r := &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
	ctx := context.Background()

	// Создаём два поста
	post1 := &model.Post{ID: uuid.NewString(), Title: "Post 1", Content: "Content 1", Author: "Author 1", AllowComments: true, CreatedAt: time.Now().Format(time.RFC3339)}
	post2 := &model.Post{ID: uuid.NewString(), Title: "Post 2", Content: "Content 2", Author: "Author 2", AllowComments: false, CreatedAt: time.Now().Format(time.RFC3339)}
	err := r.storage.CreatePost(ctx, post1)
	assert.NoError(t, err)
	err = r.storage.CreatePost(ctx, post2)
	assert.NoError(t, err)

	// Проверяем, что возвращаются оба поста
	posts, err := r.Query().Posts(ctx)
	assert.NoError(t, err)
	assert.Len(t, posts, 2)
}

func TestPost(t *testing.T) {
	r := &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
	ctx := context.Background()

	// Создаём пост
	postID := uuid.NewString()
	post := &model.Post{ID: postID, Title: "Test Post", Content: "Content", Author: "Author", AllowComments: true, CreatedAt: time.Now().Format(time.RFC3339)}
	err := r.storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	// Проверяем получение поста
	fetchedPost, err := r.Query().Post(ctx, postID)
	assert.NoError(t, err)
	assert.Equal(t, postID, fetchedPost.ID)
	assert.Equal(t, "Test Post", fetchedPost.Title)

	// Проверяем несуществующий пост
	_, err = r.Query().Post(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestComment(t *testing.T) {
	r := &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
	ctx := context.Background()

	// Создаём пост и комментарий
	postID := uuid.NewString()
	commentID := uuid.NewString()
	comment := &model.Comment{ID: commentID, PostID: postID, Author: "Jane", Text: "Comment", CreatedAt: time.Now().Format(time.RFC3339)}
	err := r.storage.CreateComment(ctx, comment)
	assert.NoError(t, err)

	// Проверяем получение комментария
	fetchedComment, err := r.Query().Comment(ctx, commentID)
	assert.NoError(t, err)
	assert.Equal(t, commentID, fetchedComment.ID)
	assert.Equal(t, "Jane", fetchedComment.Author)

	// Проверяем несуществующий комментарий
	_, err = r.Query().Comment(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestCommentAddedSubscription(t *testing.T) {
	r := &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаём пост
	postID := uuid.NewString()
	post := &model.Post{ID: postID, Title: "Test Post", Content: "Content", Author: "Author", AllowComments: true, CreatedAt: time.Now().Format(time.RFC3339)}
	err := r.storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	// Подписываемся
	commentChan, err := r.Subscription().CommentAdded(ctx, postID)
	assert.NoError(t, err)

	// Добавляем комментарий в отдельной горутине
	go func() {
		_, err := r.Mutation().AddComment(ctx, postID, nil, "Jane", "Subscribed comment")
		assert.NoError(t, err)
	}()

	// Проверяем, что комментарий получен через подписку
	select {
	case comment := <-commentChan:
		assert.Equal(t, postID, comment.PostID)
		assert.Equal(t, "Jane", comment.Author)
		assert.Equal(t, "Subscribed comment", comment.Text)
	case <-time.After(time.Second):
		t.Fatal("Did not receive comment in time")
	}
}
func TestGetCommentsByPostID(t *testing.T) {
	r := &Resolver{storage: storage.NewInMemoryStorage()}
	ctx := context.Background()
	postID := uuid.NewString()
	post := &model.Post{ID: postID, AllowComments: true}
	r.storage.CreatePost(ctx, post)
	for i := 0; i < 10; i++ {
		r.storage.CreateComment(ctx, &model.Comment{ID: uuid.NewString(), PostID: postID, Text: fmt.Sprintf("Comment %d", i)})
	}
	comments, err := r.storage.GetCommentsByPostID(ctx, postID, 5, 0)
	assert.NoError(t, err)
	assert.Len(t, comments, 5)
	comments, err = r.storage.GetCommentsByPostID(ctx, postID, 5, 5)
	assert.NoError(t, err)
	assert.Len(t, comments, 5)
}
func TestConcurrentCommentCreation(t *testing.T) {
	r := &Resolver{storage: storage.NewInMemoryStorage()}
	ctx := context.Background()
	postID := uuid.NewString()
	r.storage.CreatePost(ctx, &model.Post{ID: postID, AllowComments: true})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r.Mutation().AddComment(ctx, postID, nil, "User", fmt.Sprintf("Comment %d", i))
		}(i)
	}
	wg.Wait()
	comments, _ := r.storage.GetCommentsByPostID(ctx, postID, 100, 0)
	assert.Len(t, comments, 100)
}
