package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// encryptString encrypts a string with the given key
func encryptString(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

func insertCredentialsIntoCredsDb(conn *pgx.Conn, email string, site string, userId string, encryptedPwd string) error {
	insertSQL := `INSERT INTO credentials (email, site, userid, password) VALUES ($1, $2, $3, $4)`
	_, err := conn.Exec(context.Background(), insertSQL, email, site, userId, encryptedPwd)
	return err
}

func (s *Server) CreatePasswordEntry(c *gin.Context) {

	// Define a map to hold the request body
	var jsonMap map[string]string

	// Bind the JSON to the map
	if err := c.BindJSON(&jsonMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract the fields from the post body
	var email string
	var site string
	var userid string
	var password string
	var ok bool

	if email, ok = jsonMap["email"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no email provided in post body"})
	}

	if site, ok = jsonMap["site"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no site provided in post body"})
	}

	if userid, ok = jsonMap["userid"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no userid provided in post body"})
	}

	if password, ok = jsonMap["password"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no password provided in post body"})
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

	url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.credsDbUsername,
		s.credsDbPassword,
		s.credsDbHost,
		s.credsDbPort,
		s.credsDbName)

	credsdbConn, err := pgx.Connect(context.Background(), url)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - database connection",
		})
		return
	}
	defer credsdbConn.Close(context.Background())
	encryptedPassword, err := encryptString(password, s.encryptionKey)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - encryption error",
		})
		return
	}

	err = insertCredentialsIntoCredsDb(credsdbConn, email, site, userid, encryptedPassword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Server Error - database error",
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"msg": fmt.Sprintf("Success - added credentials for email: %s, site: %s", email, site),
	})

}
