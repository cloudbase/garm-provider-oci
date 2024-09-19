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

	"github.com/cloudbase/garm-provider-common/cloudconfig"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-oci/config"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExtraSpecsFromBootstrapParams(t *testing.T) {
	tests := []struct {
		name           string
		input          params.BootstrapInstance
		expectedOutput *extraSpecs
		errString      string
	}{
		{
			name: "full spec",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{
					"ocpus": 2,
					"memory_in_gbs": 8,
					"boot_volume_size": 256,
					"ssh_public_keys": [
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
					],
					"disable_updates": true,
					"enable_boot_debug": true,
					"extra_packages": ["package1", "package2"],
					"runner_install_template": "IyEvYmluL2Jhc2gKZWNobyBJbnN0YWxsaW5nIHJ1bm5lci4uLg==",
					"pre_install_scripts": {"setup.sh": "IyEvYmluL2Jhc2gKZWNobyBTZXR1cCBzY3JpcHQuLi4="},
					"extra_context": {"key": "value"}
				}`),
			},
			expectedOutput: &extraSpecs{
				Ocpus:          2,
				MemoryInGBs:    8,
				BootVolumeSize: 256,
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
				DisableUpdates:  true,
				EnableBootDebug: true,
				ExtraPackages:   []string{"package1", "package2"},
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					RunnerInstallTemplate: []byte("#!/bin/bash\necho Installing runner..."),
					PreInstallScripts: map[string][]byte{
						"setup.sh": []byte("#!/bin/bash\necho Setup script..."),
					},
					ExtraContext: map[string]string{"key": "value"},
				},
			},
			errString: "",
		},
		{
			name: "specs just with ocpus",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{
					"ocpus": 2
				}`),
			},
			expectedOutput: &extraSpecs{
				Ocpus: 2,
			},
			errString: "",
		},
		{
			name: "specs just with memory in gbs",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{
					"memory_in_gbs": 8
				}`),
			},
			expectedOutput: &extraSpecs{
				MemoryInGBs: 8,
			},
			errString: "",
		},
		{
			name: "specs just with boot volume size",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{
					"boot_volume_size": 256
				}`),
			},
			expectedOutput: &extraSpecs{
				BootVolumeSize: 256,
			},
			errString: "",
		},
		{
			name: "specs just with ssh public keys",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{
					"ssh_public_keys": [
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"
					]
				}`),
			},
			expectedOutput: &extraSpecs{
				SSHPublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
				},
			},
			errString: "",
		},
		{
			name: "specs just with disable_updates",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"disable_updates": true}`),
			},
			expectedOutput: &extraSpecs{
				DisableUpdates: true,
			},
			errString: "",
		},
		{
			name: "specs just with enable_boot_debug",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"enable_boot_debug": true}`),
			},
			expectedOutput: &extraSpecs{
				EnableBootDebug: true,
			},
			errString: "",
		},
		{
			name: "specs just with extra_packages",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_packages": ["package1", "package2"]}`),
			},
			expectedOutput: &extraSpecs{
				ExtraPackages: []string{"package1", "package2"},
			},
			errString: "",
		},
		{
			name: "spec just with RunnerInstallTemplate",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"runner_install_template": "IyEvYmluL2Jhc2gKZWNobyBJbnN0YWxsaW5nIHJ1bm5lci4uLg=="}`),
			},
			expectedOutput: &extraSpecs{
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					RunnerInstallTemplate: []byte("#!/bin/bash\necho Installing runner..."),
				},
			},
			errString: "",
		},
		{
			name: "spec just with PreInstallScripts",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"pre_install_scripts": {"setup.sh": "IyEvYmluL2Jhc2gKZWNobyBTZXR1cCBzY3JpcHQuLi4="}}`),
			},
			expectedOutput: &extraSpecs{
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					PreInstallScripts: map[string][]byte{
						"setup.sh": []byte("#!/bin/bash\necho Setup script..."),
					},
				},
			},
			errString: "",
		},
		{
			name: "spec just with ExtraContext",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_context": {"key": "value"}}`),
			},
			expectedOutput: &extraSpecs{
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					ExtraContext: map[string]string{"key": "value"},
				},
			},
			errString: "",
		},
		{
			name: "missing extra specs",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{}`),
			},
			expectedOutput: &extraSpecs{},
			errString:      "",
		},
		{
			name: "invalid json",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{`),
			},
			expectedOutput: nil,
			errString:      "failed to validate extra specs",
		},
		{
			name: "invalid input for ocpus - wrong data type",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"ocpus": "2"}`),
			},
			expectedOutput: nil,
			errString:      "ocpus: Invalid type. Expected: number, given: string",
		},
		{
			name: "invalid input for memory in gbs - wrong data type",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"memory_in_gbs": "8"}`),
			},
			expectedOutput: nil,
			errString:      "memory_in_gbs: Invalid type. Expected: number, given: string",
		},
		{
			name: "invalid input for boot volume size - wrong data type",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"boot_volume_size": "256"}`),
			},
			expectedOutput: nil,
			errString:      "boot_volume_size: Invalid type. Expected: integer, given: string",
		},
		{
			name: "invalid input for ssh public keys - wrong data type",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"ssh_public_keys": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"}`),
			},
			expectedOutput: nil,
			errString:      "ssh_public_keys: Invalid type. Expected: array, given: string",
		},
		{
			name: "invalid input for extra packages - wrong data type",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_packages": "package1"}`),
			},
			expectedOutput: nil,
			errString:      "extra_packages: Invalid type. Expected: array, given: string",
		},
		{
			name: "invalid input for extra context - wrong data type",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_context": "key: value"}`),
			},
			expectedOutput: nil,
			errString:      "extra_context: Invalid type. Expected: object, given: string",
		},
		{
			name: "invalid input for runner install template - wrong data type",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"runner_install_template": 123}`),
			},
			expectedOutput: nil,
			errString:      "runner_install_template: Invalid type. Expected: string, given: integer",
		},
		{
			name: "invalid input for pre install scripts - wrong data type",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"pre_install_scripts": "setup.sh"}`),
			},
			expectedOutput: nil,
			errString:      "pre_install_scripts: Invalid type. Expected: object, given: string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := newExtraSpecsFromBootstrapData(tt.input)
			if tt.errString == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			}
			require.Equal(t, tt.expectedOutput, output)
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
