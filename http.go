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

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type HTTPClient struct {
	logger               logger.Logger
	address              string
	permissionQueryPath  string
	permissionFilterPath string
	requestTimeout       time.Duration
	verbose              bool
	overrideHeaderValue  string
	httpClient           *http.Client
}

func NewHTTPClient(parentLogger logger.Logger,
	address string,
	permissionQueryPath string,
	permissionFilterPath string,
	requestTimeout time.Duration,
	verbose bool,
	overrideHeaderValue string,
	skipTLSVerify bool,
) *HTTPClient {

	// enrich request timeout with a default value if not set
	if requestTimeout == 0 {
		requestTimeout = DefaultRequestTimeOut
	}

	transport := &http.Transport{}

	// Enable this only for development purposes
	if skipTLSVerify {
		transport.TLSClientConfig = &tls.Config{
			MinVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true,
		}
	}

	newClient := HTTPClient{
		logger:               parentLogger.GetChild("opa"),
		address:              address,
		permissionQueryPath:  permissionQueryPath,
		permissionFilterPath: permissionFilterPath,
		requestTimeout:       requestTimeout,
		verbose:              verbose,
		overrideHeaderValue:  overrideHeaderValue,
		httpClient: &http.Client{
			Timeout:   requestTimeout,
			Transport: transport,
		},
	}
	return &newClient
}

// QueryPermissionsMultiResources query permissions for multiple resources at once.
// The response is a list of booleans indicating for each resource if the action against such resource
// is allowed or not.
// Therefore, it is guaranteed that len(resources) and len(results) are equal and
// resources[i] query permission is at results[i]
func (c *HTTPClient) QueryPermissionsMultiResources(ctx context.Context,
	resources []string,
	action Action,
	permissionOptions *PermissionOptions) ([]bool, error) {

	// initialize results
	results := make([]bool, len(resources))

	// If the override header value matches the configured override header value, allow without checking
	if c.overrideHeaderValue != "" && permissionOptions.OverrideHeaderValue == c.overrideHeaderValue {

		// allow them all
		for i := 0; i < len(results); i++ {
			results[i] = true
		}

		return results, nil
	}

	requestURL := fmt.Sprintf("%s%s", c.address, c.permissionFilterPath)

	// send the request
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   UserAgent,
	}
	request := PermissionFilterRequest{Input: PermissionFilterRequestInput{
		resources,
		string(action),
		permissionOptions.MemberIds,
	}}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate request body")
	}

	if c.verbose {
		c.logger.InfoWithCtx(ctx,
			"Sending request to OPA",
			"requestBody", string(requestBody),
			"requestURL", requestURL)
	}
	var responseBody []byte
	if err := retryUntilSuccessful(6*time.Second,
		1*time.Second,
		func() bool {
			responseBody, _, err = sendHTTPRequest(ctx,
				c.httpClient,
				http.MethodPost,
				requestURL,
				requestBody,
				headers,
				[]*http.Cookie{},
				http.StatusOK)
			if err != nil {
				c.logger.WarnWithCtx(ctx, "Failed to send HTTP request to OPA, retrying",
					"err", err.Error())
				return false
			}
			return true
		}); err != nil {
		if c.verbose {
			c.logger.ErrorWithCtx(ctx,
				"Failed to send HTTP request to OPA",
				"err", errors.GetErrorStackString(err, 10))
		}
		return nil, errors.Wrap(err, "Failed to send HTTP request to OPA")
	}

	if c.verbose {
		c.logger.InfoWithCtx(ctx, "Received response from OPA",
			"responseBody", string(responseBody))
	}

	permissionFilterResponse := PermissionFilterResponse{}
	if err := json.Unmarshal(responseBody, &permissionFilterResponse); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal response body")
	}

	if c.verbose {
		c.logger.InfoWithCtx(ctx, "Successfully unmarshalled permission filter response",
			"permissionFilterResponse", permissionFilterResponse)
	}

	for resourceIdx, resource := range resources {
		if slices.Contains(permissionFilterResponse.Result, resource) {
			results[resourceIdx] = true
		}
	}
	return results, nil
}

func (c *HTTPClient) QueryPermissions(ctx context.Context,
	resource string,
	action Action,
	permissionOptions *PermissionOptions) (bool, error) {

	// If the override header value matches the configured override header value, allow without checking
	if c.overrideHeaderValue != "" && permissionOptions.OverrideHeaderValue == c.overrideHeaderValue {
		return true, nil
	}

	requestURL := fmt.Sprintf("%s%s", c.address, c.permissionQueryPath)

	// send the request
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   UserAgent,
	}
	request := PermissionQueryRequest{Input: PermissionQueryRequestInput{
		resource,
		string(action),
		permissionOptions.MemberIds,
	}}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return false, errors.Wrap(err, "Failed to generate request body")
	}

	if c.verbose {
		c.logger.InfoWithCtx(ctx, "Sending request to OPA",
			"requestBody", string(requestBody),
			"requestURL", requestURL)
	}
	var responseBody []byte
	if err := retryUntilSuccessful(6*time.Second,
		1*time.Second,
		func() bool {
			responseBody, _, err = sendHTTPRequest(ctx,
				c.httpClient,
				http.MethodPost,
				requestURL,
				requestBody,
				headers,
				[]*http.Cookie{},
				http.StatusOK)
			if err != nil {
				c.logger.WarnWithCtx(ctx, "Failed to send HTTP request to OPA, retrying",
					"err", err.Error())
				return false
			}
			return true
		}); err != nil {
		if c.verbose {
			c.logger.ErrorWithCtx(ctx, "Failed to send HTTP request to OPA",
				"err", errors.GetErrorStackString(err, 10))
		}
		return false, errors.Wrap(err, "Failed to send HTTP request to OPA")
	}

	if c.verbose {
		c.logger.InfoWithCtx(ctx, "Received response from OPA",
			"responseBody", string(responseBody))
	}

	permissionResponse := PermissionQueryResponse{}
	if err := json.Unmarshal(responseBody, &permissionResponse); err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal response body")
	}

	if c.verbose {
		c.logger.InfoWithCtx(ctx, "Successfully unmarshalled permission response",
			"permissionResponse", permissionResponse)
	}

	return permissionResponse.Result, nil
}
