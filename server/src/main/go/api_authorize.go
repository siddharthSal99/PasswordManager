package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// decryptString decrypts a string with the given key
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

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (s *Server) Authorize(c *gin.Context) {
	var email string
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
		c.SetCookie("sspassman-auth", "", -1, "/", "", true, true)
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "this email is not authorized to access this resource",
		})
	} else {
		token := GenerateSecureToken(16)
		rds := s.connectToAuthTokenCache()
		defer rds.Close()
		err = s.storeInAuthTokenCache(context.Background(), rds, email, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "Server Error - database connection",
			})
			return
		}
		err = s.setRedisTTL(context.Background(), rds, email, 600)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "Server Error - database connection",
			})
			return
		}
		c.SetCookie("sspassman-auth", token, 600, "/", "", true, true)
		c.JSON(http.StatusOK, gin.H{
			"msg": "authorized",
		})
	}

}
