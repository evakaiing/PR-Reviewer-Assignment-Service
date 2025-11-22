CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY
    , username VARCHAR(255) NOT NULL UNIQUE
    , team_name VARCHAR(255) NOT NULL
    , is_active BOOLEAN NOT NULL DEFAULT true
    , created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
    , updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_team_name ON users(team_name);