package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) retrieveCredentialsFromRedis(rds *redis.Client, key string) (string, string, error) {

	kv, err := s.retrieveRedis(context.Background(), rds, key)
	storedPwdCipher := kv["pwdCipher"]
	userid := kv["userid"]

	password, err := s.decryptString(storedPwdCipher)
	if err != nil {
		return "", "", nil
	}
	return userid, password, nil
}

func (s *Server) connectToRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     s.redisHost + ":" + s.redisPort, // Assuming Redis is running on localhost
		Password: s.redisPassword,                 // No password set
		DB:       0,                               // Use default DB
	})
}

func (s *Server) setRedisTTL(ctx context.Context, rdb *redis.Client, key string, d time.Duration) error {
	_, err := rdb.Expire(ctx, key, d).Result()
	return err
}

func (s *Server) storeRedis(ctx context.Context,
	rdb *redis.Client,
	key string,
	validationCode string,
	pwdCipher string,
	userid string) error {
	// return rdb.Set(ctx, key, value, d).Err()
	_, err := rdb.HSet(ctx, key, "validationCode", validationCode, "pwdCipher", pwdCipher, "userid", userid).Result()
	return err

}

func (s *Server) retrieveRedis(ctx context.Context, rdb *redis.Client, key string) (map[string]string, error) {
	kv, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	} else {
		return kv, nil
	}
}

func (s *Server) hashAndSalt(pwd []byte) string {

	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}
