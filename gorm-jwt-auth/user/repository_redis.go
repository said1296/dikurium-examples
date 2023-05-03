package user

import (
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"server/internal/constant"
	"server/internal/email"
	"server/internal/mail"
	"server/pkg/mailparser"
	"strconv"
	"time"
)

const (
	confirmationKeyPrefix    = "Confirm:"

	passResetKeyPrefix       = "Reset:"
	passResetUserIDSetPrefix = passResetKeyPrefix + "UserID:"

	associateAddressKeyPrefix = "Associate:"

	loginBlockchainKeyPrefix = "Login:"
)

var (
	confirmationKeyTTL *time.Duration = nil
	passResetKeyTTL = 1 * time.Hour
	associateAddressKeyTTL = 2 * time.Minute
	loginBlockchainKeyTTL = 2 * time.Minute
)

type ResetHash struct {
	UserID int
}

type ConfirmationHash struct {
	UserID int
}

type AssociateAddressHash struct {
	Message string
	Address string
}

func (r *Repository) GenerateConfirmationKey(userID int) (*string, error) {
	// generate confirmation key
	confirmationKey := strconv.Itoa(int(time.Now().UnixNano())) + "-" + strconv.Itoa(userID)
	err := r.Redis.SaveHash(confirmationRedisKey(confirmationKey), ConfirmationHash{UserID: userID}, nil, confirmationKeyTTL)
	if err != nil {
		err = errors.Wrap(err, "failed to save confirmation key")
		return nil, err
	}

	return &confirmationKey, nil
}

func (r *Repository) GenerateAndSendConfirmationKey(userDB User) (err error) {
	// generate confirmation key
	confirmKey, err := r.GenerateConfirmationKey(userDB.ID)
	if err != nil {
		_ = r.Delete(userDB.ID)
		return errors.Wrap(err, "failed to generate confirmation key")
	}

	// send confirmation email
	emailTemplate := &email.Confirm{
		PreferredName:   userDB.PreferredName,
		ConfirmEmailURL: constant.ConfirmEmailURL + *confirmKey,
	}
	emailTemplate.Prepare()
	emailBytes, err := mailparser.Parse(emailTemplate)
	if err != nil {
		_ = r.Delete(userDB.ID)
		return errors.Wrap(err, "failed to parse confirmation email")
	}

	err = r.Mail.Send(mail.Email{
		From:    "no-reply@jevels.com",
		To:      userDB.Email,
		Subject: "Confirm your account",
		Message: string(emailBytes),
	})
	if err != nil {
		_ = r.Delete(userDB.ID)
		return errors.Wrap(err, "failed to send confirmation email")
	}

	return
}

func (r *Repository) GeneratePassResetKey(userID int) (*string, error) {
	// look for and delete previous keys
	passResetKeys, err := r.Redis.GetMembers(passResetUserIDSetKey(userID))
	if err != nil {
		return nil, err
	} else if len(passResetKeys)>0 {
		for _, passResetKey := range passResetKeys {
			err = r.RevokePassResetKey(userID, string(passResetKey.([]uint8)))
			if err != nil {
				return nil, err
			}
		}
	}

	// generate confirmation key
	confirmationKey := strconv.Itoa(int(time.Now().UnixNano())) + "-" + strconv.Itoa(userID)
	err = r.Redis.SaveHash(
		passResetRedisKey(confirmationKey),
		ResetHash{UserID: userID},
		[]string{passResetUserIDSetKey(userID)},
		&passResetKeyTTL,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to save confirmation key")
		return nil, err
	}

	return &confirmationKey, nil
}

func (r *Repository) GenerateAndSendPassResetKey(userDB User) (err error) {
	key, err := r.GeneratePassResetKey(userDB.ID)
	if err != nil {
		err = errors.Wrap(err, "failed to resolve ForgotPassword mutation")
		return err
	}


	emailTemplate := &email.ForgotPassword{
		PreferredName:    userDB.PreferredName,
		Key:     *key,
		URLBase: constant.ResetPasswordURL,
	}
	emailTemplate.Prepare()
	emailBytes, err := mailparser.Parse(emailTemplate)
	if err != nil {
		_ = r.RevokePassResetKey(userDB.ID, *key)
		return errors.Wrap(err, "failed to parse confirmation email")
	}

	err = r.Mail.Send(mail.Email{
		From:    "no-reply@jevels.com",
		To:      userDB.Email,
		Subject: "Forgot your password",
		Message: string(emailBytes),
	})
	if err != nil {
		_ = r.RevokePassResetKey(userDB.ID, *key)
		return errors.Wrap(err, "failed to send confirmation email")
	}

	return nil
}

func (r *Repository) GenerateAssociateAddressMessage(userID int, address string) (*string, error) {
	// look for and delete previous keys
	rawHash, err := r.Redis.GetHash(associateAddressRedisKey(userID))
	if err != nil {
		return nil, err
	} else if len(rawHash) == 0 {
		err = r.RevokeAssociateMessageKey(userID)
		if err != nil {
			return nil, err
		}
	}

	// generate confirmation key
	associateAddressMessage := strconv.Itoa(int(time.Now().UnixNano())) + "-" + strconv.Itoa(userID)
	err = r.Redis.SaveHash(
		associateAddressRedisKey(userID),
		AssociateAddressHash{
			Message: associateAddressMessage,
			Address: address,
		},
		nil,
		&associateAddressKeyTTL,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to save associate address message")
		return nil, err
	}

	return &associateAddressMessage, nil
}

func (r *Repository) GenerateLoginBlockchainMessage() (*string, error) {
	// generate confirmation key
	loginBlockchainMessage := uuid.NewString()

	err := r.Redis.Save(
		loginBlockchainRedisKey(loginBlockchainMessage),
		loginBlockchainMessage,
		nil,
		&loginBlockchainKeyTTL,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to save login blockchain message")
		return nil, err
	}

	return &loginBlockchainMessage, nil
}

func (r *Repository) LoginBlockchainMessage(key string) (*string, error) {
	loginBlockchainMessage, err := redis.String(r.Redis.Get(loginBlockchainRedisKey(key)))
	if errors.Is(err, redis.ErrNil) {
		return nil, errors.New("invalid message")
	} else if err != nil {
		return nil, err
	}
	return &loginBlockchainMessage, nil
}

func (r *Repository) AssociateAddressMessage(userID int) (*AssociateAddressHash, error) {
	rawHash, err := r.Redis.GetHash(associateAddressRedisKey(userID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get associate address message from Redis")
	} else if len(rawHash) == 0 {
		return nil, errors.New("associate address message no longer valid")
	}
	hash := AssociateAddressHash{}
	err = redis.ScanStruct(rawHash, &hash)
	return &hash, nil
}

var ErrConfirmationKeyNoLongerValid = errors.New("confirmation key is no longer valid")

func (r *Repository) UserIdFromConfirmationKey(key string) (*int, error) {
	rawHash, err := r.Redis.GetHash(confirmationRedisKey(key))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user id from confirmation key")
	} else if len(rawHash) == 0 {
		return nil, ErrConfirmationKeyNoLongerValid
	}
	hash := ConfirmationHash{}
	err = redis.ScanStruct(rawHash, &hash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse struct from Redis hash")
	}

	return &hash.UserID, nil
}

func (r *Repository) UserIDFromPassResetKey(key string) (*int, error) {
	rawHash, err := r.Redis.GetHash(passResetRedisKey(key))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user id from password reset key")
	} else if len(rawHash) == 0 {
		return nil, errors.New("password reset key is no longer valid")
	}

	hash := ResetHash{}
	err = redis.ScanStruct(rawHash, &hash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse struct from Redis hash")
	}

	userID := hash.UserID
	return &userID, nil
}

func (r *Repository) RevokeConfirmationKey(key string) error {
	return r.Redis.DeleteKey(confirmationRedisKey(key))
}

func (r *Repository) RevokePassResetKey(userID int, key string) error {
	return r.Redis.DeleteKeyAndSetMembership(passResetUserIDSetKey(userID), passResetRedisKey(key))
}

func (r *Repository) RevokeAssociateMessageKey(userID int) error {
	return r.Redis.DeleteKey(associateAddressRedisKey(userID))
}

func (r *Repository) RevokeLoginBlockchainMessageKey(key string) error {
	return r.Redis.DeleteKey(loginBlockchainRedisKey(key))
}

func confirmationRedisKey(key string) string {
	return confirmationKeyPrefix + key
}

func passResetRedisKey(key string) string {
	return passResetKeyPrefix + key
}

func passResetUserIDSetKey(userID int) string {
	return passResetUserIDSetPrefix + strconv.Itoa(userID)
}

func associateAddressRedisKey(userID int) string {
	return associateAddressKeyPrefix + strconv.Itoa(userID)
}

func loginBlockchainRedisKey(key string) string {
	return loginBlockchainKeyPrefix + key
}
