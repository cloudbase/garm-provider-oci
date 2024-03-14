// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Cloudbase Solutions SRL
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package util

import (
	"testing"

	"github.com/cloudbase/garm-provider-common/params"
	"github.com/oracle/oci-go-sdk/v49/core"
	"github.com/stretchr/testify/assert"
)

func TestOciInstanceToProviderInstance(t *testing.T) {
	id := "id"
	tests := []struct {
		name        string
		ociInstance core.Instance
		expected    params.ProviderInstance
		errString   error
	}{
		{
			name: "running instance",
			ociInstance: core.Instance{
				Id: &id,
				FreeformTags: map[string]string{
					"Name":   "name",
					"OSType": "linux",
					"OSArch": "amd64",
				},
				LifecycleState: core.InstanceLifecycleStateRunning,
			},
			expected: params.ProviderInstance{
				ProviderID: "id",
				Name:       "name",
				OSType:     params.Linux,
				OSArch:     params.Amd64,
				Status:     params.InstanceRunning,
			},
		},
		{
			name: "stopped instance",
			ociInstance: core.Instance{
				Id: &id,
				FreeformTags: map[string]string{
					"Name":   "name",
					"OSType": "linux",
					"OSArch": "amd64",
				},
				LifecycleState: core.InstanceLifecycleStateStopped,
			},
			expected: params.ProviderInstance{
				ProviderID: "id",
				Name:       "name",
				OSType:     params.Linux,
				OSArch:     params.Amd64,
				Status:     params.InstanceStopped,
			},
		},
		{
			name: "provisioning instance",
			ociInstance: core.Instance{
				Id: &id,
				FreeformTags: map[string]string{
					"Name":   "name",
					"OSType": "linux",
					"OSArch": "amd64",
				},
				LifecycleState: core.InstanceLifecycleStateProvisioning,
			},
			expected: params.ProviderInstance{
				ProviderID: "id",
				Name:       "name",
				OSType:     params.Linux,
				OSArch:     params.Amd64,
				Status:     params.InstanceStatusUnknown,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := OciInstanceToProviderInstance(tt.ociInstance)
			assert.Equal(t, tt.expected, actual)
		})
	}

}
