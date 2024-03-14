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

package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-oci/config"
	"github.com/cloudbase/garm-provider-oci/internal/client"
	"github.com/cloudbase/garm-provider-oci/internal/spec"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateInstance(t *testing.T) {
	ctx := context.Background()
	mockComputeClient := new(client.MockComputeClient)
	spec.DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return params.RunnerApplicationDownload{
			OS:           common.String("linux"),
			Architecture: common.String("amd64"),
			DownloadURL:  common.String("MockURL"),
			Filename:     common.String("garm-runner"),
		}, nil
	}
	cfg := &config.Config{
		AvailabilityDomain: "ad",
		CompartmentId:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		TenancyID:          "tenancy",
		UserID:             "user",
		Region:             "region",
		Fingerprint:        "fingerprint",
		PrivateKeyPath:     "private_key_path",
	}
	bootstrapParams := params.BootstrapInstance{
		Name:   "garm-instance",
		Flavor: "n1-standard-1",
		Image:  "projects/garm-testing/global/images/garm-image",
		Tools: []params.RunnerApplicationDownload{
			{
				OS:           common.String("linux"),
				Architecture: common.String("amd64"),
				DownloadURL:  common.String("MockURL"),
				Filename:     common.String("garm-runner"),
			},
		},
		OSType:     params.Linux,
		OSArch:     params.Amd64,
		PoolID:     "my-pool",
		ExtraSpecs: json.RawMessage(`{}`),
	}
	expectedInstance := params.ProviderInstance{
		ProviderID: "garm-instance",
		Name:       "garm-instance",
		OSType:     "linux",
		OSArch:     "amd64",
		Status:     "running",
	}

	OciProvider := OciProvider{
		ociCli:       &client.OciCli{},
		controllerID: "controller",
	}
	OciProvider.ociCli.SetComputeClient(mockComputeClient)
	OciProvider.ociCli.SetConfig(cfg)

	mockComputeClient.On("LaunchInstance", ctx, mock.Anything).Return(core.LaunchInstanceResponse{
		Instance: core.Instance{
			Id:                 common.String("garm-instance"),
			AvailabilityDomain: common.String(cfg.AvailabilityDomain),
			CompartmentId:      common.String(cfg.CompartmentId),
			DisplayName:        common.String(bootstrapParams.Name),
			Shape:              common.String(bootstrapParams.Flavor),
			LifecycleState:     core.InstanceLifecycleStateRunning,
			FreeformTags: map[string]string{
				"Name":               bootstrapParams.Name,
				"GARM_POOL_ID":       bootstrapParams.PoolID,
				"OSType":             string(bootstrapParams.OSType),
				"OSArch":             string(bootstrapParams.OSArch),
				"GARM_CONTROLLER_ID": OciProvider.controllerID,
			},
		},
	}, nil)

	result, err := OciProvider.CreateInstance(ctx, bootstrapParams)
	assert.NoError(t, err)
	assert.Equal(t, expectedInstance, result)
}

func TestGetInstancewithName(t *testing.T) {
	ctx := context.Background()
	mockComputeClient := new(client.MockComputeClient)
	cfg := &config.Config{
		AvailabilityDomain: "ad",
		CompartmentId:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		TenancyID:          "tenancy",
		UserID:             "user",
		Region:             "region",
		Fingerprint:        "fingerprint",
		PrivateKeyPath:     "private_key_path",
	}
	OciProvider := OciProvider{
		ociCli:       &client.OciCli{},
		controllerID: "controller",
	}
	OciProvider.ociCli.SetComputeClient(mockComputeClient)
	OciProvider.ociCli.SetConfig(cfg)
	inst := "garm-instance"
	expectedInstance := params.ProviderInstance{
		ProviderID: "ocid1.instance.oc1.iad.aaaaaaaamf7",
		Name:       "garm-instance",
		OSType:     "linux",
		OSArch:     "amd64",
		Status:     "running",
	}
	mockComputeClient.On("ListInstances", ctx, core.ListInstancesRequest{
		CompartmentId: &cfg.CompartmentId,
	}).Return(core.ListInstancesResponse{
		Items: []core.Instance{{
			Id:                 common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
			AvailabilityDomain: &cfg.AvailabilityDomain,
			CompartmentId:      &cfg.CompartmentId,
			Region:             &cfg.Region,
			FreeformTags:       map[string]string{"Name": inst},
			LifecycleState:     core.InstanceLifecycleStateRunning,
		}}}, nil)
	mockComputeClient.On("GetInstance", ctx, core.GetInstanceRequest{
		InstanceId: common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
	}).Return(core.GetInstanceResponse{
		Instance: core.Instance{
			Id: common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
			FreeformTags: map[string]string{
				"Name":               "garm-instance",
				"GARM_POOL_ID":       "my-pool",
				"OSType":             "linux",
				"OSArch":             "amd64",
				"GARM_CONTROLLER_ID": "controller",
			},
			LifecycleState: core.InstanceLifecycleStateRunning,
		},
	}, nil)

	result, err := OciProvider.GetInstance(ctx, inst)
	assert.NoError(t, err)
	assert.Equal(t, expectedInstance, result)

}

func TestGetInstancewithId(t *testing.T) {
	ctx := context.Background()
	mockComputeClient := new(client.MockComputeClient)
	cfg := &config.Config{
		AvailabilityDomain: "ad",
		CompartmentId:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		TenancyID:          "tenancy",
		UserID:             "user",
		Region:             "region",
		Fingerprint:        "fingerprint",
		PrivateKeyPath:     "private_key_path",
	}
	OciProvider := OciProvider{
		ociCli:       &client.OciCli{},
		controllerID: "controller",
	}
	OciProvider.ociCli.SetComputeClient(mockComputeClient)
	OciProvider.ociCli.SetConfig(cfg)
	inst := "ocid1.instance.oc1.iad.aaaaaaaamf7"
	expectedInstance := params.ProviderInstance{
		ProviderID: inst,
		Name:       "garm-instance",
		OSType:     "linux",
		OSArch:     "amd64",
		Status:     "running",
	}
	mockComputeClient.On("GetInstance", ctx, core.GetInstanceRequest{
		InstanceId: &inst,
	}).Return(core.GetInstanceResponse{
		Instance: core.Instance{
			Id: &inst,
			FreeformTags: map[string]string{
				"Name":               "garm-instance",
				"GARM_POOL_ID":       "my-pool",
				"OSType":             "linux",
				"OSArch":             "amd64",
				"GARM_CONTROLLER_ID": "controller",
			},
			LifecycleState: core.InstanceLifecycleStateRunning,
		},
	}, nil)

	result, err := OciProvider.GetInstance(ctx, inst)
	assert.NoError(t, err)
	assert.Equal(t, expectedInstance, result)

}

func TestDeleteInstanceWithName(t *testing.T) {
	ctx := context.Background()
	mockComputeClient := new(client.MockComputeClient)
	cfg := &config.Config{
		AvailabilityDomain: "ad",
		CompartmentId:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		TenancyID:          "tenancy",
		UserID:             "user",
		Region:             "region",
		Fingerprint:        "fingerprint",
		PrivateKeyPath:     "private_key_path",
	}
	OciProvider := OciProvider{
		ociCli:       &client.OciCli{},
		controllerID: "controller",
	}
	OciProvider.ociCli.SetComputeClient(mockComputeClient)
	OciProvider.ociCli.SetConfig(cfg)
	inst := "garm-instance"
	mockComputeClient.On("ListInstances", ctx, core.ListInstancesRequest{
		CompartmentId: &cfg.CompartmentId,
	}).Return(core.ListInstancesResponse{
		Items: []core.Instance{{
			Id:                 common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
			AvailabilityDomain: &cfg.AvailabilityDomain,
			CompartmentId:      &cfg.CompartmentId,
			Region:             &cfg.Region,
			FreeformTags:       map[string]string{"Name": inst},
			LifecycleState:     core.InstanceLifecycleStateRunning,
		}}}, nil)
	mockComputeClient.On("TerminateInstance", ctx, core.TerminateInstanceRequest{
		InstanceId: common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
	}).Return(core.TerminateInstanceResponse{}, nil)

	err := OciProvider.DeleteInstance(ctx, inst)
	assert.Nil(t, err)
}

func TestDeleteInstanceWithId(t *testing.T) {
	ctx := context.Background()
	mockComputeClient := new(client.MockComputeClient)
	cfg := &config.Config{
		AvailabilityDomain: "ad",
		CompartmentId:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		TenancyID:          "tenancy",
		UserID:             "user",
		Region:             "region",
		Fingerprint:        "fingerprint",
		PrivateKeyPath:     "private_key_path",
	}
	OciProvider := OciProvider{
		ociCli:       &client.OciCli{},
		controllerID: "controller",
	}
	OciProvider.ociCli.SetComputeClient(mockComputeClient)
	OciProvider.ociCli.SetConfig(cfg)
	inst := "ocid1.instance.oc1.iad.aaaaaaaamf7"
	mockComputeClient.On("TerminateInstance", ctx, core.TerminateInstanceRequest{
		InstanceId: &inst,
	}).Return(core.TerminateInstanceResponse{}, nil)

	err := OciProvider.DeleteInstance(ctx, inst)
	assert.Nil(t, err)
}

func TestListInstance(t *testing.T) {
	ctx := context.Background()
	mockComputeClient := new(client.MockComputeClient)
	cfg := &config.Config{
		AvailabilityDomain: "ad",
		CompartmentId:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		TenancyID:          "tenancy",
		UserID:             "user",
		Region:             "region",
		Fingerprint:        "fingerprint",
		PrivateKeyPath:     "private_key_path",
	}
	OciProvider := OciProvider{
		ociCli:       &client.OciCli{},
		controllerID: "controller",
	}
	OciProvider.ociCli.SetComputeClient(mockComputeClient)
	OciProvider.ociCli.SetConfig(cfg)
	expectedInstance := []params.ProviderInstance{
		{
			ProviderID: "ocid1.instance.oc1.iad.aaaaaaaamf7",
			Name:       "garm-instance",
			OSType:     "linux",
			OSArch:     "amd64",
			Status:     "running",
		},
		{
			ProviderID: "ocid1.instance.oc1.iad.aaaaaaaamf8",
			Name:       "garm-instance2",
			OSType:     "linux",
			OSArch:     "amd64",
			Status:     "running",
		},
	}
	poolID := "my-pool"

	mockComputeClient.On("ListInstances", ctx, core.ListInstancesRequest{
		CompartmentId: &cfg.CompartmentId,
	}).Return(core.ListInstancesResponse{
		Items: []core.Instance{
			{
				Id: common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
				FreeformTags: map[string]string{
					"Name":               "garm-instance",
					"GARM_POOL_ID":       "my-pool",
					"OSType":             "linux",
					"OSArch":             "amd64",
					"GARM_CONTROLLER_ID": "controller",
				},
				LifecycleState: core.InstanceLifecycleStateRunning,
			},
			{
				Id: common.String("ocid1.instance.oc1.iad.aaaaaaaamf8"),
				FreeformTags: map[string]string{
					"Name":               "garm-instance2",
					"GARM_POOL_ID":       "my-pool",
					"OSType":             "linux",
					"OSArch":             "amd64",
					"GARM_CONTROLLER_ID": "controller",
				},
				LifecycleState: core.InstanceLifecycleStateRunning,
			},
		},
	}, nil)

	result, err := OciProvider.ListInstances(ctx, poolID)
	assert.NoError(t, err)
	assert.Equal(t, expectedInstance, result)
}

func TestStop(t *testing.T) {
	ctx := context.Background()
	mockComputeClient := new(client.MockComputeClient)
	cfg := &config.Config{
		AvailabilityDomain: "ad",
		CompartmentId:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		TenancyID:          "tenancy",
		UserID:             "user",
		Region:             "region",
		Fingerprint:        "fingerprint",
		PrivateKeyPath:     "private_key_path",
	}
	OciProvider := OciProvider{
		ociCli:       &client.OciCli{},
		controllerID: "controller",
	}
	OciProvider.ociCli.SetComputeClient(mockComputeClient)
	OciProvider.ociCli.SetConfig(cfg)
	inst := "ocid1.instance.oc1.iad.aaaaaaaamf7"
	mockComputeClient.On("InstanceAction", ctx, core.InstanceActionRequest{
		InstanceId: &inst,
		Action:     core.InstanceActionActionStop,
	}).Return(core.InstanceActionResponse{}, nil)

	err := OciProvider.Stop(ctx, inst, false)
	assert.Nil(t, err)
}

func TestStart(t *testing.T) {
	ctx := context.Background()
	mockComputeClient := new(client.MockComputeClient)
	cfg := &config.Config{
		AvailabilityDomain: "ad",
		CompartmentId:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		TenancyID:          "tenancy",
		UserID:             "user",
		Region:             "region",
		Fingerprint:        "fingerprint",
		PrivateKeyPath:     "private_key_path",
	}
	OciProvider := OciProvider{
		ociCli:       &client.OciCli{},
		controllerID: "controller",
	}
	OciProvider.ociCli.SetComputeClient(mockComputeClient)
	OciProvider.ociCli.SetConfig(cfg)
	inst := "ocid1.instance.oc1.iad.aaaaaaaamf7"
	mockComputeClient.On("InstanceAction", ctx, core.InstanceActionRequest{
		InstanceId: &inst,
		Action:     core.InstanceActionActionStart,
	}).Return(core.InstanceActionResponse{}, nil)

	err := OciProvider.Start(ctx, inst)
	assert.Nil(t, err)
}
