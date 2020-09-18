package service

import (
	"context"
	"crypto/x509"
	"net/http"

	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"
)

// AuthorizerdMock is a mock of Authorizerd
type AuthorizerdMock struct {
	InitFunc              func(context.Context) error
	StartFunc             func(context.Context) <-chan error
	VerifyFunc            func(r *http.Request, act, res string) (authorizerd.Principal, error)
	VerifyAccessTokenFunc func(ctx context.Context, tok, act, res string, cert *x509.Certificate) (authorizerd.Principal, error)
	VerifyRoleTokenFunc   func(ctx context.Context, tok, act, res string) (authorizerd.Principal, error)
	VerifyRoleJWTFunc     func(ctx context.Context, tok, act, res string) error
	VerifyRoleCertFunc    func(ctx context.Context, peerCerts []*x509.Certificate, act, res string) (authorizerd.Principal, error)
	GetPolicyCacheFunc    func(ctx context.Context) map[string]interface{}
}

// Init is a mock implementation of Authorizerd.Init
func (am *AuthorizerdMock) Init(ctx context.Context) error {
	return am.InitFunc(ctx)
}

// Start is a mock implementation of Authorizerd.Start
func (am *AuthorizerdMock) Start(ctx context.Context) <-chan error {
	return am.StartFunc(ctx)
}

// Verify is a mock implementation of Authorizerd.Verify
func (am *AuthorizerdMock) Verify(r *http.Request, act, res string) error {
	_, err := am.VerifyFunc(r, act, res)
	return err
}

// Authorize is a mock implementation of Authorizerd.Authorize
func (am *AuthorizerdMock) Authorize(r *http.Request, act, res string) (authorizerd.Principal, error) {
	return am.VerifyFunc(r, act, res)
}

// VerifyAccessToken is a mock implementation of Authorizerd.VerifyAccessToken
func (am *AuthorizerdMock) VerifyAccessToken(ctx context.Context, tok, act, res string, cert *x509.Certificate) error {
	_, err := am.VerifyAccessTokenFunc(ctx, tok, act, res, cert)
	return err
}

// AuthorizeAccessToken is a mock implementation of Authorizerd.AuthorizeAccessToken
func (am *AuthorizerdMock) AuthorizeAccessToken(ctx context.Context, tok, act, res string, cert *x509.Certificate) (authorizerd.Principal, error) {
	return am.VerifyAccessTokenFunc(ctx, tok, act, res, cert)
}

// VerifyRoleToken is a mock implementation of Authorizerd.VerifyRoleToken
func (am *AuthorizerdMock) VerifyRoleToken(ctx context.Context, tok, act, res string) error {
	_, err := am.VerifyRoleTokenFunc(ctx, tok, act, res)
	return err
}

// AuthorizeRoleToken is a mock implementation of Authorizerd.AuthorizeRoleToken
func (am *AuthorizerdMock) AuthorizeRoleToken(ctx context.Context, tok, act, res string) (authorizerd.Principal, error) {
	return am.VerifyRoleTokenFunc(ctx, tok, act, res)
}

// VerifyRoleJWT is a mock implementation of Authorizerd.VerifyRoleJWT
func (am *AuthorizerdMock) VerifyRoleJWT(ctx context.Context, tok, act, res string) error {
	return am.VerifyRoleJWTFunc(ctx, tok, act, res)
}

// VerifyRoleCert is a mock implementation of Authorizerd.VerifyRoleCert
func (am *AuthorizerdMock) VerifyRoleCert(ctx context.Context, peerCerts []*x509.Certificate, act, res string) error {
	_, err := am.VerifyRoleCertFunc(ctx, peerCerts, act, res)
	return err
}

// AuthorizeRoleCert is a mock implementation of Authorizerd.AuthorizeRoleCert
func (am *AuthorizerdMock) AuthorizeRoleCert(ctx context.Context, peerCerts []*x509.Certificate, act, res string) (authorizerd.Principal, error) {
	return am.VerifyRoleCertFunc(ctx, peerCerts, act, res)
}

// GetPolicyCache is a mock implementation of Authorizerd.GetPolicyCache
func (am *AuthorizerdMock) GetPolicyCache(ctx context.Context) map[string]interface{} {
	return am.GetPolicyCacheFunc(ctx)
}
