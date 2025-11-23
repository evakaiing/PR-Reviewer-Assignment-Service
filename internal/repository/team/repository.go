package team

import (
	"context"
	"database/sql"
	"fmt"
	model "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/model"
	def "github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository"
)

var _ def.TeamRepository = (*repository)(nil)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Add(ctx context.Context, team *model.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %v", err)
	}
	defer tx.Rollback()

	createTeamQuery := `
        INSERT INTO teams (team_name)
        VALUES ($1)
        ON CONFLICT (team_name) DO NOTHING
        RETURNING team_id
    `

	var teamID int
	err = tx.QueryRowContext(ctx, createTeamQuery, team.TeamName).Scan(&teamID)
	if err == sql.ErrNoRows {
		return model.ErrTeamExists
	}

	if err != nil {
		return fmt.Errorf("failed to create team: %v", err)
	}

	updateOrInsertQuery :=  `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, member := range team.Members {
		_, err := tx.ExecContext(ctx, updateOrInsertQuery, 
            member.UserID,
            member.Username,
            team.TeamName,
            member.IsActive)
		if err != nil {
			return fmt.Errorf("failed to add member to team: %v", err)
		}
	}

	return tx.Commit()
}

func (r *repository) Get(ctx context.Context, teamName string) (*model.Team, error) {
	query := `
		SELECT 
			user_id,
			username,
			is_active
		FROM users
		WHERE team_name = $1
	`
	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	defer rows.Close()

	members := make([]*model.TeamMember, 0)
	for rows.Next() {
		m := &model.TeamMember{}
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		members = append(members, m)
	}

	if len(members) == 0 {
		return nil, model.ErrNotFound
	}

	team := &model.Team{
		TeamName: teamName,
		Members:  members,
	}
	return team, nil
}
