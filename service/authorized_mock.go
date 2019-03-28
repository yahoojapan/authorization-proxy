package service

import "context"

type AuthorizedMock struct {
	StartProviderdFunc  func(context.Context) <-chan error
	VerifyRoleTokenFunc func(ctx context.Context, tok, act, res string) error
}

func (am *AuthorizedMock) StartProviderd(ctx context.Context) <-chan error {
	return am.StartProviderdFunc(ctx)
}

func (am *AuthorizedMock) VerifyRoleToken(ctx context.Context, tok, act, res string) error {
	return am.VerifyRoleTokenFunc(ctx, tok, act, res)
}
