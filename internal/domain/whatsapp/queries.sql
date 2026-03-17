-- name: GetWhatsappLogs :many
SELECT *
FROM whatsapp_logs
WHERE organisation_id = $1
ORDER BY created_at DESC;

-- name: CreateWhatsappLog :one
INSERT INTO whatsapp_logs (
    organisation_id, invoice_id, recipient_phone, message, status, error_message, sent_at, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW()
)
RETURNING *;
