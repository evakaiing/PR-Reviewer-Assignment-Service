package model

type Team struct {
	TeamName string        `json:"team_name" valid:"required"`
	Members  []*TeamMember `json:"members" valid:"required"`
}
