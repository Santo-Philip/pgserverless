package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexbic/platform/internal/pgroles/models"
	"github.com/nexbic/platform/pkg/database"
	"github.com/nexbic/platform/pkg/helpers"
)

type PgRolesService struct {
	db *database.DB
}

func NewPgRolesService(db *database.DB) *PgRolesService {
	return &PgRolesService{db: db}
}

func quoteLiteral(val string) string {
	return "'" + strings.ReplaceAll(val, "'", "''") + "'"
}

func (s *PgRolesService) ListRoles(ctx context.Context) ([]models.RoleInfo, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			r.rolname,
			r.oid,
			r.rolsuper,
			r.rolinherit,
			r.rolcreaterole,
			r.rolcreatedb,
			r.rolcanlogin,
			r.rolreplication,
			r.rolconnlimit,
			r.rolvaliduntil,
			COALESCE(array_agg(DISTINCT m.rolname ORDER BY m.rolname) FILTER (WHERE m.rolname IS NOT NULL), '{}') AS member_of
		FROM pg_catalog.pg_roles r
		LEFT JOIN pg_catalog.pg_auth_members am ON r.oid = am.member
		LEFT JOIN pg_catalog.pg_roles m ON am.roleid = m.oid
		GROUP BY r.oid, r.rolname, r.rolsuper, r.rolinherit, r.rolcreaterole, r.rolcreatedb,
			r.rolcanlogin, r.rolreplication, r.rolconnlimit, r.rolvaliduntil
		ORDER BY r.rolname`)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles []models.RoleInfo
	for rows.Next() {
		var r models.RoleInfo
		var validUntil *time.Time
		if err := rows.Scan(
			&r.Rolname, &r.OID, &r.Rolsuper, &r.Rolinherit, &r.Rolcreaterole,
			&r.Rolcreatedb, &r.Rolcanlogin, &r.Rolreplication, &r.Rolconnlimit,
			&validUntil, &r.MemberOf,
		); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		if validUntil != nil {
			vs := validUntil.Format("2006-01-02 15:04:05-07")
			r.Rolvaliduntil = &vs
		}
		roles = append(roles, r)
	}

	if roles == nil {
		roles = []models.RoleInfo{}
	}

	return roles, nil
}

func (s *PgRolesService) RoleExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := s.db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = $1)`, name,
	).Scan(&exists)
	return exists, err
}

func (s *PgRolesService) GetRole(ctx context.Context, name string) (*models.RoleInfo, error) {
	row := s.db.Pool.QueryRow(ctx, `
		SELECT
			r.rolname, r.oid, r.rolsuper, r.rolinherit, r.rolcreaterole, r.rolcreatedb,
			r.rolcanlogin, r.rolreplication, r.rolconnlimit, r.rolvaliduntil,
			COALESCE(array_agg(DISTINCT m.rolname ORDER BY m.rolname) FILTER (WHERE m.rolname IS NOT NULL), '{}') AS member_of
		FROM pg_catalog.pg_roles r
		LEFT JOIN pg_catalog.pg_auth_members am ON r.oid = am.member
		LEFT JOIN pg_catalog.pg_roles m ON am.roleid = m.oid
		WHERE r.rolname = $1
		GROUP BY r.oid, r.rolname, r.rolsuper, r.rolinherit, r.rolcreaterole, r.rolcreatedb,
			r.rolcanlogin, r.rolreplication, r.rolconnlimit, r.rolvaliduntil`, name)

	var roleInfo models.RoleInfo
	var validUntil *time.Time
	err := row.Scan(
		&roleInfo.Rolname, &roleInfo.OID, &roleInfo.Rolsuper, &roleInfo.Rolinherit, &roleInfo.Rolcreaterole,
		&roleInfo.Rolcreatedb, &roleInfo.Rolcanlogin, &roleInfo.Rolreplication, &roleInfo.Rolconnlimit,
		&validUntil, &roleInfo.MemberOf,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get role: %w", err)
	}
	if validUntil != nil {
		vs := validUntil.Format("2006-01-02 15:04:05-07")
		roleInfo.Rolvaliduntil = &vs
	}
	return &roleInfo, nil
}

func (s *PgRolesService) CreateRole(ctx context.Context, req *models.CreateRoleRequest) error {
	exists, err := s.RoleExists(ctx, req.Name)
	if err != nil {
		return fmt.Errorf("check role exists: %w", err)
	}
	if exists {
		return fmt.Errorf("role %q already exists", req.Name)
	}

	var opts []string

	if req.Password != "" {
		opts = append(opts, fmt.Sprintf("PASSWORD %s", quoteLiteral(req.Password)))
	} else {
		opts = append(opts, "PASSWORD NULL")
	}

	if req.Login != nil {
		if *req.Login {
			opts = append(opts, "LOGIN")
		} else {
			opts = append(opts, "NOLOGIN")
		}
	}
	if req.Superuser != nil {
		if *req.Superuser {
			opts = append(opts, "SUPERUSER")
		} else {
			opts = append(opts, "NOSUPERUSER")
		}
	}
	if req.Createdb != nil {
		if *req.Createdb {
			opts = append(opts, "CREATEDB")
		} else {
			opts = append(opts, "NOCREATEDB")
		}
	}
	if req.Createrole != nil {
		if *req.Createrole {
			opts = append(opts, "CREATEROLE")
		} else {
			opts = append(opts, "NOCREATEROLE")
		}
	}
	if req.Replication != nil {
		if *req.Replication {
			opts = append(opts, "REPLICATION")
		} else {
			opts = append(opts, "NOREPLICATION")
		}
	}
	if req.ConnectionLimit != nil {
		opts = append(opts, fmt.Sprintf("CONNECTION LIMIT %d", *req.ConnectionLimit))
	}
	if req.ValidUntil != nil {
		opts = append(opts, fmt.Sprintf("VALID UNTIL %s", quoteLiteral(*req.ValidUntil)))
	}

	sql := fmt.Sprintf("CREATE ROLE %s WITH %s", helpers.QuoteIdent(req.Name), strings.Join(opts, " "))
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("create role: %w", err)
	}

	for _, memberOf := range req.InRoles {
		grantsql := fmt.Sprintf("GRANT %s TO %s", helpers.QuoteIdent(memberOf), helpers.QuoteIdent(req.Name))
		if _, err := s.db.Pool.Exec(ctx, grantsql); err != nil {
			return fmt.Errorf("grant role membership %q: %w", memberOf, err)
		}
	}

	return nil
}

func (s *PgRolesService) AlterRole(ctx context.Context, name string, req *models.AlterRoleRequest) error {
	var opts []string

	if req.Password != nil {
		opts = append(opts, fmt.Sprintf("PASSWORD %s", quoteLiteral(*req.Password)))
	}
	if req.Login != nil {
		if *req.Login {
			opts = append(opts, "LOGIN")
		} else {
			opts = append(opts, "NOLOGIN")
		}
	}
	if req.Superuser != nil {
		if *req.Superuser {
			opts = append(opts, "SUPERUSER")
		} else {
			opts = append(opts, "NOSUPERUSER")
		}
	}
	if req.Createdb != nil {
		if *req.Createdb {
			opts = append(opts, "CREATEDB")
		} else {
			opts = append(opts, "NOCREATEDB")
		}
	}
	if req.Createrole != nil {
		if *req.Createrole {
			opts = append(opts, "CREATEROLE")
		} else {
			opts = append(opts, "NOCREATEROLE")
		}
	}
	if req.Replication != nil {
		if *req.Replication {
			opts = append(opts, "REPLICATION")
		} else {
			opts = append(opts, "NOREPLICATION")
		}
	}
	if req.ConnectionLimit != nil {
		opts = append(opts, fmt.Sprintf("CONNECTION LIMIT %d", *req.ConnectionLimit))
	}
	if req.ValidUntil != nil {
		opts = append(opts, fmt.Sprintf("VALID UNTIL %s", quoteLiteral(*req.ValidUntil)))
	}
	if req.Name != nil {
		sql := fmt.Sprintf("ALTER ROLE %s RENAME TO %s", helpers.QuoteIdent(name), helpers.QuoteIdent(*req.Name))
		if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("rename role: %w", err)
		}
	}

	if len(opts) > 0 {
		sql := fmt.Sprintf("ALTER ROLE %s WITH %s", helpers.QuoteIdent(name), strings.Join(opts, " "))
		if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("alter role: %w", err)
		}
	}

	return nil
}

func (s *PgRolesService) DropRole(ctx context.Context, name string) error {
	exists, err := s.RoleExists(ctx, name)
	if err != nil {
		return fmt.Errorf("check role exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("role %q does not exist", name)
	}

	var objCount int
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pg_catalog.pg_shdepend d
		JOIN pg_catalog.pg_roles r ON d.refobjid = r.oid
		WHERE r.rolname = $1 AND d.deptype = 'o'`, name).Scan(&objCount)
	if err != nil {
		return fmt.Errorf("check role dependencies: %w", err)
	}
	if objCount > 0 {
		return fmt.Errorf("role %q owns %d object(s); reassign or drop them first", name, objCount)
	}

	sql := fmt.Sprintf("DROP ROLE %s", helpers.QuoteIdent(name))
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("drop role: %w", err)
	}
	return nil
}

func (s *PgRolesService) GrantDatabase(ctx context.Context, role, databaseName, permissions string) error {
	sql := fmt.Sprintf("GRANT %s ON DATABASE %s TO %s",
		permissions, helpers.QuoteIdent(databaseName), helpers.QuoteIdent(role))
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("grant database: %w", err)
	}
	return nil
}

func (s *PgRolesService) GrantSchema(ctx context.Context, role, schema, permissions string) error {
	sql := fmt.Sprintf("GRANT %s ON SCHEMA %s TO %s",
		permissions, helpers.QuoteIdent(schema), helpers.QuoteIdent(role))
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("grant schema: %w", err)
	}
	return nil
}

func (s *PgRolesService) GrantTable(ctx context.Context, role, schema, table, permissions string) error {
	var sql string
	if schema != "" && table != "" {
		sql = fmt.Sprintf("GRANT %s ON TABLE %s.%s TO %s",
			permissions, helpers.QuoteIdent(schema), helpers.QuoteIdent(table), helpers.QuoteIdent(role))
	} else {
		sql = fmt.Sprintf("GRANT %s ON ALL TABLES IN SCHEMA %s TO %s",
			permissions, helpers.QuoteIdent(schema), helpers.QuoteIdent(role))
	}
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("grant table: %w", err)
	}
	return nil
}

func (s *PgRolesService) RevokeDatabase(ctx context.Context, role, databaseName, permissions string, cascade bool) error {
	sql := fmt.Sprintf("REVOKE %s ON DATABASE %s FROM %s",
		permissions, helpers.QuoteIdent(databaseName), helpers.QuoteIdent(role))
	if cascade {
		sql += " CASCADE"
	}
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("revoke database: %w", err)
	}
	return nil
}

func (s *PgRolesService) RevokeSchema(ctx context.Context, role, schema, permissions string, cascade bool) error {
	sql := fmt.Sprintf("REVOKE %s ON SCHEMA %s FROM %s",
		permissions, helpers.QuoteIdent(schema), helpers.QuoteIdent(role))
	if cascade {
		sql += " CASCADE"
	}
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("revoke schema: %w", err)
	}
	return nil
}

func (s *PgRolesService) RevokeTable(ctx context.Context, role, schema, table, permissions string, cascade bool) error {
	var sql string
	if schema != "" && table != "" {
		sql = fmt.Sprintf("REVOKE %s ON TABLE %s.%s FROM %s",
			permissions, helpers.QuoteIdent(schema), helpers.QuoteIdent(table), helpers.QuoteIdent(role))
	} else {
		sql = fmt.Sprintf("REVOKE %s ON ALL TABLES IN SCHEMA %s FROM %s",
			permissions, helpers.QuoteIdent(schema), helpers.QuoteIdent(role))
	}
	if cascade {
		sql += " CASCADE"
	}
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("revoke table: %w", err)
	}
	return nil
}

func (s *PgRolesService) AddMember(ctx context.Context, role, memberOf string) error {
	sql := fmt.Sprintf("GRANT %s TO %s", helpers.QuoteIdent(memberOf), helpers.QuoteIdent(role))
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (s *PgRolesService) RemoveMember(ctx context.Context, role, memberOf string) error {
	sql := fmt.Sprintf("REVOKE %s FROM %s", helpers.QuoteIdent(memberOf), helpers.QuoteIdent(role))
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	return nil
}

func (s *PgRolesService) ListDatabasePrivileges(ctx context.Context, databaseName string) ([]models.DatabasePrivilege, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			datname,
			CASE WHEN grantee = 0 THEN 'PUBLIC' ELSE r.rolname END AS grantee,
			priv,
			grantable
		FROM (
			SELECT
				dat.datname,
				(aclexplode(COALESCE(dat.datacl, acldefault('d', dat.datdba)))).grantee AS grantee,
				(aclexplode(COALESCE(dat.datacl, acldefault('d', dat.datdba)))).privilege_type AS priv,
				(aclexplode(COALESCE(dat.datacl, acldefault('d', dat.datdba)))).is_grantable AS grantable
			FROM pg_catalog.pg_database dat
			WHERE dat.datname = $1
		) sub
		LEFT JOIN pg_catalog.pg_roles r ON r.oid = sub.grantee
		ORDER BY grantee, priv`, databaseName)
	if err != nil {
		return nil, fmt.Errorf("list database privileges: %w", err)
	}
	defer rows.Close()

	var privileges []models.DatabasePrivilege
	for rows.Next() {
		var p models.DatabasePrivilege
		if err := rows.Scan(&p.Database, &p.Grantee, &p.PrivilegeType, &p.Grantable); err != nil {
			return nil, fmt.Errorf("scan database privilege: %w", err)
		}
		privileges = append(privileges, p)
	}

	if privileges == nil {
		privileges = []models.DatabasePrivilege{}
	}
	return privileges, nil
}

func (s *PgRolesService) ListSchemaPrivileges(ctx context.Context, schemaName string) ([]models.SchemaPrivilege, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT table_schema, grantee, privilege_type,
			CASE WHEN is_grantable = 'YES' THEN TRUE ELSE FALSE END
		FROM information_schema.schema_privileges
		WHERE table_schema = $1
		ORDER BY grantee, privilege_type`, schemaName)
	if err != nil {
		return nil, fmt.Errorf("list schema privileges: %w", err)
	}
	defer rows.Close()

	var privileges []models.SchemaPrivilege
	for rows.Next() {
		var p models.SchemaPrivilege
		if err := rows.Scan(&p.Schema, &p.Grantee, &p.PrivilegeType, &p.Grantable); err != nil {
			return nil, fmt.Errorf("scan schema privilege: %w", err)
		}
		privileges = append(privileges, p)
	}

	if privileges == nil {
		privileges = []models.SchemaPrivilege{}
	}
	return privileges, nil
}

func (s *PgRolesService) ListTablePrivileges(ctx context.Context, schemaName, tableName string) ([]models.TablePrivilege, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT table_schema, table_name, grantee, privilege_type,
			CASE WHEN is_grantable = 'YES' THEN TRUE ELSE FALSE END
		FROM information_schema.table_privileges
		WHERE ($1 = '' OR table_schema = $1) AND ($2 = '' OR table_name = $2)
		ORDER BY table_schema, table_name, grantee, privilege_type`, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("list table privileges: %w", err)
	}
	defer rows.Close()

	var privileges []models.TablePrivilege
	for rows.Next() {
		var p models.TablePrivilege
		if err := rows.Scan(&p.Schema, &p.Table, &p.Grantee, &p.PrivilegeType, &p.Grantable); err != nil {
			return nil, fmt.Errorf("scan table privilege: %w", err)
		}
		privileges = append(privileges, p)
	}

	if privileges == nil {
		privileges = []models.TablePrivilege{}
	}
	return privileges, nil
}

func (s *PgRolesService) GetRoleMembers(ctx context.Context, role string) ([]models.RoleInfo, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT
			r.rolname, r.oid, r.rolsuper, r.rolinherit, r.rolcreaterole, r.rolcreatedb,
			r.rolcanlogin, r.rolreplication, r.rolconnlimit, r.rolvaliduntil,
			COALESCE(array_agg(DISTINCT m.rolname ORDER BY m.rolname) FILTER (WHERE m.rolname IS NOT NULL), '{}') AS member_of
		FROM pg_catalog.pg_auth_members am
		JOIN pg_catalog.pg_roles r ON am.member = r.oid
		JOIN pg_catalog.pg_roles parent ON am.roleid = parent.oid
		LEFT JOIN pg_catalog.pg_auth_members am2 ON r.oid = am2.member
		LEFT JOIN pg_catalog.pg_roles m ON am2.roleid = m.oid
		WHERE parent.rolname = $1
		GROUP BY r.oid, r.rolname, r.rolsuper, r.rolinherit, r.rolcreaterole, r.rolcreatedb,
			r.rolcanlogin, r.rolreplication, r.rolconnlimit, r.rolvaliduntil
		ORDER BY r.rolname`, role)
	if err != nil {
		return nil, fmt.Errorf("get role members: %w", err)
	}
	defer rows.Close()

	var members []models.RoleInfo
	for rows.Next() {
		var r models.RoleInfo
		var validUntil *time.Time
		if err := rows.Scan(
			&r.Rolname, &r.OID, &r.Rolsuper, &r.Rolinherit, &r.Rolcreaterole,
			&r.Rolcreatedb, &r.Rolcanlogin, &r.Rolreplication, &r.Rolconnlimit,
			&validUntil, &r.MemberOf,
		); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		if validUntil != nil {
			vs := validUntil.Format("2006-01-02 15:04:05-07")
			r.Rolvaliduntil = &vs
		}
		members = append(members, r)
	}

	if members == nil {
		members = []models.RoleInfo{}
	}
	return members, nil
}

func (s *PgRolesService) ResetPassword(ctx context.Context, role, password string) error {
	sql := fmt.Sprintf("ALTER ROLE %s WITH PASSWORD %s", helpers.QuoteIdent(role), quoteLiteral(password))
	if _, err := s.db.Pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	return nil
}
