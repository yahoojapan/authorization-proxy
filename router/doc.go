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

/*
Package service manages the main logic of Authorization Proxy.
It contains a token updater to periodically update the N-token for communicating with Athenz,
and policy updater to periodically update Athenz policy,
and athenz config updater to periodically updater Athenz Data.
*/

// Package router provides the router and API routes implementations.
package router
