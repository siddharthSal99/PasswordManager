package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) decryptString(cipherText string) (string, error) {
	data, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(s.encryptionKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherTextData := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plainTextData, err := gcm.Open(nil, nonce, cipherTextData, nil)
	if err != nil {
		return "", err
	}

	return string(plainTextData), nil
}

func (s *Server) retrieveCredentialsFromRedis(rds *redis.Client, key string) (string, string, error) {

	kv, err := s.retrieveFromCache(context.Background(), rds, key)
	if err != nil {
		return "", "", err
	}
	storedPwdCipher := kv["pwdCipher"]
	userid := kv["userid"]

	password, err := s.decryptString(storedPwdCipher)
	if err != nil {
		return "", "", err
	}
	return userid, password, nil
}

func (s *Server) connectToValidationCodeCache() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     s.credsCacheHost + ":" + s.credsCachePort, // Assuming Redis is running on localhost
		Password: s.credsCachePassword,                      // No password set
		DB:       0,                                         // Use default DB
	})
}

func (s *Server) connectToAuthTokenCache() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     s.authTokenCacheHost + ":" + s.authTokenCachePort, // Assuming Redis is running on localhost
		Password: s.authTokenCachePassword,                          // No password set
		DB:       0,                                                 // Use default DB
	})
}

func (s *Server) setRedisTTL(ctx context.Context, rdb *redis.Client, key string, d time.Duration) error {
	_, err := rdb.Expire(ctx, key, d).Result()
	return err
}

func (s *Server) storeValidationCode(ctx context.Context,
	rdb *redis.Client,
	key string,
	validationCode string) error {
	// return rdb.Set(ctx, key, value, d).Err()
	_, err := rdb.HSet(ctx, key, "validationCode", validationCode).Result()
	if err != nil {
		return err
	}
	_, err = rdb.Expire(ctx, key, s.validationCodeDuration).Result()
	return err

}

func (s *Server) storeInAuthTokenCache(ctx context.Context,
	rdb *redis.Client,
	key string,
	authToken string) error {
	// return rdb.Set(ctx, key, value, d).Err()
	_, err := rdb.HSet(ctx, key, "authToken", authToken).Result()
	return err

}

func (s *Server) retrieveFromAuthTokenCache(rdb *redis.Client, key string) (map[string]string, error) {
	return s.retrieveFromCache(context.Background(), rdb, key)
}

func (s *Server) retrieveFromCredsCache(rdb *redis.Client, key string) (map[string]string, error) {
	return s.retrieveFromCache(context.Background(), rdb, key)
}

func (s *Server) retrieveFromCache(ctx context.Context, rdb *redis.Client, key string) (map[string]string, error) {
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
