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

	"github.com/nuclio/logger"
)

type NopClient struct {
	logger  logger.Logger
	verbose bool
}

func NewNopClient(parentLogger logger.Logger, verbose bool) *NopClient {
	newClient := NopClient{
		logger:  parentLogger.GetChild("opa"),
		verbose: verbose,
	}
	return &newClient
}

func (c *NopClient) QueryPermissionsMultiResources(ctx context.Context,
	resources []string, action Action, permissionOptions *PermissionOptions) ([]bool, error) {
	if c.verbose {
		c.logger.InfoWithCtx(ctx,
			"Skipping permission query for multi resources",
			"resources", resources,
			"action", action,
			"permissionOptions", permissionOptions)
	}
	results := make([]bool, len(resources))
	for i := 0; i < len(results); i++ {
		results[i] = true
	}
	return results, nil
}

func (c *NopClient) QueryPermissions(ctx context.Context, resource string, action Action, permissionOptions *PermissionOptions) (bool, error) {
	if c.verbose {
		c.logger.InfoWith("Skipping permission query",
			"resource", resource,
			"action", action,
			"permissionOptions", permissionOptions)
	}
	return true, nil
}
