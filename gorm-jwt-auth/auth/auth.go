package auth

import (
	"os"
	"server/internal/redisrepo"
	"time"
)

type Auth struct {
	RedisRepo *redisrepo.RedisRepo
	JwtSecret string
	TimeToLive time.Duration
}

func NewAuth(redisRepo *redisrepo.RedisRepo) *Auth {
	auth := &Auth{
		RedisRepo: redisRepo,
	}
	auth.Init()
	return auth
}

func (a *Auth) Init() {
	a.JwtSecret = os.Getenv("JWT_SECRET")
	a.TimeToLive = (24 * time.Hour) * 7
}
