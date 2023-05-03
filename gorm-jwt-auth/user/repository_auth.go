package user

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"server/api/graphql/graph"
	"time"
)

func (r *Repository) ChangePassword(userID int, password string) error {
	passwordHash, err := hashPassword(password)
	if err != nil {
		err = errors.Wrap(err, "failed to hash password")
		return err
	}

	return r.DB.Model(&User{}).Where(DBNamesUser.ID, userID).Update(DBNamesUser.Password, passwordHash).Error
}

func (r *Repository) Login(email string, password string) (*User, error) {
	usersDB, err := r.GetUsers(graph.UsersFilter{
		Email: &email,
	})

	if err != nil {
		return nil, err
	} else if len(usersDB) == 0 {
		return nil, errors.New("email is not registered")
	}

	err = validatePasswordHash(password, usersDB[0].Password)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, errors.New("incorrect password")
	} else if err != nil {
		err = errors.Wrap(err, "failed to login")
		return nil, err
	}

	return usersDB[0], nil
}

func (r * Repository) ConfirmUser(userID int) error {
	return r.DB.Model(&User{}).Where(DBNamesUser.ID, userID).Update(DBNamesUser.RegisterTime, time.Now()).Error
}
