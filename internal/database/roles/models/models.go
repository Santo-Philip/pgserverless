package models

type RoleInfo struct {
	Rolname        string   `json:"rolname"`
	OID            uint32   `json:"oid"`
	Rolsuper       bool     `json:"rolsuper"`
	Rolinherit     bool     `json:"rolinherit"`
	Rolcreaterole  bool     `json:"rolcreaterole"`
	Rolcreatedb    bool     `json:"rolcreatedb"`
	Rolcanlogin    bool     `json:"rolcanlogin"`
	Rolreplication bool     `json:"rolreplication"`
	Rolconnlimit   int32    `json:"rolconnlimit"`
	Rolvaliduntil  *string  `json:"rolvaliduntil,omitempty"`
	MemberOf       []string `json:"member_of"`
	RolSysID       *string  `json:"rolsysid,omitempty"`
}

type CreateRoleRequest struct {
	Name           string   `json:"name"`
	Password       string   `json:"password,omitempty"`
	Login          *bool    `json:"login,omitempty"`
	Superuser      *bool    `json:"superuser,omitempty"`
	Createdb       *bool    `json:"createdb,omitempty"`
	Createrole     *bool    `json:"createrole,omitempty"`
	Replication    *bool    `json:"replication,omitempty"`
	ConnectionLimit *int32  `json:"connection_limit,omitempty"`
	ValidUntil     *string  `json:"valid_until,omitempty"`
	InRoles        []string `json:"in_roles,omitempty"`
}

type AlterRoleRequest struct {
	Password       *string `json:"password,omitempty"`
	Login          *bool   `json:"login,omitempty"`
	Superuser      *bool   `json:"superuser,omitempty"`
	Createdb       *bool   `json:"createdb,omitempty"`
	Createrole     *bool   `json:"createrole,omitempty"`
	Replication    *bool   `json:"replication,omitempty"`
	ConnectionLimit *int32 `json:"connection_limit,omitempty"`
	ValidUntil     *string `json:"valid_until,omitempty"`
	Name           *string `json:"name,omitempty"`
}

type DropRoleRequest struct {
	Name string `json:"name"`
}

type GrantRequest struct {
	Role         string `json:"role"`
	Schema       string `json:"schema,omitempty"`
	Table        string `json:"table,omitempty"`
	PermissionType string `json:"permission_type"`
	GrantOption  bool   `json:"grant_option,omitempty"`
}

type RevokeRequest struct {
	Role           string `json:"role"`
	Schema         string `json:"schema,omitempty"`
	Table          string `json:"table,omitempty"`
	PermissionType string `json:"permission_type"`
	Cascade        bool   `json:"cascade,omitempty"`
}

type MembershipRequest struct {
	Role       string `json:"role"`
	MemberOf   string `json:"member_of"`
	AdminOption bool  `json:"admin_option,omitempty"`
}

type DatabasePrivilege struct {
	Database      string `json:"database"`
	Grantee       string `json:"grantee"`
	PrivilegeType string `json:"privilege_type"`
	Grantable     bool   `json:"grantable"`
}

type SchemaPrivilege struct {
	Schema        string `json:"schema"`
	Grantee       string `json:"grantee"`
	PrivilegeType string `json:"privilege_type"`
	Grantable     bool   `json:"grantable"`
}

type TablePrivilege struct {
	Schema        string `json:"schema"`
	Table         string `json:"table"`
	Grantee       string `json:"grantee"`
	PrivilegeType string `json:"privilege_type"`
	Grantable     bool   `json:"grantable"`
}

type PasswordRequest struct {
	Password string `json:"password"`
}
