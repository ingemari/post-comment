package storage

import (
	"context"
	"fmt"
	"post-comment-app/graph/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) CreatePost(ctx context.Context, post *model.Post) error {
	query := `INSERT INTO posts (id, title, content, author, allow_comments, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.pool.Exec(ctx, query, post.ID, post.Title, post.Content, post.Author, post.AllowComments, post.CreatedAt)
	return err
}

func (s *PostgresStorage) GetPost(ctx context.Context, id string) (*model.Post, error) {
	query := `SELECT id, title, content, author, allow_comments, created_at FROM posts WHERE id = $1`
	post := &model.Post{}
	err := s.pool.QueryRow(ctx, query, id).Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("post not found")
	}
	return post, err
}

func (s *PostgresStorage) GetPosts(ctx context.Context) ([]*model.Post, error) {
	query := `SELECT id, title, content, author, allow_comments, created_at FROM posts ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (s *PostgresStorage) CreateComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	query := `INSERT INTO comments (post_id, parent_id, author, text, created_at)
VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int64
	var parentID *int64
	if comment.ParentID != nil {
		var pid int64
		if _, err := fmt.Sscanf(*comment.ParentID, "%d", &pid); err != nil {
			return nil, fmt.Errorf("invalid parent_id format")
		}
		parentID = &pid
	}
	err := s.pool.QueryRow(ctx, query, comment.PostID, parentID, comment.Author, comment.Text, comment.CreatedAt).Scan(&id)
	if err != nil {
		return nil, err
	}
	comment.ID = fmt.Sprintf("%d", id)
	return comment, nil
}

func (s *PostgresStorage) GetComment(ctx context.Context, id string) (*model.Comment, error) {
	query := `SELECT id, post_id, parent_id, author, text, created_at FROM comments WHERE id = $1`
	comment := &model.Comment{}
	var parentID *int64
	err := s.pool.QueryRow(ctx, query, id).Scan(&comment.ID, &comment.PostID, &parentID, &comment.Author, &comment.Text, &comment.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("comment not found")
	}
	if parentID != nil {
		parentIDStr := fmt.Sprintf("%d", *parentID)
		comment.ParentID = &parentIDStr
	}
	return comment, err
}

func (s *PostgresStorage) GetCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	query := `SELECT id, post_id, parent_id, author, text, created_at FROM comments
WHERE post_id = $1 AND parent_id IS NULL
ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.pool.Query(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		comment := &model.Comment{}
		var parentID *int64
		if err := rows.Scan(&comment.ID, &comment.PostID, &parentID, &comment.Author, &comment.Text, &comment.CreatedAt); err != nil {
			return nil, err
		}
		if parentID != nil {
			parentIDStr := fmt.Sprintf("%d", *parentID)
			comment.ParentID = &parentIDStr
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (s *PostgresStorage) GetRepliesByCommentID(ctx context.Context, commentID string, limit, offset int) ([]*model.Comment, error) {
	query := `SELECT id, post_id, parent_id, author, text, created_at FROM comments
WHERE parent_id = $1
ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.pool.Query(ctx, query, commentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		comment := &model.Comment{}
		var parentID *int64
		if err := rows.Scan(&comment.ID, &comment.PostID, &parentID, &comment.Author, &comment.Text, &comment.CreatedAt); err != nil {
			return nil, err
		}
		if parentID != nil {
			parentIDStr := fmt.Sprintf("%d", *parentID)
			comment.ParentID = &parentIDStr
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (s *PostgresStorage) Close() {
	s.pool.Close()
}
