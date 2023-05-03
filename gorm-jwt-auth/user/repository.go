package user

import (
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/api/graphql/graph"
	"server/internal/mail"
	"server/internal/redisrepo"
	"strings"
)

type Repository struct {
	DB    *gorm.DB
	Redis *redisrepo.RedisRepo
	Mail  *mail.Mail
}

func addUserFilters(filter graph.UsersFilter, query *gorm.DB) *gorm.DB {
	// Where
	if filter.Ids != nil {
		query = query.Where(DBNamesUser.ID+" IN ?", filter.Ids)
	}

	if filter.Name != nil {
		query = query.
			Where(DBNamesUser.FirstName+" ILIKE ?", "%"+*filter.Name+"%").
			Or(DBNamesUser.LastName+" ILIKE ?", "%"+*filter.Name+"%").
			Or(DBNamesUser.PreferredName+" ILIKE ?", "%"+*filter.Name+"%")
	}

	if filter.Email != nil {
		query = query.
			Where(DBNamesUser.Email+" ILIKE ?", "%"+*filter.Email+"%")
	}

	if len(filter.Roles) > 0 {
		query = query.Joins("INNER JOIN "+DBNamesUserHasRoles.TableName+" ON "+DBNamesUserHasRoles.UserID+" = "+DBNamesUser.ID+" AND "+DBNamesUserHasRoles.RoleID+" IN ?", filter.Roles)
	}

	if filter.HasBankAccount != nil && *filter.HasBankAccount {
		query = query.Where(DBNamesUser.HasBankAccount, true)
	}

	if filter.HasCompleteProfile != nil && *filter.HasCompleteProfile {
		query = query.Where(DBNamesUser.HasCompleteProfile, true)
	}

	if filter.HasUploadedOneNft != nil && *filter.HasUploadedOneNft {
		query = query.Where(DBNamesUser.HasUploadedOneNft, true)
	}

	if filter.StripeTransferCapabilityStatusActive != nil && *filter.StripeTransferCapabilityStatusActive {
		query = query.Where(DBNamesUser.StripeTransferCapabilityStatus, stripe.CapabilityStatusActive)
	}

	if filter.Confirmed != nil && *filter.Confirmed {
		query = query.Where(DBNamesUser.RegisterTime + " IS NOT NULL")
	}

	// Order By
	if filter.OrderBy != nil {
		if filter.OrderBy.ID != nil {
			query = query.Order(DBNamesUser.ID + " " + filter.OrderBy.ID.String())
		}

		if filter.OrderBy.PreferredName != nil {
			query = query.Order(DBNamesUser.PreferredName + " " + filter.OrderBy.PreferredName.String())
		}

		if filter.OrderBy.LastName != nil {
			query = query.Order(DBNamesUser.LastName + " " + filter.OrderBy.LastName.String())
		}
	}

	// Pagination
	if filter.Pagination != nil {
		query = query.Limit(filter.Pagination.Limit).
			Offset((filter.Pagination.Page - 1) * filter.Pagination.Limit)
	}

	return query
}

func (r *Repository) GetUsers(filter graph.UsersFilter) ([]*User, error) {
	var usersDB []*User

	query := r.DB

	query = addUserFilters(filter, query)

	// Perform query
	err := query.Find(&usersDB).Error
	if err != nil {
		err = errors.Wrap(err, "failed to get users from DB")
		return nil, err
	}

	return usersDB, nil
}

func (r *Repository) Count(filter graph.UsersFilter) (int, error) {
	var count int64

	query := r.DB.Model(&User{})

	filter.OrderBy = nil
	filter.Pagination = nil

	query = addUserFilters(filter, query)

	// Perform query
	err := query.Count(&count).Error
	if err != nil {
		err = errors.Wrap(err, "failed to get users from DB")
		return 0, err
	}

	return int(count), nil
}

func (r *Repository) GetProfile(userID int) (*Profile, error) {
	var profile Profile
	err := r.DB.First(&profile, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get profile from DB")
	}
	return &profile, nil
}

func (r *Repository) UpsertProfile(userID int, input *graph.ProfileInput) (*Profile, error) {
	var toUpdate []string

	var profile = Profile{
		UserID: userID,
	}
	if input.Description != nil {
		profile.Description = strings.TrimSpace(*input.Description)
		toUpdate = append(toUpdate, DBNamesProfile.Description)
	}

	err := r.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
		DoUpdates: clause.AssignmentColumns(toUpdate),
	}).Create(&profile).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to update profile")
	}

	return &profile, nil
}

func (r *Repository) Create(input graph.CreateUserInput) (*User, error) {
	usersDB, err := r.GetUsers(graph.UsersFilter{
		Email: &input.Email,
	})
	if err != nil {
		err = errors.Wrap(err, "failed to check if email exists")
		return nil, err
	}

	if len(usersDB) > 0 {
		return nil, errors.New("email already registered")
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		err = errors.Wrap(err, "failed to hash password")
		return nil, err
	}

	userDB := &User{
		FirstName:     input.FirstName,
		LastName:      input.LastName,
		PreferredName: input.PreferredName,
		Password:      passwordHash,
		Email:         input.Email,
		RegisterTime:  nil,
		Roles: []*Role{
			{
				ID: RoleUserID,
			},
		},
		Country: input.Country,
	}

	r.DB.Create(&userDB)

	return userDB, nil
}

func (r * Repository) Save(userDB *User) (*User, error) {
	err := r.DB.Save(&userDB).Error
	if err != nil {
		return nil, err
	}

	return userDB, nil
}

func (r *Repository) Delete(userID int) error {
	err := r.DB.Where(DBNamesUserHasRoles.UserID, userID).Delete(&UserHasRoles{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to delete roles from user from DB")
	}
	err = r.DB.Where(DBNamesUserHasAddresses.UserID, userID).Delete(&UserHasAddresses{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to delete addresses from user from DB")
	}
	err = r.DB.Delete(&User{}, userID).Error
	if err != nil {
		return errors.Wrap(err, "failed to delete user from DB")
	}

	return nil
}

// Helpers

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func validatePasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
