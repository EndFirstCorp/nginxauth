package main

import (
	"errors"
	"time"
)

type Backender interface {
	// UserBackender. Write out since it contains duplicate BackendCloser
	AddUser(email string) error
	GetUser(email string) (*User, error)
	UpdateUser(email, fullname string, company string, pictureURL string) error

	// LoginBackender. Write out since it contains duplicate BackendCloser
	CreateLogin(email, passwordHash, fullName, homeDirectory string, uidNumber, gidNumber int, mailQuota, fileQuota string) (*UserLogin, error)
	GetLogin(email, loginProvider string) (*UserLogin, error)
	UpdateEmail(email string, password string, newEmail string) (*UserLoginSession, error)
	UpdatePassword(email string, oldPassword string, newPassword string) (*UserLoginSession, error)

	SessionBackender
}

type BackendCloser interface {
	Close() error
}

type UserBackender interface {
	AddUser(email string) error
	GetUser(email string) (*User, error)
	UpdateUser(email, fullname string, company string, pictureURL string) error
	BackendCloser
}

type LoginBackender interface {
	CreateLogin(email, passwordHash, fullName, homeDirectory string, uidNumber, gidNumber int, mailQuota, fileQuota string) (*UserLogin, error)
	GetLogin(email, loginProvider string) (*UserLogin, error)
	UpdateEmail(email string, password string, newEmail string) (*UserLoginSession, error)
	UpdatePassword(email string, oldPassword string, newPassword string) (*UserLoginSession, error)
	BackendCloser
}

type SessionBackender interface {
	CreateEmailSession(email, emailVerifyHash string) error
	GetEmailSession(verifyHash string) (*emailSession, error)
	DeleteEmailSession(verifyHash string) error

	CreateSession(email string, sessionHash string, sessionRenewTimeUTC, sessionExpireTimeUTC time.Time, rememberMe bool, rememberMeSelector, rememberMeTokenHash string, rememberMeRenewTimeUTC, rememberMeExpireTimeUTC time.Time) (*UserLoginSession, *UserLoginRememberMe, error)
	GetSession(sessionHash string) (*UserLoginSession, error)
	RenewSession(sessionHash string, renewTimeUTC time.Time) (*UserLoginSession, error)
	InvalidateSession(sessionHash string) error
	InvalidateSessions(email string) error

	GetRememberMe(selector string) (*UserLoginRememberMe, error)
	RenewRememberMe(selector string, renewTimeUTC time.Time) (*UserLoginRememberMe, error)
	InvalidateRememberMe(selector string) error
	BackendCloser
}

var errEmailVerifyHashExists = errors.New("DB: Email verify hash already exists")
var errInvalidEmailVerifyHash = errors.New("DB: Invalid verify code")
var errInvalidRenewTimeUTC = errors.New("DB: Invalid RenewTimeUTC")
var errInvalidSessionHash = errors.New("DB: Invalid SessionHash")
var errRememberMeSelectorExists = errors.New("DB: RememberMe selector already exists")
var errUserNotFound = errors.New("DB: User not found")
var errLoginNotFound = errors.New("DB: Login not found")
var errSessionNotFound = errors.New("DB: Session not found")
var errSessionAlreadyExists = errors.New("DB: Session already exists")
var errRememberMeNotFound = errors.New("DB: RememberMe not found")
var errRememberMeNeedsRenew = errors.New("DB: RememberMe needs to be renewed")
var errRememberMeExpired = errors.New("DB: RememberMe is expired")
var errUserAlreadyExists = errors.New("DB: User already exists")

type emailSession struct {
	Email           string
	EmailVerifyHash string
}

type User struct {
	FullName          string
	PrimaryEmail      string
	LockoutEndTimeUTC *time.Time
	AccessFailedCount int
}

type UserLogin struct {
	Email           string
	LoginProviderID int
	ProviderKey     string
}

type UserLoginSession struct {
	Email         string
	SessionHash   string
	RenewTimeUTC  time.Time
	ExpireTimeUTC time.Time
}

type UserLoginRememberMe struct {
	Email         string
	Selector      string
	TokenHash     string
	RenewTimeUTC  time.Time
	ExpireTimeUTC time.Time
}

type UserLoginProvider struct {
	LoginProviderID   int
	Name              string
	OAuthClientID     string
	OAuthClientSecret string
	OAuthURL          string
}

type AuthError struct {
	message    string
	innerError error
	shouldLog  bool
	error
}

func newLoggedError(message string, innerError error) *AuthError {
	return &AuthError{message: message, innerError: innerError, shouldLog: true}
}

func newAuthError(message string, innerError error) *AuthError {
	return &AuthError{message: message, innerError: innerError}
}

func (a *AuthError) Error() string {
	return a.message
}

func (a *AuthError) Trace() string {
	trace := a.message + "\n"
	indent := "  "
	inner := a.innerError
	for inner != nil {
		trace += indent + inner.Error() + "\n"
		e, ok := inner.(*AuthError)
		if !ok {
			break
		}
		indent += "  "
		inner = e.innerError
	}
	return trace
}

type backend struct {
	u UserBackender
	l LoginBackender
	s SessionBackender
	BackendCloser
}

func (b *backend) GetLogin(email, loginProvider string) (*UserLogin, error) {
	return b.l.GetLogin(email, loginProvider)
}

func (b *backend) CreateSession(email, sessionHash string, sessionRenewTimeUTC, sessionExpireTimeUTC time.Time, rememberMe bool, rememberMeSelector, rememberMeTokenHash string, rememberMeRenewTimeUTC, rememberMeExpireTimeUTC time.Time) (*UserLoginSession, *UserLoginRememberMe, error) {
	return b.s.CreateSession(email, sessionHash, sessionRenewTimeUTC, sessionExpireTimeUTC, rememberMe, rememberMeSelector, rememberMeTokenHash, rememberMeRenewTimeUTC, rememberMeExpireTimeUTC)
}

func (b *backend) GetSession(sessionHash string) (*UserLoginSession, error) {
	return b.s.GetSession(sessionHash)
}

func (b *backend) RenewSession(sessionHash string, renewTimeUTC time.Time) (*UserLoginSession, error) {
	return b.s.RenewSession(sessionHash, renewTimeUTC)
}

func (b *backend) GetRememberMe(selector string) (*UserLoginRememberMe, error) {
	return b.s.GetRememberMe(selector)
}

func (b *backend) RenewRememberMe(selector string, renewTimeUTC time.Time) (*UserLoginRememberMe, error) {
	return b.s.RenewRememberMe(selector, renewTimeUTC)
}

func (b *backend) CreateEmailSession(email, emailVerifyHash string) error {
	return b.s.CreateEmailSession(email, emailVerifyHash)
}

func (b *backend) GetEmailSession(emailVerifyHash string) (*emailSession, error) {
	return b.s.GetEmailSession(emailVerifyHash)
}

func (b *backend) DeleteEmailSession(emailVerifyHash string) error {
	return b.s.DeleteEmailSession(emailVerifyHash)
}

func (b *backend) AddUser(email string) error {
	return b.u.AddUser(email)
}

func (b *backend) GetUser(email string) (*User, error) {
	return b.u.GetUser(email)
}

func (b *backend) UpdateUser(email, fullname string, company string, pictureURL string) error {
	return b.u.UpdateUser(email, fullname, company, pictureURL)
}

func (b *backend) CreateLogin(email, passwordHash, fullName, homeDirectory string, uidNumber, gidNumber int, mailQuota, fileQuota string) (*UserLogin, error) {
	return b.l.CreateLogin(email, passwordHash, fullName, homeDirectory, uidNumber, gidNumber, mailQuota, fileQuota)
}

func (b *backend) UpdateEmail(email string, password string, newEmail string) (*UserLoginSession, error) {
	return b.l.UpdateEmail(email, password, newEmail)
}

func (b *backend) UpdatePassword(email string, oldPassword string, newPassword string) (*UserLoginSession, error) {
	return b.l.UpdatePassword(email, oldPassword, newPassword)
}

func (b *backend) InvalidateSession(sessionHash string) error {
	return b.s.InvalidateSession(sessionHash)
}

func (b *backend) InvalidateSessions(email string) error {
	return b.s.InvalidateSessions(email)
}

func (b *backend) InvalidateRememberMe(selector string) error {
	return b.s.InvalidateRememberMe(selector)
}

func (b *backend) Close() error {
	if err := b.s.Close(); err != nil {
		return err
	}
	if err := b.u.Close(); err != nil {
		return err
	}
	if err := b.l.Close(); err != nil {
		return err
	}
	return nil
}
