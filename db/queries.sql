-- name: InsertConfig :one
INSERT INTO app_schema_name.configuration (
    api_access_change_me,
    refresh_interval,
    request_timeout,
    asset_filter,
    active,
    enable,
    project_ids,
    user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpsertConfig :one
INSERT INTO app_schema_name.configuration (
    id,
    api_access_change_me,
    refresh_interval,
    request_timeout,
    asset_filter,
    active,
    enable,
    project_ids,
    user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) ON CONFLICT (id) DO UPDATE SET
    api_access_change_me = EXCLUDED.api_access_change_me,
    refresh_interval = EXCLUDED.refresh_interval,
    request_timeout = EXCLUDED.request_timeout,
    asset_filter = EXCLUDED.asset_filter,
    active = EXCLUDED.active,
    enable = EXCLUDED.enable,
    project_ids = EXCLUDED.project_ids,
    user_id = EXCLUDED.user_id
RETURNING *;

-- name: GetConfig :one
SELECT * FROM app_schema_name.configuration
WHERE id = $1;

-- name: GetConfigs :many
SELECT * FROM app_schema_name.configuration;

-- name: DeleteConfig :exec
DELETE FROM app_schema_name.configuration WHERE id = $1;

-- name: SetConfigActiveState :exec
UPDATE app_schema_name.configuration
SET active = $2
WHERE id = $1;

-- name: SetAllConfigsInactive :exec
UPDATE app_schema_name.configuration
SET active = FALSE;

-- name: InsertAsset :exec
INSERT INTO app_schema_name.asset (
    configuration_id,
    project_id,
    global_asset_id,
    provider_id,
    asset_id
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetAssetId :one
SELECT asset_id FROM app_schema_name.asset
WHERE configuration_id = $1
AND project_id = $2
AND global_asset_id = $3
LIMIT 1;

-- name: GetAssetById :one
SELECT a.*, c.*
FROM app_schema_name.asset a
JOIN app_schema_name.configuration c ON a.configuration_id = c.id
WHERE a.id = $1;
