package handler

// PrincipalMock is a mock of Principal
type PrincipalMock struct {
	NameFunc       func() string
	RolesFunc      func() []string
	DomainFunc     func() string
	IssueTimeFunc  func() int64
	ExpiryTimeFunc func() int64
}

// OAuthAccessTokenMock is a mock of OAuthAccessToken
type OAuthAccessTokenMock struct {
	PrincipalMock
	ClientIDFunc func() string
}

// Name is a mock implementation of Principal
func (p *PrincipalMock) Name() string {
	return p.NameFunc()
}

// Roles is a mock implementation of Principal
func (p *PrincipalMock) Roles() []string {
	return p.RolesFunc()
}

// Domain is a mock implementation of Principal
func (p *PrincipalMock) Domain() string {
	return p.DomainFunc()
}

// IssueTime is a mock implementation of Principal
func (p *PrincipalMock) IssueTime() int64 {
	return p.IssueTimeFunc()
}

// ExpiryTime is a mock implementation of Principal
func (p *PrincipalMock) ExpiryTime() int64 {
	return p.ExpiryTimeFunc()
}

// AuthorizedRoles is a mock implementation of Principal
func (p *PrincipalMock) AuthorizedRoles() []string {
	return p.AuthorizedRoles()
}

// ClientID is a mock implementation of OAuthAccessToken
func (oat *OAuthAccessTokenMock) ClientID() string {
	return oat.ClientIDFunc()
}
