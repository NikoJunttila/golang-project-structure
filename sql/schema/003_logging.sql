-- +goose Up
CREATE TABLE admin_audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_user_id TEXT NOT NULL,
    admin_email TEXT NOT NULL,
    target_user_id TEXT NOT NULL,        -- WHO's data was accessed
    action TEXT NOT NULL,       -- VIEW, CREATE, UPDATE, DELETE
    resource TEXT NOT NULL,     -- users, orders, reports, etc.
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    query_params TEXT NOT NULL,
    request_body TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    request_id TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS admin_audit_logs;
