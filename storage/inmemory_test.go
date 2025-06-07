package storage

import (
	"context"
	"fmt"
	"post-comment-app/graph/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStorage()

	t.Run("NewInMemoryStorage", func(t *testing.T) {
		s := NewInMemoryStorage()
		assert.NotNil(t, s)
		assert.Empty(t, s.posts)
		assert.Empty(t, s.comments)
	})

	t.Run("CreatePost", func(t *testing.T) {
		post := &model.Post{
			ID:            uuid.NewString(),
			Title:         "Test Post",
			Content:       "Content",
			Author:        "Author",
			AllowComments: true,
			CreatedAt:     time.Now().Format(time.RFC3339),
		}
		err := store.CreatePost(ctx, post)
		assert.NoError(t, err)

		// Проверяем, что пост добавлен
		posts, err := store.GetPosts(ctx)
		assert.NoError(t, err)
		assert.Len(t, posts, 1)
		assert.Equal(t, post.ID, posts[0].ID)
	})

	t.Run("GetPost", func(t *testing.T) {
		post := &model.Post{
			ID:            uuid.NewString(),
			Title:         "Another Post",
			Content:       "More Content",
			Author:        "Another Author",
			AllowComments: false,
			CreatedAt:     time.Now().Format(time.RFC3339),
		}
		_ = store.CreatePost(ctx, post)

		// Успешное получение
		retrieved, err := store.GetPost(ctx, post.ID)
		assert.NoError(t, err)
		assert.Equal(t, post.Title, retrieved.Title)

		// Несуществующий пост
		_, err = store.GetPost(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Equal(t, "post not found", err.Error())
	})

	t.Run("GetPosts", func(t *testing.T) {
		// Уже есть два поста из предыдущих тестов
		posts, err := store.GetPosts(ctx)
		assert.NoError(t, err)
		assert.Len(t, posts, 2)

		// Пустое хранилище
		emptyStore := NewInMemoryStorage()
		posts, err = emptyStore.GetPosts(ctx)
		assert.NoError(t, err)
		assert.Empty(t, posts)
	})

	t.Run("CreateComment", func(t *testing.T) {
		postID := uuid.NewString()
		_ = store.CreatePost(ctx, &model.Post{ID: postID, Title: "Post for Comment", Content: "Content", Author: "Author", AllowComments: true})

		comment := &model.Comment{
			PostID:    postID,
			Author:    "User",
			Text:      "Test comment",
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		// Успешное создание
		created, err := store.CreateComment(ctx, comment)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)

		// Проверка, что комментарий добавлен
		retrieved, err := store.GetComment(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, comment.Text, retrieved.Text)

		// Несуществующий пост
		invalidComment := &model.Comment{PostID: "non-existent-post", Author: "User", Text: "Invalid"}
		_, err = store.CreateComment(ctx, invalidComment)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post with ID non-existent-post not found")

		// Комментарий с parent_id
		parentComment := &model.Comment{PostID: postID, Author: "User", Text: "Parent comment"}
		parentCreated, _ := store.CreateComment(ctx, parentComment)
		reply := &model.Comment{PostID: postID, ParentID: &parentCreated.ID, Author: "User", Text: "Reply"}
		_, err = store.CreateComment(ctx, reply)
		assert.NoError(t, err)

		// Несуществующий parent_id
		invalidReply := &model.Comment{PostID: postID, ParentID: stringPtr("non-existent-parent"), Author: "User", Text: "Invalid reply"}
		_, err = store.CreateComment(ctx, invalidReply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent comment with ID non-existent-parent not found")
	})

	t.Run("GetComment", func(t *testing.T) {
		postID := uuid.NewString()
		_ = store.CreatePost(ctx, &model.Post{ID: postID, Title: "Post for Comment", Content: "Content", Author: "Author", AllowComments: true})
		comment := &model.Comment{PostID: postID, Author: "User", Text: "Another comment"}
		created, _ := store.CreateComment(ctx, comment)

		// Успешное получение
		retrieved, err := store.GetComment(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, comment.Text, retrieved.Text)

		// Несуществующий комментарий
		_, err = store.GetComment(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Equal(t, "comment not found", err.Error())
	})

	t.Run("GetCommentsByPostID", func(t *testing.T) {
		postID := uuid.NewString()
		_ = store.CreatePost(ctx, &model.Post{ID: postID, Title: "Post for Comments", Content: "Content", Author: "Author", AllowComments: true})

		// Добавляем три комментария
		for i := 0; i < 3; i++ {
			_, _ = store.CreateComment(ctx, &model.Comment{PostID: postID, Author: "User", Text: fmt.Sprintf("Comment %d", i)})
		}
		// Добавляем комментарий к другому посту
		otherPostID := uuid.NewString()
		_ = store.CreatePost(ctx, &model.Post{ID: otherPostID, Title: "Other Post", Content: "Content", Author: "Author", AllowComments: true})
		_, _ = store.CreateComment(ctx, &model.Comment{PostID: otherPostID, Author: "User", Text: "Other comment"})

		// Пагинация: первые два комментария
		comments, err := store.GetCommentsByPostID(ctx, postID, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, comments, 2)

		// Пагинация: третий комментарий
		comments, err = store.GetCommentsByPostID(ctx, postID, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, comments, 1)

		// Пагинация: за пределами
		comments, err = store.GetCommentsByPostID(ctx, postID, 2, 10)
		assert.NoError(t, err)
		assert.Empty(t, comments)

		// Несуществующий пост
		comments, err = store.GetCommentsByPostID(ctx, "non-existent-post", 2, 0)
		assert.NoError(t, err)
		assert.Empty(t, comments)
	})

	t.Run("GetRepliesByCommentID", func(t *testing.T) {
		postID := uuid.NewString()
		_ = store.CreatePost(ctx, &model.Post{ID: postID, Title: "Post for Replies", Content: "Content", Author: "Author", AllowComments: true})
		parentComment := &model.Comment{PostID: postID, Author: "User", Text: "Parent comment"}
		parentCreated, _ := store.CreateComment(ctx, parentComment)

		// Добавляем три ответа
		for i := 0; i < 3; i++ {
			_, _ = store.CreateComment(ctx, &model.Comment{PostID: postID, ParentID: &parentCreated.ID, Author: "User", Text: fmt.Sprintf("Reply %d", i)})
		}

		// Пагинация: первые два ответа
		replies, err := store.GetRepliesByCommentID(ctx, parentCreated.ID, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, replies, 2)

		// Пагинация: третий ответ
		replies, err = store.GetRepliesByCommentID(ctx, parentCreated.ID, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, replies, 1)

		// Пагинация: за пределами
		replies, err = store.GetRepliesByCommentID(ctx, parentCreated.ID, 2, 10)
		assert.NoError(t, err)
		assert.Empty(t, replies)

		// Несуществующий комментарий
		replies, err = store.GetRepliesByCommentID(ctx, "non-existent-comment", 2, 0)
		assert.NoError(t, err)
		assert.Empty(t, replies)
	})
}

// Вспомогательная функция для создания указателя на строку
func stringPtr(s string) *string {
	return &s
}
