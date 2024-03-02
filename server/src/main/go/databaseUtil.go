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

func (s *Server) isAuthorized(email string) (bool, error) {
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
		return false, err
	}
	defer authdbConn.Close(context.Background())
	var res string
	err = authdbConn.QueryRow(context.Background(), "SELECT role FROM authorized WHERE email=$1", email).Scan(&res)
	return err == nil, nil
}

func (s *Server) insertCredentialsIntoCredsDb(email string, site string, userId string, encryptedPwd string) error {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.credsDbUsername,
		s.credsDbPassword,
		s.credsDbHost,
		s.credsDbPort,
		s.credsDbName)

	// Connect using pgx
	credsdbConn, err := pgx.Connect(context.Background(), url)

	if err != nil {
		return err
	}
	defer credsdbConn.Close(context.Background())
	insertSQL := `INSERT INTO credentials (email, site, userid, password) VALUES ($1, $2, $3, $4)`
	_, err = credsdbConn.Exec(context.Background(), insertSQL, email, site, userId, encryptedPwd)
	return err
}

func (s *Server) updateCredentialsIntoCredsDb(email string, site string, userId string, encryptedPwd string) error {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.credsDbUsername,
		s.credsDbPassword,
		s.credsDbHost,
		s.credsDbPort,
		s.credsDbName)

	// Connect using pgx
	credsdbConn, err := pgx.Connect(context.Background(), url)

	if err != nil {
		return err
	}
	defer credsdbConn.Close(context.Background())
	updateSQL := `UPDATE credentials SET userid=$3, password=$4 WHERE email=$1 AND site=$2`
	_, err = credsdbConn.Exec(context.Background(), updateSQL, email, site, userId, encryptedPwd)
	return err
}

func (s *Server) deleteCredentialsFromDatabase(email string, site string) error {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.credsDbUsername,
		s.credsDbPassword,
		s.credsDbHost,
		s.credsDbPort,
		s.credsDbName)

	// Connect using pgx
	credsdbConn, err := pgx.Connect(context.Background(), url)

	if err != nil {
		return err
	}
	defer credsdbConn.Close(context.Background())

	deleteSQL := `DELETE FROM credentials WHERE email=$1 AND site=$2`
	_, err = credsdbConn.Exec(context.Background(), deleteSQL, email, site)
	return err
}
