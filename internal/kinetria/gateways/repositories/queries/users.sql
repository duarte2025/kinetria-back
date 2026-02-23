-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, profile_image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, name, email, password_hash, profile_image_url, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, profile_image_url, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT id, name, email, password_hash, profile_image_url, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;
