package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// decryptString decrypts a string with the given key
func decryptString(cipherText, key string) (string, error) {
	data, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
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

func isAuthorized(conn *pgx.Conn, email string) bool {
	// log.Println("email:", email)
	var res string
	err := conn.QueryRow(context.Background(), "SELECT role FROM authorized WHERE email=$1", email).Scan(&res)
	// if err != nil {
	// 	log.Println("Error:", err)
	// }
	return err == nil
}

func getCredentials(conn *pgx.Conn, email string, site string) (string, string, error) {
	var userid string
	var password string
	err := conn.QueryRow(context.Background(), "SELECT userid, password FROM credentials WHERE email=$1 AND site=$2", email, site).Scan(&userid, &password)

	if err != nil {
		return "", "", err
	}
	return userid, password, nil
}

func (s *Server) GetPasswordIfExists(c *gin.Context) {
	var email string
	var site string
	/*
		Extract the query parameters for email and site
	*/
	if val, ok := c.GetQuery("email"); ok {
		email = val
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "No email provided",
		})
		return
	}

	if val, ok := c.GetQuery("site"); ok {
		site = val
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "No site provided",
		})
		return
	}

	/*
		Connect to auth db, check if the email exists in the auth db
	*/
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.authDbUsername,
		s.authDbPassword,
		s.authDbHost,
		s.authDbPort,
		s.authDbName)

	// Connect using pgx
	authdbConn, err := pgx.Connect(context.Background(), url)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - database connection",
		})
		return
	}
	defer authdbConn.Close(context.Background())

	// Check if email is in auth db
	if !isAuthorized(authdbConn, email) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "this email is not authorized to access this resource",
		})
		return
	}

	/*
		Get uid and password from creds db and make sure it exists.
	*/
	url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.credsDbUsername,
		s.credsDbPassword,
		s.credsDbHost,
		s.credsDbPort,
		s.credsDbName)

	// Connect using pgx
	credsdbConn, err := pgx.Connect(context.Background(), url)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - database connection",
		})
		return
	}
	defer credsdbConn.Close(context.Background())

	userid, pwdCipher, err := getCredentials(credsdbConn, email, site)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": fmt.Sprintf("Credentials not found for email: %s and site: %s", email, site),
		})
		return
	}

	// c.JSON(http.StatusOK, gin.H{
	// 	"msg": fmt.Sprintf("Credentials: username: %s, password: %s ", userid, pwd),
	// })

	/*
		Validate the email with a code
	*/
	// Generate the code
	code := ""
	for i := 0; i < 6; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}
	fmt.Println("code:", code)

	// fmt.Sprintf("%s, %s", password, userid)

	// Connect to redis
	rds := connectToRedis(fmt.Sprintf("%s:%s", s.redisHost, s.redisPort), s.redisPassword, 0)
	defer rds.Close()

	// store the code, email, site, username, password in redis
	key := email + ":" + site
	err = storeRedis(context.Background(), rds, key, hashAndSalt([]byte(code)), pwdCipher, userid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - validation code generation",
		})
		return
	}

	if err = setRedisTTL(context.Background(), rds, key, 2*time.Minute); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - validation code generation",
		})
		return
	}

	/*
		This code will really be in the /password/validate endpoint
	*/

	/*
		Respond with the retrieved password after validation
	*/
	c.JSON(http.StatusOK, gin.H{
		"msg": "Code generated",
	})

}
