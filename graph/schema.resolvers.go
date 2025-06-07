package graph

import (
	"context"
	"fmt"
	"log"
	"post-comment-app/graph/model"
	"time"

	"github.com/google/uuid"
)

func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, author string, allowComments bool) (*model.Post, error) {
	if title == "" || content == "" || author == "" {
		return nil, fmt.Errorf("title, content, and author must not be empty")
	}

	post := &model.Post{
		ID:            uuid.NewString(),
		Title:         title,
		Content:       content,
		Author:        author,
		AllowComments: allowComments,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	if err := r.storage.CreatePost(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (r *mutationResolver) AddComment(ctx context.Context, postID string, parentID *string, author string, text string) (*model.Comment, error) {
	log.Printf("Adding comment to postID: %s, parentID: %v", postID, parentID)

	if author == "" || text == "" {
		log.Println("Author and text must not be empty")
		return nil, fmt.Errorf("author and text must not be empty")
	}

	if len(text) > 2000 {
		log.Println("Comment too long")
		return nil, fmt.Errorf("comment too long")
	}

	post, err := r.storage.GetPost(ctx, postID)
	if err != nil {
		log.Printf("Error getting post: %v", err)
		return nil, fmt.Errorf("post not found")
	}
	if !post.AllowComments {
		log.Println("Comments are not allowed for this post")
		return nil, fmt.Errorf("comments are not allowed for this post")
	}

	if parentID != nil {
		_, err := r.storage.GetComment(ctx, *parentID)
		if err != nil {
			log.Printf("Error getting parent comment: %v", err)
			return nil, fmt.Errorf("parent comment not found")
		}
	}

	comment := &model.Comment{
		PostID:    postID,
		ParentID:  parentID,
		Author:    author,
		Text:      text,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	createdComment, err := r.storage.CreateComment(ctx, comment)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if subs, ok := r.subscribers[postID]; ok {
		for _, ch := range subs {
			select {
			case ch <- createdComment:
				log.Printf("Sent notification for comment %s to subscriber", createdComment.ID)
			default:
				log.Printf("Skipped notification for comment %s: channel full", createdComment.ID)
			}
		}
	}

	log.Printf("Comment %s created successfully", createdComment.ID)
	return createdComment, nil
}

func (r *queryResolver) Posts(ctx context.Context) ([]*model.Post, error) {
	return r.storage.GetPosts(ctx)
}

func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	log.Printf("Fetching post with ID: %s", id)
	post, err := r.storage.GetPost(ctx, id)
	if err != nil {
		log.Printf("Error fetching post: %v", err)
		return nil, err
	}
	log.Printf("Post %s fetched successfully", id)
	return post, nil
}

func (r *queryResolver) Comment(ctx context.Context, id string) (*model.Comment, error) {
	log.Printf("Fetching comment with ID: %s", id)
	comment, err := r.storage.GetComment(ctx, id)
	if err != nil {
		log.Printf("Error fetching comment: %v", err)
		return nil, err
	}
	log.Printf("Comment %s fetched successfully", id)
	return comment, nil
}

func (r *subscriptionResolver) CommentAdded(ctx context.Context, postID string) (<-chan *model.Comment, error) {
	log.Printf("New subscription for postID: %s", postID)
	ch := make(chan *model.Comment, 1)

	r.mu.Lock()
	r.subscribers[postID] = append(r.subscribers[postID], ch)
	r.mu.Unlock()

	go func() {
		<-ctx.Done()
		log.Printf("Cleaning up subscription for postID: %s", postID)
		r.mu.Lock()
		defer r.mu.Unlock()
		subs := r.subscribers[postID]
		for i, subscriber := range subs {
			if subscriber == ch {
				r.subscribers[postID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}()

	return ch, nil
}

func (r *Resolver) Mutation() MutationResolver         { return &mutationResolver{r} }
func (r *Resolver) Query() QueryResolver               { return &queryResolver{r} }
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
