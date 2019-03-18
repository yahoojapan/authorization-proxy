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
