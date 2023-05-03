package user

import (
	"time"
)

// @GormDBNames
type User struct {
	ID            int
	FirstName     string
	LastName      string
	PreferredName string
	RegisterTime  *time.Time `gorm:"type:timestamp without time zone;"`
	Password      string     `gorm:"type:char(60)"`
	Email         string     `gorm:"unique;"`
	Country       string

	Roles               []*Role `gorm:"many2many:user_has_roles;"`
	UserHasAddresses    []UserHasAddresses
	Profile             Profile
	UserHasNftsOffchain []UserHasOffchainNfts
	StripeCustomer      StripeID

	HasCompleteProfile    bool
	HasBankAccount        bool
	HasUploadedOneNft     bool
	StripeTransferCapabilityStatus 		string `gorm:"default:inactive"`
}

func (u *User) HasRole(roleID int) bool {
	for _, role := range u.Roles {
		if roleID == role.ID {
			return true
		}
	}
	return false
}
