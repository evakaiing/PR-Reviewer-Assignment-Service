CREATE TABLE pull_request_statuses (
    status_id SERIAL PRIMARY KEY
    , status_name VARCHAR(50) NOT NULL UNIQUE
    , created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO pull_request_statuses (status_name) VALUES 
    ('OPEN'),
    ('MERGED');

CREATE INDEX idx_pr_statuses_name ON pull_request_statuses(status_name);