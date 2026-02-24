-- name: AppendAuditLog :exec
INSERT INTO audit_log (id, user_id, entity_type, entity_id, action, action_data, occurred_at, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
