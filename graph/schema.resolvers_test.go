package graph

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"post-comment-app/graph/model"
	"post-comment-app/storage"
)

func setupResolver() *Resolver {
	return &Resolver{
		storage:     storage.NewInMemoryStorage(),
		subscribers: make(map[string][]chan *model.Comment),
	}
}

func TestResolver(t *testing.T) {
	ctx := context.Background()

	t.Run("CreatePost", func(t *testing.T) {
		r := setupResolver()

		t.Run("ValidInput", func(t *testing.T) {
			post, err := r.Mutation().CreatePost(ctx, "Test Title", "Test Content", "John Doe", true)
			assert.NoError(t, err)
			assert.NotNil(t, post)
			assert.Equal(t, "Test Title", post.Title)
			assert.Equal(t, "Test Content", post.Content)
			assert.Equal(t, "John Doe", post.Author)
			assert.True(t, post.AllowComments)
			assert.NotEmpty(t, post.ID)
			assert.NotEmpty(t, post.CreatedAt)

			// Проверяем, что пост сохранен
			fetched, err := r.Query().Post(ctx, post.ID)
			assert.NoError(t, err)
			assert.Equal(t, post.ID, fetched.ID)
		})

		t.Run("EmptyInput", func(t *testing.T) {
			_, err := r.Mutation().CreatePost(ctx, "", "Content", "Author", true)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "title, content, and author must not be empty")

			_, err = r.Mutation().CreatePost(ctx, "Title", "", "Author", true)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "title, content, and author must not be empty")

			_, err = r.Mutation().CreatePost(ctx, "Title", "Content", "", true)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "title, content, and author must not be empty")
		})
	})

	t.Run("AddComment", func(t *testing.T) {
		r := setupResolver()
		postID := uuid.NewString()
		err := r.storage.CreatePost(ctx, &model.Post{
			ID:            postID,
			Title:         "Test Post",
			Content:       "Content",
			Author:        "Author",
			AllowComments: true,
			CreatedAt:     time.Now().Format(time.RFC3339),
		})
		assert.NoError(t, err)

		t.Run("ValidComment", func(t *testing.T) {
			comment, err := r.Mutation().AddComment(ctx, postID, nil, "Jane", "Great post!")
			assert.NoError(t, err)
			assert.NotNil(t, comment)
			assert.Equal(t, postID, comment.PostID)
			assert.Nil(t, comment.ParentID)
			assert.Equal(t, "Jane", comment.Author)
			assert.Equal(t, "Great post!", comment.Text)
			assert.NotEmpty(t, comment.ID)
			assert.NotEmpty(t, comment.CreatedAt)

			// Проверяем, что комментарий сохранен
			fetched, err := r.Query().Comment(ctx, comment.ID)
			assert.NoError(t, err)
			assert.Equal(t, comment.ID, fetched.ID)
		})

		t.Run("TooLongComment", func(t *testing.T) {
			tooLongText := string(make([]byte, 2001))
			_, err := r.Mutation().AddComment(ctx, postID, nil, "Jane", tooLongText)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "comment too long")
		})

		t.Run("InvalidPostID", func(t *testing.T) {
			_, err := r.Mutation().AddComment(ctx, "non-existent-post", nil, "Jane", "Comment")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "post not found")
		})

		t.Run("CommentsDisabled", func(t *testing.T) {
			disabledPostID := uuid.NewString()
			err := r.storage.CreatePost(ctx, &model.Post{
				ID:            disabledPostID,
				Title:         "No Comments",
				Content:       "Content",
				Author:        "Author",
				AllowComments: false,
				CreatedAt:     time.Now().Format(time.RFC3339),
			})
			assert.NoError(t, err)

			_, err = r.Mutation().AddComment(ctx, disabledPostID, nil, "Jane", "Comment")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "comments are not allowed")
		})

		t.Run("InvalidParentID", func(t *testing.T) {
			invalidParentID := "non-existent-parent"
			_, err := r.Mutation().AddComment(ctx, postID, &invalidParentID, "Jane", "Reply")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "parent comment not found")
		})

		t.Run("EmptyInput", func(t *testing.T) {
			_, err := r.Mutation().AddComment(ctx, postID, nil, "", "Comment")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "author and text must not be empty")

			_, err = r.Mutation().AddComment(ctx, postID, nil, "Jane", "")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "author and text must not be empty")
		})

		t.Run("ValidReply", func(t *testing.T) {
			parentComment, err := r.Mutation().AddComment(ctx, postID, nil, "Jane", "Parent comment")
			assert.NoError(t, err)

			reply, err := r.Mutation().AddComment(ctx, postID, &parentComment.ID, "Bob", "Reply")
			assert.NoError(t, err)
			assert.NotNil(t, reply)
			assert.Equal(t, parentComment.ID, *reply.ParentID)
		})
	})

	t.Run("Posts", func(t *testing.T) {
		r := setupResolver()

		// Создаем два поста
		post1, err := r.Mutation().CreatePost(ctx, "Post 1", "Content 1", "Author 1", true)
		assert.NoError(t, err)
		post2, err := r.Mutation().CreatePost(ctx, "Post 2", "Content 2", "Author 2", false)
		assert.NoError(t, err)

		posts, err := r.Query().Posts(ctx)
		assert.NoError(t, err)
		assert.Len(t, posts, 2)
		assert.Contains(t, []string{post1.ID, post2.ID}, posts[0].ID)
		assert.Contains(t, []string{post1.ID, post2.ID}, posts[1].ID)

		// Пустое хранилище
		emptyR := setupResolver()
		posts, err = emptyR.Query().Posts(ctx)
		assert.NoError(t, err)
		assert.Empty(t, posts)
	})

	t.Run("Post", func(t *testing.T) {
		r := setupResolver()

		post, err := r.Mutation().CreatePost(ctx, "Test Post", "Content", "Author", true)
		assert.NoError(t, err)

		fetchedPost, err := r.Query().Post(ctx, post.ID)
		assert.NoError(t, err)
		assert.Equal(t, post.ID, fetchedPost.ID)
		assert.Equal(t, "Test Post", fetchedPost.Title)
		assert.Equal(t, "Content", fetchedPost.Content)
		assert.Equal(t, "Author", fetchedPost.Author)
		assert.True(t, fetchedPost.AllowComments)

		_, err = r.Query().Post(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")
	})

	t.Run("Comment", func(t *testing.T) {
		r := setupResolver()
		postID := uuid.NewString()
		err := r.storage.CreatePost(ctx, &model.Post{
			ID:            postID,
			Title:         "Test Post",
			Content:       "Content",
			Author:        "Author",
			AllowComments: true,
			CreatedAt:     time.Now().Format(time.RFC3339),
		})
		assert.NoError(t, err)

		comment, err := r.Mutation().AddComment(ctx, postID, nil, "Jane", "Comment")
		assert.NoError(t, err)

		fetchedComment, err := r.Query().Comment(ctx, comment.ID)
		assert.NoError(t, err)
		assert.Equal(t, comment.ID, fetchedComment.ID)
		assert.Equal(t, "Jane", fetchedComment.Author)
		assert.Equal(t, "Comment", fetchedComment.Text)
		assert.Equal(t, postID, fetchedComment.PostID)

		_, err = r.Query().Comment(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "comment not found")
	})

	t.Run("CommentAddedSubscription", func(t *testing.T) {
		r := setupResolver()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		postID := uuid.NewString()
		err := r.storage.CreatePost(ctx, &model.Post{
			ID:            postID,
			Title:         "Test Post",
			Content:       "Content",
			Author:        "Author",
			AllowComments: true,
			CreatedAt:     time.Now().Format(time.RFC3339),
		})
		assert.NoError(t, err)

		// Подписываем два канала
		commentChan1, err := r.Subscription().CommentAdded(ctx, postID)
		assert.NoError(t, err)
		commentChan2, err := r.Subscription().CommentAdded(ctx, postID)
		assert.NoError(t, err)

		// Добавляем комментарий
		go func() {
			_, err := r.Mutation().AddComment(ctx, postID, nil, "Jane", "Subscribed comment")
			assert.NoError(t, err)
		}()

		// Проверяем оба канала
		select {
		case comment := <-commentChan1:
			assert.Equal(t, postID, comment.PostID)
			assert.Equal(t, "Jane", comment.Author)
			assert.Equal(t, "Subscribed comment", comment.Text)
		case <-time.After(time.Second):
			t.Fatal("Did not receive comment in channel 1")
		}

		select {
		case comment := <-commentChan2:
			assert.Equal(t, postID, comment.PostID)
			assert.Equal(t, "Jane", comment.Author)
			assert.Equal(t, "Subscribed comment", comment.Text)
		case <-time.After(time.Second):
			t.Fatal("Did not receive comment in channel 2")
		}

		// Проверяем отмену подписки
		cancel()
		time.Sleep(100 * time.Millisecond) // Даем время на очистку
		r.mu.RLock()
		assert.Empty(t, r.subscribers[postID])
		r.mu.RUnlock()
	})

	t.Run("GetCommentsByPostID", func(t *testing.T) {
		r := setupResolver()
		postID := uuid.NewString()
		err := r.storage.CreatePost(ctx, &model.Post{
			ID:            postID,
			Title:         "Test Post",
			Content:       "Content",
			Author:        "Author",
			AllowComments: true,
			CreatedAt:     time.Now().Format(time.RFC3339),
		})
		assert.NoError(t, err)

		// Добавляем 10 комментариев
		for i := 0; i < 10; i++ {
			_, err := r.Mutation().AddComment(ctx, postID, nil, "User", fmt.Sprintf("Comment %d", i))
			assert.NoError(t, err)
		}

		// Первые 5 комментариев
		comments, err := r.storage.GetCommentsByPostID(ctx, postID, 5, 0)
		assert.NoError(t, err)
		assert.Len(t, comments, 5)
		for i, c := range comments {
			assert.Equal(t, fmt.Sprintf("Comment %d", i), c.Text) // Порядок добавления
		}

		// Следующие 5 комментариев
		comments, err = r.storage.GetCommentsByPostID(ctx, postID, 5, 5)
		assert.NoError(t, err)
		assert.Len(t, comments, 5)
		for i, c := range comments {
			assert.Equal(t, fmt.Sprintf("Comment %d", i+5), c.Text) // Порядок добавления
		}

		// За пределами
		comments, err = r.storage.GetCommentsByPostID(ctx, postID, 5, 10)
		assert.NoError(t, err)
		assert.Empty(t, comments)

		// Несуществующий пост
		comments, err = r.storage.GetCommentsByPostID(ctx, "non-existent-post", 5, 0)
		assert.NoError(t, err)
		assert.Empty(t, comments)
	})

	t.Run("ConcurrentCommentCreation", func(t *testing.T) {
		r := setupResolver()
		postID := uuid.NewString()
		err := r.storage.CreatePost(ctx, &model.Post{
			ID:            postID,
			Title:         "Test Post",
			Content:       "Content",
			Author:        "Author",
			AllowComments: true,
			CreatedAt:     time.Now().Format(time.RFC3339),
		})
		assert.NoError(t, err)

		var wg sync.WaitGroup
		commentCount := 100
		commentIDs := make(map[string]bool)
		var mu sync.Mutex

		for i := 0; i < commentCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				comment, err := r.Mutation().AddComment(ctx, postID, nil, "User", fmt.Sprintf("Comment %d", i))
				assert.NoError(t, err)
				mu.Lock()
				commentIDs[comment.ID] = true
				mu.Unlock()
			}(i)
		}
		wg.Wait()

		comments, err := r.storage.GetCommentsByPostID(ctx, postID, commentCount, 0)
		assert.NoError(t, err)
		assert.Len(t, comments, commentCount)
		assert.Len(t, commentIDs, commentCount) // Проверяем уникальность ID
	})
}
