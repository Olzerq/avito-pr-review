CREATE TABLE teams (
                       team_name TEXT PRIMARY KEY
);


CREATE TABLE users (
                       user_id    TEXT PRIMARY KEY,
                       username   TEXT NOT NULL,
                       team_name  TEXT NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
                       is_active  BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE pull_requests (
                               pull_request_id   TEXT PRIMARY KEY,
                               pull_request_name TEXT NOT NULL,
                               author_id         TEXT NOT NULL REFERENCES users(user_id),
                               status            TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
                               merged_at         TIMESTAMP NULL
);

CREATE TABLE pull_request_reviewers (
                                        pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
                                        user_id         TEXT NOT NULL REFERENCES users(user_id),
                                        PRIMARY KEY (pull_request_id, user_id)
);

CREATE INDEX idx_users_team_active ON users(team_name, is_active);
CREATE INDEX idx_reviewers_user ON pull_request_reviewers(user_id);