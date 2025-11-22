CREATE TABLE team_members (
    team_id INTEGER NOT NULL
    , user_id VARCHAR(255) NOT NULL
    , is_active BOOLEAN NOT NULL DEFAULT true
    , joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
    
    , PRIMARY KEY (team_id, user_id)
    
    , CONSTRAINT team_members_team_id_fkey 
        FOREIGN KEY (team_id) 
        REFERENCES teams(team_id) 
        ON DELETE CASCADE
    
    , CONSTRAINT team_members_user_id_fkey 
        FOREIGN KEY (user_id) 
        REFERENCES users(user_id) 
        ON DELETE CASCADE
);

CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_team_members_is_active ON team_members(is_active);