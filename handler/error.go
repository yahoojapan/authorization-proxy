package handler

type RFC7807WithAthenz struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Status        int
	InvalidParams []InvalidParam `json:"invalid-params,ommitempty"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance"`
	RoleToken     string         `json:"role_tolen"`
}

type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

const (
	ProblemJSONContentType        = "application/problem+json"
	HttpStatusClientClosedRequest = 499
)
