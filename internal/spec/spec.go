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
				}
			},
			"additionalProperties": false
		}
	`
)

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
	Ocpus          float32  `json:"ocpus"`
	MemoryInGBs    float32  `json:"memory_in_gbs"`
	BootVolumeSize int64    `json:"boot_volume_size"`
	SSHPublicKeys  []string `json:"ssh_public_keys"`
}

func GetRunnerSpecFromBootstrapParams(cfg *config.Config, data params.BootstrapInstance, controllerID string) (*RunnerSpec, error) {
	tools, err := util.GetTools(data.OSType, data.OSArch, data.Tools)
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
	}

	spec.MergeExtraSpecs(extraSpecs)
	spec.SetUserData()

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
	switch r.BootstrapParams.OSType {
	case params.Linux:
		udata, err := cloudconfig.GetCloudConfig(r.BootstrapParams, r.Tools, r.BootstrapParams.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to generate userdata: %w", err)
		}
		return []byte(udata), nil
	case params.Windows:
		udata, err := cloudconfig.GetCloudConfig(r.BootstrapParams, r.Tools, r.BootstrapParams.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to generate userdata: %w", err)
		}
		return []byte(udata), nil
	}
	return nil, fmt.Errorf("unsupported OS type for cloud config: %s", r.BootstrapParams.OSType)
}
