package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func getCredentials(conn *pgx.Conn, email string, site string) (string, string, error) {
	var userid string
	var password string
	err := conn.QueryRow(context.Background(), "SELECT userid, password FROM credentials WHERE email=$1 AND site=$2", email, site).Scan(&userid, &password)

	if err != nil {
		return "", "", err
	}
	return userid, password, nil
}

func (s *Server) getCredentialsFromDatabase(email string, site string) (string, string, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.credsDbUsername,
		s.credsDbPassword,
		s.credsDbHost,
		s.credsDbPort,
		s.credsDbName)

	// Connect using pgx
	credsdbConn, err := pgx.Connect(context.Background(), url)

	if err != nil {
		return "", "", err
	}
	defer credsdbConn.Close(context.Background())

	userid, pwdCipher, err := getCredentials(credsdbConn, email, site)

	if err != nil {
		return "", "", err
	}
	return userid, pwdCipher, nil
}
