-- name: GetTemplates :many
SELECT *
FROM templates
WHERE organisation_id = $1
ORDER BY created_at DESC;

-- name: GetTemplateByID :one
SELECT *
FROM templates
WHERE id = $1 AND organisation_id = $2;

-- name: CreateTemplate :one
INSERT INTO templates (
    organisation_id, name, html_content, is_default, thumbnail_url, created_by, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
RETURNING *;

-- name: UpdateTemplate :one
UPDATE templates
SET
    name = COALESCE(sqlc.narg('name'), name),
    html_content = COALESCE(sqlc.narg('html_content'), html_content),
    is_default = COALESCE(sqlc.narg('is_default'), is_default),
    thumbnail_url = COALESCE(sqlc.narg('thumbnail_url'), thumbnail_url),
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: DeleteTemplate :exec
DELETE FROM templates
WHERE id = $1 AND organisation_id = $2;
