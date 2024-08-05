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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/cloudbase/garm-provider-common/cloudconfig"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-common/util"
	"github.com/cloudbase/garm-provider-oci/config"
	"github.com/xeipuuv/gojsonschema"
)

const (
	defaultMemoryAllocation float32 = 4
	defaultOcpusAllocation  float32 = 1
	defaultBootVolumeSize   int64   = 255
	jsonSchema              string  = `
		{
			"$schema": "http://cloudbase.it/garm-provider-oci/schemas/extra_specs#",
			"type": "object",
			"properties": {
				"ocpus": {
					"type": "number",
					"description": "Number of OCPUs"
				},
				"memory_in_gbs": {
					"type": "number",
					"description": "Memory in GBs"
				},
				"boot_volume_size": {
					"type": "number",
					"description": "Boot volume size in GB"
				},
				"ssh_public_keys": {
					"type": "array",
					"description": "List of SSH public keys",
					"items": {
						"type": "string",
						"description": "A SSH public key"
					}
				},
				"disable_updates": {
					"type": "boolean",
					"description": "Disable automatic updates on the VM."
				},
				"enable_boot_debug": {
					"type": "boolean",
					"description": "Enable boot debug on the VM."
				},
				"extra_packages": {
					"type": "array",
					"description": "Extra packages to install on the VM.",
					"items": {
						"type": "string"
					}
				},
				"runner_install_template": {
					"type": "string",
					"description": "This option can be used to override the default runner install template. If used, the caller is responsible for the correctness of the template as well as the suitability of the template for the target OS. Use the extra_context extra spec if your template has variables in it that need to be expanded."
				},
				"extra_context": {
					"type": "object",
					"description": "Extra context that will be passed to the runner_install_template.",
					"additionalProperties": {
						"type": "string"
					}
				},
				"pre_install_scripts": {
					"type": "object",
					"description": "A map of pre-install scripts that will be run before the runner install script. These will run as root and can be used to prep a generic image before we attempt to install the runner. The key of the map is the name of the script as it will be written to disk. The value is a byte array with the contents of the script.",
					"additionalProperties": {
						"type": "string"
					}
				}
			},
			"additionalProperties": false
		}
	`
)

type ToolFetchFunc func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error)

var DefaultToolFetch ToolFetchFunc = util.GetTools

func jsonSchemaValidation(schema json.RawMessage) error {
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	extraSpecsLoader := gojsonschema.NewBytesLoader(schema)
	result, err := gojsonschema.Validate(schemaLoader, extraSpecsLoader)
	if err != nil {
		return fmt.Errorf("failed to validate schema: %w", err)
	}
	if !result.Valid() {
		return fmt.Errorf("schema validation failed: %s", result.Errors())
	}
	return nil
}

func newExtraSpecsFromBootstrapData(data params.BootstrapInstance) (*extraSpecs, error) {
	spec := &extraSpecs{}

	if err := jsonSchemaValidation(data.ExtraSpecs); err != nil {
		return nil, fmt.Errorf("failed to validate extra specs: %w", err)
	}

	if len(data.ExtraSpecs) > 0 {
		if err := json.Unmarshal(data.ExtraSpecs, spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal extra specs: %w", err)
		}
	}

	return spec, nil
}

type extraSpecs struct {
	Ocpus           float32  `json:"ocpus"`
	MemoryInGBs     float32  `json:"memory_in_gbs"`
	BootVolumeSize  int64    `json:"boot_volume_size"`
	SSHPublicKeys   []string `json:"ssh_public_keys"`
	DisableUpdates  *bool    `json:"disable_updates"`
	EnableBootDebug *bool    `json:"enable_boot_debug"`
	ExtraPackages   []string `json:"extra_packages"`
}

func GetRunnerSpecFromBootstrapParams(cfg *config.Config, data params.BootstrapInstance, controllerID string) (*RunnerSpec, error) {
	tools, err := DefaultToolFetch(data.OSType, data.OSArch, data.Tools)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %s", err)
	}

	extraSpecs, err := newExtraSpecsFromBootstrapData(data)
	if err != nil {
		return nil, fmt.Errorf("error loading extra specs: %w", err)
	}

	spec := &RunnerSpec{
		AvailabilityDomain: cfg.AvailabilityDomain,
		CompartmentID:      cfg.CompartmentId,
		SubnetID:           cfg.SubnetID,
		NsgID:              cfg.NsgID,
		ControllerID:       controllerID,
		Tools:              tools,
		BootstrapParams:    data,
		ExtraPackages:      extraSpecs.ExtraPackages,
	}

	spec.MergeExtraSpecs(extraSpecs)
	if err := spec.SetUserData(); err != nil {
		return nil, fmt.Errorf("error setting extra specs: %w", err)
	}

	return spec, nil
}

type RunnerSpec struct {
	AvailabilityDomain string
	CompartmentID      string
	SubnetID           string
	NsgID              string
	BootVolumeSize     int64
	UserData           string
	ControllerID       string
	Ocpus              float32
	MemoryInGBs        float32
	SSHPublicKeys      []string
	DisableUpdates     bool
	ExtraPackages      []string
	EnableBootDebug    bool
	Tools              params.RunnerApplicationDownload
	BootstrapParams    params.BootstrapInstance
	mux                sync.Mutex
}

func (r *RunnerSpec) MergeExtraSpecs(extraSpecs *extraSpecs) {
	r.Ocpus = defaultOcpusAllocation
	if extraSpecs.Ocpus > 0 {
		r.Ocpus = extraSpecs.Ocpus
	}
	r.MemoryInGBs = defaultMemoryAllocation
	if extraSpecs.MemoryInGBs > 0 {
		r.MemoryInGBs = extraSpecs.MemoryInGBs
	}
	r.BootVolumeSize = defaultBootVolumeSize
	if extraSpecs.BootVolumeSize > 0 {
		r.BootVolumeSize = extraSpecs.BootVolumeSize
	}
	if len(extraSpecs.SSHPublicKeys) > 0 {
		r.SSHPublicKeys = extraSpecs.SSHPublicKeys
	}
	if extraSpecs.DisableUpdates != nil {
		r.DisableUpdates = *extraSpecs.DisableUpdates
	}
	if extraSpecs.EnableBootDebug != nil {
		r.EnableBootDebug = *extraSpecs.EnableBootDebug
	}
}

func (r *RunnerSpec) SetUserData() error {
	r.mux.Lock()
	defer r.mux.Unlock()
	customData, err := r.ComposeUserData()
	if err != nil {
		return fmt.Errorf("failed to compose userdata: %w", err)
	}

	if len(customData) == 0 {
		return fmt.Errorf("failed to generate custom data")
	}

	asBase64 := base64.StdEncoding.EncodeToString(customData)
	r.UserData = asBase64
	return nil
}

func (r *RunnerSpec) ComposeUserData() ([]byte, error) {
	bootstrapParams := r.BootstrapParams
	bootstrapParams.UserDataOptions.DisableUpdatesOnBoot = r.DisableUpdates
	bootstrapParams.UserDataOptions.ExtraPackages = r.ExtraPackages
	bootstrapParams.UserDataOptions.EnableBootDebug = r.EnableBootDebug
	switch r.BootstrapParams.OSType {
	case params.Linux, params.Windows:
		udata, err := cloudconfig.GetCloudConfig(bootstrapParams, r.Tools, bootstrapParams.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to generate userdata: %w", err)
		}
		return []byte(udata), nil
	}
	return nil, fmt.Errorf("unsupported OS type for cloud config: %s", bootstrapParams.OSType)
}
