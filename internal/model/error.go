package model

import "errors"

var (
	ErrTeamExists  = errors.New("team already exists")
	ErrPrExists    = errors.New("PR already exists")
	ErrPrMerged    = errors.New("PR already merged")
	ErrNotAssigned = errors.New("not assigned")
	ErrNoCandidate = errors.New("no candidate")
	ErrNotFound    = errors.New("not found")
	ErrInvalidInput  = errors.New("invalid input")
    ErrMissingParam  = errors.New("missing required parameter") 
)

type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}
