-- name: CreateAuditLog :one
INSERT INTO admin_audit_logs (
    admin_user_id,
    admin_email,
    target_user_id,
    action,
    resource,
    method,
    path,
    query_params,
    request_body,
    ip_address,
    user_agent,
    status_code,
    response_time_ms,
    timestamp,
    request_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetAuditLogByID :one
SELECT * FROM admin_audit_logs 
WHERE id = ?;

-- name: GetAuditLogsByAdminUser :many
SELECT * FROM admin_audit_logs 
WHERE admin_user_id = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsByTargetUser :many
SELECT * FROM admin_audit_logs 
WHERE target_user_id = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsByAction :many
SELECT * FROM admin_audit_logs 
WHERE action = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsByResource :many
SELECT * FROM admin_audit_logs 
WHERE resource = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsByDateRange :many
SELECT * FROM admin_audit_logs 
WHERE timestamp BETWEEN ? AND ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsByAdminAndDateRange :many
SELECT * FROM admin_audit_logs 
WHERE admin_user_id = ? 
  AND timestamp BETWEEN ? AND ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsByTargetUserAndAction :many
SELECT * FROM admin_audit_logs 
WHERE target_user_id = ? 
  AND action = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsByIPAddress :many
SELECT * FROM admin_audit_logs 
WHERE ip_address = ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsByRequestID :one
SELECT * FROM admin_audit_logs 
WHERE request_id = ?;

-- name: GetRecentAuditLogs :many
SELECT * FROM admin_audit_logs 
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetFailedAuditLogs :many
SELECT * FROM admin_audit_logs 
WHERE status_code >= 400
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetSlowAuditLogs :many
SELECT * FROM admin_audit_logs 
WHERE response_time_ms > ?
ORDER BY response_time_ms DESC
LIMIT ?;

-- name: CountAuditLogsByAdmin :one
SELECT COUNT(*) FROM admin_audit_logs 
WHERE admin_user_id = ?;

-- name: CountAuditLogsByAction :one
SELECT COUNT(*) FROM admin_audit_logs 
WHERE action = ?;

-- name: CountAuditLogsByDateRange :one
SELECT COUNT(*) FROM admin_audit_logs 
WHERE timestamp BETWEEN ? AND ?;

-- name: GetAuditLogStatsByAdmin :many
SELECT 
    admin_user_id,
    admin_email,
    COUNT(*) as total_actions,
    COUNT(DISTINCT target_user_id) as unique_targets,
    COUNT(DISTINCT action) as unique_actions,
    AVG(response_time_ms) as avg_response_time,
    MAX(timestamp) as last_activity
FROM admin_audit_logs 
GROUP BY admin_user_id, admin_email
ORDER BY total_actions DESC;

-- name: GetAuditLogStatsByResource :many
SELECT 
    resource,
    COUNT(*) as total_actions,
    COUNT(DISTINCT admin_user_id) as unique_admins,
    AVG(response_time_ms) as avg_response_time,
    SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_count
FROM admin_audit_logs 
GROUP BY resource
ORDER BY total_actions DESC;

-- name: GetAuditLogStatsByAction :many
SELECT 
    action,
    COUNT(*) as total_count,
    COUNT(DISTINCT admin_user_id) as unique_admins,
    COUNT(DISTINCT target_user_id) as unique_targets,
    AVG(response_time_ms) as avg_response_time
FROM admin_audit_logs 
GROUP BY action
ORDER BY total_count DESC;

-- name: SearchAuditLogs :many
SELECT * FROM admin_audit_logs 
WHERE (
    admin_email LIKE '%' || ? || '%' OR
    target_user_id LIKE '%' || ? || '%' OR
    action LIKE '%' || ? || '%' OR
    resource LIKE '%' || ? || '%' OR
    path LIKE '%' || ? || '%'
)
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetAuditLogsWithPagination :many
SELECT * FROM admin_audit_logs 
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: DeleteOldAuditLogs :exec
DELETE FROM admin_audit_logs 
WHERE timestamp < ?;

-- name: GetAuditLogsByMultipleCriteria :many
SELECT * FROM admin_audit_logs 
WHERE 
    (? = '' OR admin_user_id = ?) AND
    (? = '' OR target_user_id = ?) AND
    (? = '' OR action = ?) AND
    (? = '' OR resource = ?) AND
    (? = 0 OR status_code = ?) AND
    timestamp BETWEEN ? AND ?
ORDER BY timestamp DESC
LIMIT ?;

-- name: GetSuspiciousAuditLogs :many
-- Logs with multiple failed attempts from same admin/IP
SELECT a.* FROM admin_audit_logs a
WHERE (a.admin_user_id, a.ip_address) IN (
    SELECT b.admin_user_id, b.ip_address 
    FROM admin_audit_logs b
    WHERE b.status_code >= 400 
      AND b.timestamp >= ?
    GROUP BY b.admin_user_id, b.ip_address 
    HAVING COUNT(*) >= ?
)
ORDER BY a.timestamp DESC;

-- name: GetDataAccessAuditLogs :many
-- Specifically for sensitive data access (VIEW actions)
SELECT * FROM admin_audit_logs 
WHERE action = 'VIEW' 
  AND resource IN ('users', 'orders', 'payments', 'personal_data')
  AND timestamp BETWEEN ? AND ?
ORDER BY timestamp DESC
LIMIT ?;
