package service

import (
	"github.com/go-redis/redis/v8"
	"xisu/backend-go/internal/service/jwxt"
)

// Backward-compatible aliases after extracting JWXT code into service/jwxt.
type SerializableCookie = jwxt.SerializableCookie

type CachedJWXTSession = jwxt.CachedJWXTSession

type JwxtDirectService = jwxt.JwxtDirectService

func NewJwxtDirectService(redisClient *redis.Client) *JwxtDirectService {
	return jwxt.NewJwxtDirectService(redisClient)
}
