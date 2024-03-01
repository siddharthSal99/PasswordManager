package main

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func (s *Server) connectToValidationCodeCache() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     s.credsCacheHost + ":" + s.credsCachePort, // Assuming Redis is running on localhost
		Password: s.credsCachePassword,                      // No password set
		DB:       0,                                         // Use default DB
	})
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

func (s *Server) retrieveValidationCode(rdb *redis.Client, email string, site string) (map[string]string, error) {
	return s.retrieveFromCache(context.Background(), rdb, email+":"+site)
}

func (s *Server) retrieveFromCache(ctx context.Context, rdb *redis.Client, key string) (map[string]string, error) {
	kv, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	} else {
		return kv, nil
	}
}

/*
	Use the below auth token cache for cookie-based auth
*/

// func (s *Server) connectToAuthTokenCache() *redis.Client {
// 	return redis.NewClient(&redis.Options{
// 		Addr:     s.authTokenCacheHost + ":" + s.authTokenCachePort, // Assuming Redis is running on localhost
// 		Password: s.authTokenCachePassword,                          // No password set
// 		DB:       0,                                                 // Use default DB
// 	})
// }

// func (s *Server) storeInAuthTokenCache(ctx context.Context,
// 	rdb *redis.Client,
// 	key string,
// 	authToken string) error {
// 	// return rdb.Set(ctx, key, value, d).Err()
// 	_, err := rdb.HSet(ctx, key, "authToken", authToken).Result()
// 	return err

// }

// func (s *Server) retrieveFromAuthTokenCache(rdb *redis.Client, key string) (map[string]string, error) {
// 	return s.retrieveFromCache(context.Background(), rdb, key)
// }
