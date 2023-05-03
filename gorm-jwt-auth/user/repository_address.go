package user

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

func (r *Repository) Addresses(userId int) ([]*UserHasAddresses, error) {
	var userHasAddressesDB []*UserHasAddresses
	err := r.DB.
		Where(DBNamesUserHasAddresses.UserID, userId).
		Find(&userHasAddressesDB).Error
	if err != nil {
		err = errors.Wrap(err, "failed to get user addresses from DB")
		return nil, err
	}

	return userHasAddressesDB, nil
}

func (r *Repository) UserFromAddress(address string) (*User, error) {
	var userDB *User
	err := r.DB.
		Joins("INNER JOIN "+DBNamesUserHasAddresses.TableName+" ON LOWER("+DBNamesUserHasAddresses.AddressID+") = ? AND "+
			DBNamesUserHasAddresses.UserID+" = "+DBNamesUser.ID, strings.ToLower(address)).
		First(&userDB).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil  {
		err = errors.Wrap(err, "failed to get user from address from DB")
		return nil, err
	}

	return userDB, nil
}

func (r *Repository) AssociateAddress(userID int, address string) error {
	userDB, err := r.UserFromAddress(address)
	if err != nil {
		return errors.Wrap(err, "failed to associate address to user")
	} else if userDB == nil {
		err = r.DB.Create(&UserHasAddresses{
			AddressID: strings.ToLower(address),
			UserID:    userID,
		}).Error
		if err != nil {
			return errors.Wrap(err, "failed to associate address to user")
		}
		return nil
	} else {
		return errors.New("that address is already associated to other account")
	}
}
