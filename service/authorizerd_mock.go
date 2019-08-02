package service

import (
	"context"
	"crypto/x509"
)

// AuthorizerdMock is a mock of Authorizerd
type AuthorizerdMock struct {
	StartFunc           func(context.Context) <-chan error
	VerifyRoleTokenFunc func(ctx context.Context, tok, act, res string) error
	VerifyRoleJWTFunc   func(ctx context.Context, tok, act, res string) error
	VerifyRoleCertFunc  func(ctx context.Context, peerCerts []*x509.Certificate, act, res string) error
	GetPolicyCacheFunc  func(ctx context.Context) map[string]interface{}
}

// Start is a mock implementation of Authorizerd.Start
func (am *AuthorizerdMock) Start(ctx context.Context) <-chan error {
	return am.StartFunc(ctx)
}

// VerifyRoleToken is a mock implementation of Authorizerd.VerifyRoleToken
func (am *AuthorizerdMock) VerifyRoleToken(ctx context.Context, tok, act, res string) error {
	return am.VerifyRoleTokenFunc(ctx, tok, act, res)
}

// VerifyRoleJWT is a mock implementation of Authorizerd.VerifyRoleJWT
func (am *AuthorizerdMock) VerifyRoleJWT(ctx context.Context, tok, act, res string) error {
	return am.VerifyRoleJWTFunc(ctx, tok, act, res)
}

// VerifyRoleCert is a mock implementation of Authorizerd.VerifyRoleCert
func (am *AuthorizerdMock) VerifyRoleCert(ctx context.Context, peerCerts []*x509.Certificate, act, res string) error {
	return am.VerifyRoleCertFunc(ctx, peerCerts, act, res)
}

// GetPolicyCache is a mock implementation of Authorizerd.GetPolicyCache
func (am *AuthorizerdMock) GetPolicyCache(ctx context.Context) map[string]interface{} {
	return am.GetPolicyCacheFunc(ctx)
}
