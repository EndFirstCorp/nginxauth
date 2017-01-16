package main

import (
	"errors"
	"github.com/robarchibald/onedb"
)

type backendDbUser struct {
	Db onedb.DBer

	GetUserLoginQuery string
	AddUserQuery      string
	VerifyEmailQuery  string
	UpdateUserQuery   string
	CreateLoginQuery  string
}

func newBackendDbUser(server string, port int, username, password, database string, getUserLoginQuery, addUserQuery, verifyEmailQuery, updateUserQuery string) (UserBackender, error) {
	db, err := onedb.NewPgx(server, uint16(port), username, password, database)
	if err != nil {
		return nil, err
	}
	return &backendDbUser{Db: db,
		GetUserLoginQuery: getUserLoginQuery,
		AddUserQuery:      addUserQuery,
		VerifyEmailQuery:  verifyEmailQuery,
		UpdateUserQuery:   updateUserQuery}, nil
}

func (u *backendDbUser) GetLogin(email, loginProvider string) (*UserLogin, error) {
	var login *UserLogin
	return login, u.Db.QueryStruct(onedb.NewSqlQuery(u.GetUserLoginQuery, email, loginProvider), login)
}

func (u *backendDbUser) AddUser(email, emailVerifyHash string) error {
	return u.Db.Execute(onedb.NewSqlQuery(u.AddUserQuery, email, emailVerifyHash))
}

func (u *backendDbUser) VerifyEmail(emailVerifyHash string) (string, error) {
	var user *User
	err := u.Db.QueryStructRow(onedb.NewSqlQuery(u.VerifyEmailQuery, emailVerifyHash), user)
	if err != nil || user == nil {
		return "", errors.New("Unable to verify email: " + err.Error())
	}
	return user.PrimaryEmail, err
}

func (u *backendDbUser) UpdateUser(emailVerifyHash, fullname string, company string, pictureURL string) (string, error) {
	return "email", nil
}

func (u *backendDbUser) Close() error {
	return nil
}