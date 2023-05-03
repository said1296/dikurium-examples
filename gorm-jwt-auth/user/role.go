package user

const (
	RoleAdminID = 1
	RoleUserID = 2
	RoleDesignerID = 3
)

// @GormDBNames
type Role struct {
	ID int
	Name string

	Users []*User `gorm:"many2many:user_has_roles;"`
}
