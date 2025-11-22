package model

type TeamMember struct {
	UserID   string `json:"user_id" valid:"required"`
	Username string `json:"username" valid:"required"`
	IsActive bool   `json:"is_active" valid:"required"`
}

type User struct {
	TeamMember
	TeamName string `json:"team_name" valid:"required"`
}
