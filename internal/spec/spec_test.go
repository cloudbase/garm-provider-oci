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

package spec

import (
	"encoding/json"
	"testing"

	"github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-oci/config"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonSchemaValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     json.RawMessage
		errString string
	}{
		{
			name: "valid",
			input: json.RawMessage(`{
				"ocpus": 2,
				"memory_in_gbs": 8,
				"boot_volume_size": 256,
				"ssh_public_keys": [
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
				]
			}`),
			errString: "",
		},
		{
			name: "invalid",
			input: json.RawMessage(`{
				"ocpus": 2,
				"memory_in_gbs": 8,
				"boot_volume_size": 256,
				"ssh_public_keys": [
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
				],
				"extra": "extra"
			}`),
			errString: "Additional property extra is not allowed",
		},
		{
			name: "invalid ocpus",
			input: json.RawMessage(`{
				"ocpus": "2",
				"memory_in_gbs": 8,
				"boot_volume_size": 256,
				"ssh_public_keys": [
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
				]
			}`),
			errString: "ocpus: Invalid type. Expected: number, given: string",
		},
		{
			name: "invalid memory_in_gbs",
			input: json.RawMessage(`{
				"ocpus": 2,
				"memory_in_gbs": "8",
				"boot_volume_size": 256,
				"ssh_public_keys": [
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
				]
			}`),
			errString: "memory_in_gbs: Invalid type. Expected: number, given: string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jsonSchemaValidation(tt.input)

			if tt.errString == "" {
				assert.NoError(t, err, "Expected no error for valid input")
			} else {
				if assert.Error(t, err, "Expected an error for invalid input") {
					assert.Contains(t, err.Error(), tt.errString, "Error message does not match")
				}
			}
		})
	}
}

func TestNewExtraSpecsFromBootstrapData(t *testing.T) {
	tests := []struct {
		name      string
		input     params.BootstrapInstance
		errString string
	}{
		{
			name: "valid",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{
					"ocpus": 2,
					"memory_in_gbs": 8,
					"boot_volume_size": 256,
					"ssh_public_keys": [
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
					]
				}`),
			},
			errString: "",
		},
		{
			name: "invalid",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{
					"ocpus": 2,
					"memory_in_gbs": 8,
					"boot_volume_size": 256,
					"ssh_public_keys": [
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
					],
					"extra": "extra"
				}`),
			},
			errString: "Additional property extra is not allowed",
		},
		{
			name: "empty",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{}`),
			},
			errString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newExtraSpecsFromBootstrapData(tt.input)

			if tt.errString == "" {
				require.Nil(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestGetRunnerSpecFromBootstrapParams(t *testing.T) {
	Mocktools := params.RunnerApplicationDownload{
		OS:           common.String("linux"),
		Architecture: common.String("amd64"),
		DownloadURL:  common.String("MockURL"),
		Filename:     common.String("garm-runner"),
	}
	DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return Mocktools, nil
	}
	data := params.BootstrapInstance{
		OSType: params.Linux,
		ExtraSpecs: json.RawMessage(`{
			"ocpus": 2,
			"memory_in_gbs": 8,
			"boot_volume_size": 256,
			"ssh_public_keys": [
				"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
			]
		}`),
	}
	config := &config.Config{
		AvailabilityDomain: "MockAvailabilityDomain",
		CompartmentId:      "MockCompartmentId",
		SubnetID:           "MockSubnetID",
		NsgID:              "MockNsgID",
		TenancyID:          "MockTenancyID",
		UserID:             "MockUserID",
		Region:             "MockRegion",
		Fingerprint:        "MockFingerprint",
		PrivateKeyPath:     "MockPrivateKeyPath",
	}
	ExpectedRunnerSpec := &RunnerSpec{
		AvailabilityDomain: "MockAvailabilityDomain",
		CompartmentID:      "MockCompartmentId",
		SubnetID:           "MockSubnetID",
		NsgID:              "MockNsgID",
		BootVolumeSize:     256,
		UserData:           "",
		ControllerID:       "MockControllerID",
		Ocpus:              2,
		MemoryInGBs:        8,
		SSHPublicKeys: []string{
			"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
			"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
		},
		Tools:           Mocktools,
		BootstrapParams: data,
	}

	spec, err := GetRunnerSpecFromBootstrapParams(config, data, "MockControllerID")
	assert.Nil(t, err)
	spec.UserData = ""
	assert.Equal(t, ExpectedRunnerSpec, spec)
}

func TestMergeExtraSpecs(t *testing.T) {
	tests := []struct {
		name     string
		spec     *RunnerSpec
		extra    *extraSpecs
		expected *RunnerSpec
	}{
		{
			name: "empty",
			spec: &RunnerSpec{},
			extra: &extraSpecs{
				Ocpus:          2,
				MemoryInGBs:    8,
				BootVolumeSize: 256,
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
			},
			expected: &RunnerSpec{
				Ocpus:          2,
				MemoryInGBs:    8,
				BootVolumeSize: 256,
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
			},
		},
		{
			name: "non-empty",
			spec: &RunnerSpec{
				Ocpus:          1,
				MemoryInGBs:    4,
				BootVolumeSize: 128,
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
			},
			extra: &extraSpecs{
				Ocpus:          2,
				MemoryInGBs:    8,
				BootVolumeSize: 256,
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
			},
			expected: &RunnerSpec{
				Ocpus:          2,
				MemoryInGBs:    8,
				BootVolumeSize: 256,
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
			},
		},
		{
			name: "empty extra",
			spec: &RunnerSpec{
				Ocpus:          1,
				MemoryInGBs:    4,
				BootVolumeSize: 255,
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
			},
			extra: &extraSpecs{},
			expected: &RunnerSpec{
				Ocpus:          1,
				MemoryInGBs:    4,
				BootVolumeSize: 255,
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.spec.MergeExtraSpecs(tt.extra)
			assert.Equal(t, tt.expected, tt.spec)
		})
	}
}
