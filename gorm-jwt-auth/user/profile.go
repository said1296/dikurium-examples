package user

// @GormDBNames
type Profile struct {
	UserID int `gorm:"primaryKey"`
	Description string
	HasImage bool
}
