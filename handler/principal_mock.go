package handler

import (
	"github.com/yahoojapan/athenz-authorizer/v4/access"
	"github.com/yahoojapan/athenz-authorizer/v4/role"
)

// RoleTokenMock is a mock of Principal
type RoleTokenMock struct {
	t role.Token
}

// Name is a mock implementation of Principal
func (rt *RoleTokenMock) Name() string {
	return rt.t.Principal
}

// Roles is a mock implementation of Principal
func (rt *RoleTokenMock) Roles() []string {
	return rt.t.Roles
}

// Domain is a mock implementation of Principal
func (rt *RoleTokenMock) Domain() string {
	return rt.t.Domain
}

// IssueTime is a mock implementation of Principal
func (rt *RoleTokenMock) IssueTime() int64 {
	return rt.t.TimeStamp.Unix()
}

// ExpiryTime is a mock implementation of Principal
func (rt *RoleTokenMock) ExpiryTime() int64 {
	return rt.t.ExpiryTime.Unix()
}

// OAuth2AccessTokenMock is a mock of Principal
type OAuth2AccessTokenMock struct {
	t access.OAuth2AccessTokenClaim
}

// Name is a mock implementation of Principal
func (oat *OAuth2AccessTokenMock) Name() string {
	return oat.t.Subject
}

// Roles is a mock implementation of Principal
func (oat *OAuth2AccessTokenMock) Roles() []string {
	return oat.t.Scope
}

// Domain is a mock implementation of Principal
func (oat *OAuth2AccessTokenMock) Domain() string {
	return oat.t.Audience
}

// IssueTime is a mock implementation of Principal
func (oat *OAuth2AccessTokenMock) IssueTime() int64 {
	return oat.t.IssuedAt
}

// ExpiryTime is a mock implementation of Principal
func (oat *OAuth2AccessTokenMock) ExpiryTime() int64 {
	return oat.t.ExpiresAt
}

// ClientID is a mock implementation of OAuthAccessToken
func (oat *OAuth2AccessTokenMock) ClientID() string {
	return oat.t.ClientID
}
