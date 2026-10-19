package access

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRoleNotFound = errors.New("role not found")
	ErrSystemRole   = errors.New("system role access cannot be changed")
	ErrNotFound     = errors.New("access not found")
	ErrNameTaken    = errors.New("access name already exists")
	ErrInvalid      = errors.New("invalid access")
)

// Record is a named access that can be aligned to a role.
type Record struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Resource  string    `json:"resource"`
	CanCreate bool      `json:"can_create"`
	CanRead   bool      `json:"can_read"`
	CanUpdate bool      `json:"can_update"`
	CanDelete bool      `json:"can_delete"`
}

type RoleLinks struct {
	IsSystem  bool        `json:"is_system"`
	AccessIDs []uuid.UUID `json:"access_ids"`
}

type RoleAccess struct {
	IsSystem bool    `json:"is_system"`
	Access   []Grant `json:"access"`
}

// ListForUser returns the signed-in user's page access.
// A system role, or an admin account that has no role row yet, receives every page.
func ListForUser(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) (RoleAccess, error) {
	var roleName string
	var isSystem bool
	var roleID *uuid.UUID
	err := db.QueryRow(ctx, `
		SELECT u.role, COALESCE(r.is_system, false), r.id
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(&roleName, &isSystem, &roleID)
	if err != nil {
		return RoleAccess{}, err
	}
	if privileged(roleName, isSystem, roleID != nil) {
		return RoleAccess{IsSystem: true, Access: FullGrants()}, nil
	}
	if roleID == nil {
		return RoleAccess{Access: Normalize(nil)}, nil
	}
	grants, err := grantsForRole(ctx, db, *roleID)
	if err != nil {
		return RoleAccess{}, err
	}
	return RoleAccess{Access: grants}, nil
}

func privileged(roleName string, isSystem bool, hasRole bool) bool {
	if isSystem {
		return true
	}
	return !hasRole && strings.EqualFold(strings.TrimSpace(roleName), "admin")
}

func List(ctx context.Context, db *pgxpool.Pool, orgID uuid.UUID) ([]Record, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, resource, can_create, can_read, can_update, can_delete
		FROM accesses
		WHERE organisation_id = $1
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := []Record{}
	for rows.Next() {
		var record Record
		if err := rows.Scan(&record.ID, &record.Name, &record.Resource, &record.CanCreate, &record.CanRead, &record.CanUpdate, &record.CanDelete); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func Create(ctx context.Context, db *pgxpool.Pool, orgID uuid.UUID, name, resource string, canCreate, canRead, canUpdate, canDelete bool) (Record, error) {
	record, err := prepare(name, resource, canCreate, canRead, canUpdate, canDelete)
	if err != nil {
		return Record{}, err
	}
	err = db.QueryRow(ctx, `
		INSERT INTO accesses (organisation_id, name, resource, can_create, can_read, can_update, can_delete, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`, orgID, record.Name, record.Resource, record.CanCreate, record.CanRead, record.CanUpdate, record.CanDelete).Scan(&record.ID)
	if isUnique(err) {
		return Record{}, ErrNameTaken
	}
	return record, err
}

func Update(ctx context.Context, db *pgxpool.Pool, orgID, id uuid.UUID, name, resource string, canCreate, canRead, canUpdate, canDelete bool) (Record, error) {
	record, err := prepare(name, resource, canCreate, canRead, canUpdate, canDelete)
	if err != nil {
		return Record{}, err
	}
	record.ID = id
	tag, err := db.Exec(ctx, `
		UPDATE accesses
		SET name = $3, resource = $4, can_create = $5, can_read = $6, can_update = $7, can_delete = $8, updated_at = NOW()
		WHERE id = $1 AND organisation_id = $2
	`, id, orgID, record.Name, record.Resource, record.CanCreate, record.CanRead, record.CanUpdate, record.CanDelete)
	if isUnique(err) {
		return Record{}, ErrNameTaken
	}
	if err != nil {
		return Record{}, err
	}
	if tag.RowsAffected() == 0 {
		return Record{}, ErrNotFound
	}
	return record, nil
}

func Delete(ctx context.Context, db *pgxpool.Pool, orgID, id uuid.UUID) error {
	tag, err := db.Exec(ctx, `DELETE FROM accesses WHERE id = $1 AND organisation_id = $2`, id, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func LinksForRole(ctx context.Context, db *pgxpool.Pool, orgID, roleID uuid.UUID) (RoleLinks, error) {
	isSystem, err := roleInOrg(ctx, db, orgID, roleID)
	if err != nil {
		return RoleLinks{}, err
	}
	if isSystem {
		return RoleLinks{IsSystem: true, AccessIDs: []uuid.UUID{}}, nil
	}
	rows, err := db.Query(ctx, `
		SELECT ra.access_id
		FROM role_accesses ra
		JOIN accesses a ON a.id = ra.access_id
		WHERE ra.role_id = $1 AND a.organisation_id = $2
		ORDER BY a.name
	`, roleID, orgID)
	if err != nil {
		return RoleLinks{}, err
	}
	defer rows.Close()
	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return RoleLinks{}, err
		}
		ids = append(ids, id)
	}
	return RoleLinks{AccessIDs: ids}, rows.Err()
}

func SetRoleLinks(ctx context.Context, db *pgxpool.Pool, orgID, roleID uuid.UUID, accessIDs []uuid.UUID) (RoleLinks, error) {
	isSystem, err := roleInOrg(ctx, db, orgID, roleID)
	if err != nil {
		return RoleLinks{}, err
	}
	if isSystem {
		return RoleLinks{}, ErrSystemRole
	}
	unique := map[uuid.UUID]struct{}{}
	for _, id := range accessIDs {
		unique[id] = struct{}{}
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return RoleLinks{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM role_accesses WHERE role_id = $1`, roleID); err != nil {
		return RoleLinks{}, err
	}
	saved := []uuid.UUID{}
	for id := range unique {
		tag, err := tx.Exec(ctx, `
			INSERT INTO role_accesses (role_id, access_id)
			SELECT $1, id FROM accesses WHERE id = $2 AND organisation_id = $3
		`, roleID, id, orgID)
		if err != nil {
			return RoleLinks{}, err
		}
		if tag.RowsAffected() == 1 {
			saved = append(saved, id)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return RoleLinks{}, err
	}
	return RoleLinks{AccessIDs: saved}, nil
}

func Allowed(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, resource, action string) (bool, error) {
	record, err := ListForUser(ctx, db, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return Allows(record.Access, resource, action), nil
}

func prepare(name, resource string, canCreate, canRead, canUpdate, canDelete bool) (Record, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Record{}, ErrInvalid
	}
	grant := Normalize([]Grant{{
		Resource:  resource,
		CanCreate: canCreate,
		CanRead:   canRead,
		CanUpdate: canUpdate,
		CanDelete: canDelete,
	}})
	var match Grant
	found := false
	for _, item := range grant {
		if item.Resource == resource {
			match = item
			found = true
			break
		}
	}
	if !found || (!match.CanCreate && !match.CanRead && !match.CanUpdate && !match.CanDelete) {
		return Record{}, ErrInvalid
	}
	return Record{
		Name:      name,
		Resource:  resource,
		CanCreate: match.CanCreate,
		CanRead:   match.CanRead,
		CanUpdate: match.CanUpdate,
		CanDelete: match.CanDelete,
	}, nil
}

func ListForRole(ctx context.Context, db *pgxpool.Pool, orgID, roleID uuid.UUID) (RoleAccess, error) {
	isSystem, err := roleInOrg(ctx, db, orgID, roleID)
	if err != nil {
		return RoleAccess{}, err
	}
	if isSystem {
		return RoleAccess{IsSystem: true, Access: FullGrants()}, nil
	}
	grants, err := grantsForRole(ctx, db, roleID)
	if err != nil {
		return RoleAccess{}, err
	}
	return RoleAccess{Access: grants}, nil
}

func ReplaceForRole(ctx context.Context, db *pgxpool.Pool, orgID, roleID uuid.UUID, grants []Grant) (RoleAccess, error) {
	isSystem, err := roleInOrg(ctx, db, orgID, roleID)
	if err != nil {
		return RoleAccess{}, err
	}
	if isSystem {
		return RoleAccess{}, ErrSystemRole
	}
	normalized := Normalize(grants)
	tx, err := db.Begin(ctx)
	if err != nil {
		return RoleAccess{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM accesses WHERE role_id = $1`, roleID); err != nil {
		return RoleAccess{}, err
	}
	for _, grant := range normalized {
		if !grant.CanCreate && !grant.CanRead && !grant.CanUpdate && !grant.CanDelete {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO accesses (role_id, resource, can_create, can_read, can_update, can_delete, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		`, roleID, grant.Resource, grant.CanCreate, grant.CanRead, grant.CanUpdate, grant.CanDelete); err != nil {
			return RoleAccess{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return RoleAccess{}, err
	}
	return RoleAccess{Access: normalized}, nil
}

func grantsForRole(ctx context.Context, db *pgxpool.Pool, roleID uuid.UUID) ([]Grant, error) {
	rows, err := db.Query(ctx, `
		SELECT resource, can_create, can_read, can_update, can_delete
		FROM accesses
		WHERE role_id = $1
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var grants []Grant
	for rows.Next() {
		var grant Grant
		if err := rows.Scan(&grant.Resource, &grant.CanCreate, &grant.CanRead, &grant.CanUpdate, &grant.CanDelete); err != nil {
			return nil, err
		}
		grants = append(grants, grant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return Merge(grants), nil
}

func roleInOrg(ctx context.Context, db *pgxpool.Pool, orgID, roleID uuid.UUID) (bool, error) {
	var isSystem bool
	err := db.QueryRow(ctx, `SELECT is_system FROM roles WHERE id = $1 AND organisation_id = $2`, roleID, orgID).Scan(&isSystem)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrRoleNotFound
	}
	return isSystem, err
}

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
