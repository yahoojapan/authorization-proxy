package service

import (
	"context"
	"crypto/x509"
)

type AuthorizedMock struct {
	StartFunc           func(context.Context) <-chan error
	VerifyRoleTokenFunc func(ctx context.Context, tok, act, res string) error
	VerifyRoleJWTFunc   func(ctx context.Context, tok, act, res string) error
	VerifyRoleCertFunc  func(ctx context.Context, peerCerts []*x509.Certificate, act, res string) error
	GetPolicyCacheFunc  func(ctx context.Context) map[string]interface{}
}

func (am *AuthorizedMock) Start(ctx context.Context) <-chan error {
	return am.StartFunc(ctx)
}

func (am *AuthorizedMock) VerifyRoleToken(ctx context.Context, tok, act, res string) error {
	return am.VerifyRoleTokenFunc(ctx, tok, act, res)
}

func (am *AuthorizedMock) VerifyRoleJWT(ctx context.Context, tok, act, res string) error {
	return am.VerifyRoleJWTFunc(ctx, tok, act, res)
}

func (am *AuthorizedMock) VerifyRoleCert(ctx context.Context, peerCerts []*x509.Certificate, act, res string) error {
	return am.VerifyRoleCertFunc(ctx, peerCerts, act, res)
}

func (am *AuthorizedMock) GetPolicyCache(ctx context.Context) map[string]interface{} {
	return am.GetPolicyCacheFunc(ctx)
}
