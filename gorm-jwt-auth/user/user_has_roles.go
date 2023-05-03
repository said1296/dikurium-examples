package user

// @GormDBNames
type UserHasRoles struct {
	UserID	int `gorm:"primaryKey;"`
	RoleID	int `gorm:"primaryKey"`
}
