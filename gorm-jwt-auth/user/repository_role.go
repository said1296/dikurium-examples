package user

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/api/graphql/graph"
)

type RoleFilter struct {
	ID   int
	Name int
}

func (r *Repository) GetRoles(filter graph.RolesFilter) ([]*Role, error) {
	var rolesDB []*Role

	query := r.DB

	// Where
	if filter.ID != nil {
		query = query.Where(DBNamesRole.ID, filter.ID)
	}

	if filter.Name != nil {
		query = query.
			Where(DBNamesRole.Name+" ILIKE ?", "%"+*filter.Name+"%")
	}

	if filter.User != nil {
		if filter.User.ID != nil {
			query = query.Joins("INNER JOIN " + DBNamesUserHasRoles.TableName + " ON " + DBNamesUserHasRoles.UserID + " = ? AND " +
				DBNamesUserHasRoles.RoleID + " = " + DBNamesRole.ID, filter.User.ID)
		}

	}

	// Perform query
	err := query.Find(&rolesDB).Error
	if err != nil {
		err = errors.Wrap(err, "failed to get role from DB")
		return nil, err
	}

	return rolesDB, nil
}

func (r *Repository) GetUsersWithRole(role *Role) ([]*User, error) {
	var roleDB Role
	err := r.DB.Where(role).Preload(DBNamesRole.Users).First(&roleDB).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("role doesn't exist")
	}

	return roleDB.Users, nil
}

func (r Repository) SetRole(userID int, roleID int, activate bool) error {
	// Check if User exists
	var userDB *User
	err := r.DB.First(&userDB, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user doesn't exist")
	}

	// Check if Role exists
	var roleDB *Role
	err = r.DB.First(&roleDB, roleID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("role doesn't exist")
	}

	userHasRole := &UserHasRoles{
		UserID: userDB.ID,
		RoleID: roleDB.ID,
	}

	if activate {
		r.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&userHasRole)
	} else {
		r.DB.Where(&userHasRole).Delete(&UserHasRoles{})
	}

	return err
}
