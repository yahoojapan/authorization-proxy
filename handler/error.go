/*
Copyright (C)  2018 Yahoo Japan Corporation Athenz team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handler

// RFC7807Error represents the error message fulfilling RFC7807 standard.
type RFC7807Error struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Status        int
	InvalidParams []InvalidParam `json:"invalid-params,omitempty"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance"`
	RoleToken     string         `json:"role_token"`
}

// InvalidParam represents the invalid parameters requested by the user.
type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

const (
	// ProblemJSONContentType represents the media type of the error response
	ProblemJSONContentType = "application/problem+json"

	// HTTPStatusClientClosedRequest represents a non-standard status code meaning that the client closed the connection before the server answered the request
	HTTPStatusClientClosedRequest = 499

	// ErrMsgUnverified "unauthenticated/unauthorized"
	ErrMsgUnverified = "unauthenticated/unauthorized"

	// ErrGRPCMetadataNotFound "grpc metadata not found"
	ErrGRPCMetadataNotFound = "grpc metadata not found"

	// ErrRoleTokenNotFound "role token not found"
	ErrRoleTokenNotFound = "role token not found"
)
