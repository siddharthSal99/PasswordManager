package main

import (
	"context"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func isAuthorized(conn *pgx.Conn, email string) bool {
	var res string
	err := conn.QueryRow(context.Background(), "SELECT role FROM authorized WHERE email=$1", email).Scan(&res)
	return err == nil
}

func (s *Server) createAndStoreCode(email string, site string) (string, error) {
	// Generate the code
	code := ""
	for i := 0; i < 6; i++ {
		code += strconv.Itoa(mathrand.Intn(10))
	}

	// Connect to redis
	credsRds := s.connectToValidationCodeCache()
	defer credsRds.Close()

	// store the code, email, site in redis
	key := email + ":" + site
	saltedCode, err := s.hashAndSalt([]byte(code))
	if err != nil {
		return "", err
	}
	err = s.storeValidationCode(context.Background(), credsRds, key, saltedCode)
	if err != nil {
		return "", err
	}
	return code, nil
}

func sendCode(code string, email string) {
	// TODO: Implement email service send code.
	fmt.Println("code:", code, "email:", email)
}

func (s *Server) AuthorizeAndSendCode(c *gin.Context) {
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
			"msg": "Server Error",
		})
		return
	}
	defer authdbConn.Close(context.Background())

	// Check if email is in auth db
	if !isAuthorized(authdbConn, email) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "this email is not authorized to access this resource",
		})
	} else {
		code, err := s.createAndStoreCode(email, site)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "Server Error",
			})
		}
		sendCode(code, email)
		c.JSON(http.StatusOK, gin.H{
			"msg": "code sent",
		})
	}
}
