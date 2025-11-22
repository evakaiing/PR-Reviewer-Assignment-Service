package model

import (
	"time"
)

type PullRequest struct {
	PullRequestShort
	AssignedReviewers []string   `json:"assigned_reviewers" valid:"required"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id" valid:"required"`
	PullRequestName string `json:"pull_request_name" valid:"required"`
	AuthorID        string `json:"author_id" valid:"required"`
	Status          string `json:"status" valid:"in(OPEN|MERGED)"`
}

type PullRequestPayload struct {
	PullRequestID   string `json:"pull_request_id" valid:"required"`
	PullRequestName string `json:"pull_request_name" valid:"required"`
	AuthorID        string `json:"author_id" valid:"required"`
}
