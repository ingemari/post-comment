package graph

import (
	"post-comment-app/graph/model"
	"sync"
)

type Resolver struct {
	Posts            []*model.Post
	Comments         []*model.Comment
	postIDCounter    int
	commentIDCounter int
	mu               sync.Mutex
}

func NewResolver() *Resolver {
	return &Resolver{
		Posts:            make([]*model.Post, 0),
		Comments:         make([]*model.Comment, 0),
		postIDCounter:    1,
		commentIDCounter: 1,
	}
}

func (r *Resolver) buildCommentTree(postID string, parentID *string, all []*model.Comment) []*model.Comment {
	var tree []*model.Comment
	for _, comment := range all {
		if comment.PostID == postID && ((comment.ParentID == nil && parentID == nil) ||
			(comment.ParentID != nil && parentID != nil && *comment.ParentID == *parentID)) {
			children := r.buildCommentTree(postID, &comment.ID, all)
			comment.Children = children
			tree = append(tree, comment)
		}
	}
	return tree
}
