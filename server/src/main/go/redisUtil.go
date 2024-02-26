package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

func connectToRedis(addr string, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,     // Assuming Redis is running on localhost
		Password: password, // No password set
		DB:       db,       // Use default DB
	})
}

func setRedisTTL(ctx context.Context, rdb *redis.Client, key string, d time.Duration) error {
	_, err := rdb.Expire(ctx, key, d).Result()
	return err
}

func storeRedis(ctx context.Context,
	rdb *redis.Client,
	key string,
	validationCode string,
	pwdCipher string,
	userid string) error {
	// return rdb.Set(ctx, key, value, d).Err()
	_, err := rdb.HSet(ctx, key, "validationCode", validationCode, "pwdCipher", pwdCipher, "userid", userid).Result()
	return err

}

func retrieveRedis(ctx context.Context, rdb *redis.Client, key string) (map[string]string, error) {
	kv, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	} else {
		return kv, nil
	}
}

func hashAndSalt(pwd []byte) string {

	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
