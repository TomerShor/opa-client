/*
Copyright 2025 The Nuclio Authors.

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

package opaclient

import "time"

type ClientKind string

const (
	ClientKindHTTP ClientKind = "http"
	ClientKindNop  ClientKind = "nop"
	ClientKindMock ClientKind = "mock"

	DefaultClientKind     = ClientKindNop
	DefaultRequestTimeOut = 10 * time.Second
)

type Config struct {

	// OPA server address
	Address string `json:"address,omitempty"`

	// client kind to use (nop | http | mock)
	ClientKind ClientKind `json:"clientKind,omitempty"`

	// timeout period when querying opa server
	RequestTimeout int `json:"requestTimeout,omitempty"`

	// the path used when querying single resource against opa server (e.g.: /v1/data/somewhere/authz/allow)
	PermissionQueryPath string `json:"permissionQueryPath,omitempty"`

	// the path used when querying multiple resources against opa server (e.g.: /v1/data/somewhere/authz/filter_allowed)
	PermissionFilterPath string `json:"permissionFilterPath,omitempty"`

	// for extra verbosity
	Verbose bool `json:"verbose,omitempty"`

	// the header value for bypassing OPA if needed
	OverrideHeaderValue string `json:"overrideHeaderValue,omitempty"`

	// SkipTLSVerify indicates whether to skip TLS verification for the OPA server
	SkipTLSVerify bool `json:"skipTLSVerify,omitempty"`
}

type PermissionOptions struct {
	MemberIds           []string
	RaiseForbidden      bool
	OverrideHeaderValue string
}

type PermissionQueryRequestInput struct {
	Resource string   `json:"resource,omitempty"`
	Action   string   `json:"action,omitempty"`
	Ids      []string `json:"ids,omitempty"`
}

type PermissionQueryRequest struct {
	Input PermissionQueryRequestInput `json:"input,omitempty"`
}

type PermissionFilterRequestInput struct {
	Resources []string `json:"resources,omitempty"`
	Action    string   `json:"action,omitempty"`
	Ids       []string `json:"ids,omitempty"`
}

type PermissionFilterRequest struct {
	Input PermissionFilterRequestInput `json:"input,omitempty"`
}

type PermissionQueryResponse struct {
	Result bool `json:"result,omitempty"`
}

type PermissionFilterResponse struct {
	Result []string `json:"result,omitempty"`
}

type Action string

const (
	ActionRead   Action = "read"
	ActionList   Action = "list"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)
