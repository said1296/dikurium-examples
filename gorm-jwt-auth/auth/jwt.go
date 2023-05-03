package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const (
	superkey = "jwt"
)

type Claims struct {
	UserID     int
	Expiration int64
}

// Type used to store to redis as hash
type JwtHash struct {
	UserID int
}

const (
	redisKeyPrefixJWT      = "JWT:"
	redisIndexPrefixUserId = "JWT:UserId:"
)



func (c *Claims) Valid() error {
	if valid, delta := c.ValidateExpiration(); !valid {
		return errors.New("token is expired by " + strconv.Itoa(int(delta)))
	}
	return nil
}

func (c *Claims) ValidateExpiration() (bool, int64) {
	now := time.Now().Unix()
	return now <= c.Expiration, c.Expiration - now
}

func (a *Auth) GenerateJWT(claims *Claims) (string, error) {
	// generate jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtString, err := token.SignedString([]byte(a.JwtSecret))
	if err != nil {
		return "", err
	}

	// save to redis
	err = a.RedisRepo.SaveHash(
		jwtRedisKey(jwtString),
		JwtHash{UserID: claims.UserID},
		[]string{userIdIndexKey(claims.UserID)},
		&a.TimeToLive,
	)
	if err != nil {
		return "", err
	}

	return jwtString, nil
}

func (a *Auth) CalculateExpiration() int64 {
	return time.Now().Add(a.TimeToLive).Unix()
}

func (a *Auth) ValidateAndGetClaims(jwtString string) (*Claims, error) {
	ok, err := a.RedisRepo.KeyExists(jwtRedisKey(jwtString))
	if err != nil {
		return nil, err
	} else if !ok {
		err = errors.New("token doesn't exist")
		return nil, err
	}

	token, err := jwt.ParseWithClaims(jwtString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.JwtSecret), nil
	})
	if err != nil {
		err = errors.Wrap(err, "failed to parse jwt")
	}

	claims, ok := token.Claims.(*Claims)
	if ok {
		if !token.Valid {
			err = errors.Wrap(err, "invalid token")
		}
	} else {
		err = errors.Wrap(err, "failed to cast claims")
		return nil, err
	}

	return claims, nil
}

func (a *Auth) RevokeToken(jwt string) error {
	return a.RedisRepo.DeleteKey(jwtRedisKey(jwt))
}

func (a *Auth) GetTTL(jwt string) (int, error) {
	return a.RedisRepo.GetTTL(jwtRedisKey(jwt))
}

func jwtRedisKey(jwt string) string {
	return redisKeyPrefixJWT + jwt
}

func userIdIndexKey(userId int) string {
	return redisIndexPrefixUserId + strconv.Itoa(userId)
}
