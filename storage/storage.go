package storage

import (
	"context"
	"post-comment-app/graph/model"
)

type Storage interface {
	CreatePost(ctx context.Context, post *model.Post) error
	GetPosts(ctx context.Context) ([]*model.Post, error)
	GetPost(ctx context.Context, id string) (*model.Post, error)
	CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	GetComment(ctx context.Context, id string) (*model.Comment, error)
	GetCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error)
	GetRepliesByCommentID(ctx context.Context, commentID string, limit, offset int) ([]*model.Comment, error)
}
