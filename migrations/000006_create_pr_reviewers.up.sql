CREATE TABLE pull_request_reviewers (
    pull_request_id VARCHAR(255) NOT NULL
    , reviewer_user_id VARCHAR(255) NOT NULL
    , assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
    
    , PRIMARY KEY (pull_request_id, reviewer_user_id)
    
    , CONSTRAINT pr_reviewers_pull_request_id_fkey 
        FOREIGN KEY (pull_request_id) 
        REFERENCES pull_requests(pull_request_id) 
        ON DELETE CASCADE
    
    , CONSTRAINT pr_reviewers_reviewer_user_id_fkey 
        FOREIGN KEY (reviewer_user_id) 
        REFERENCES users(user_id) 
        ON DELETE CASCADE
);

CREATE INDEX idx_pr_reviewers_pull_request_id ON pull_request_reviewers(pull_request_id);
CREATE INDEX idx_pr_reviewers_reviewer_user_id ON pull_request_reviewers(reviewer_user_id);
