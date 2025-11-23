CREATE TABLE pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY
    , pull_request_name VARCHAR(255) NOT NULL
    , author_id VARCHAR(255) NOT NULL
    , status_id INTEGER NOT NULL DEFAULT 1 -- OPEN
    , createdAt TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
    , mergedAt TIMESTAMP WITH TIME ZONE
    
    , CONSTRAINT pull_requests_author_id_fkey 
        FOREIGN KEY (author_id) 
        REFERENCES users(user_id) 
        ON DELETE CASCADE
    , CONSTRAINT pull_requests_status_id_fkey 
        FOREIGN KEY (status_id) 
        REFERENCES pull_request_statuses(status_id) 
        ON DELETE RESTRICT
);

CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status_id ON pull_requests(status_id);
CREATE INDEX idx_pull_requests_author_status ON pull_requests(author_id, status_id);
CREATE INDEX idx_pull_requests_created_at ON pull_requests(createdAt DESC);