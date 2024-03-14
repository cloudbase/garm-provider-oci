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

package client

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-oci/config"
	"github.com/cloudbase/garm-provider-oci/internal/spec"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/core"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstance(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	spec := spec.RunnerSpec{
		AvailabilityDomain: "ad",
		CompartmentID:      "compartment",
		SubnetID:           "subnet",
		NsgID:              "nsg",
		BootVolumeSize:     256,
		UserData:           "userdata",
		ControllerID:       "controller",
		Ocpus:              2,
		MemoryInGBs:        8,
		SSHPublicKeys:      []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC"},
		Tools: params.RunnerApplicationDownload{
			OS:           common.String("linux"),
			Architecture: common.String("amd64"),
			DownloadURL:  common.String("MockURL"),
			Filename:     common.String("garm-runner"),
		},
		BootstrapParams: params.BootstrapInstance{
			Name:   "garm-instance",
			Flavor: "VM.Standard.E4.Flex",
			Image:  "ocid1.image.oc1.iad.aaaaaaaamf7",
			OSType: params.Linux,
			OSArch: "amd64",
		},
	}

	expectedInstance := core.Instance{
		AvailabilityDomain: &spec.AvailabilityDomain,
		CompartmentId:      &spec.CompartmentID,
		DisplayName:        &spec.BootstrapParams.Name,
		Shape:              &spec.BootstrapParams.Flavor,
		FreeformTags: map[string]string{
			"Name":               spec.BootstrapParams.Name,
			"GARM_POOL_ID":       spec.BootstrapParams.PoolID,
			"OSType":             string(spec.BootstrapParams.OSType),
			"OSArch":             string(spec.BootstrapParams.OSArch),
			"GARM_CONTROLLER_ID": spec.ControllerID,
		},
		Metadata: map[string]string{
			"user_data":           spec.UserData,
			"ssh_authorized_keys": strings.Join(spec.SSHPublicKeys, "\n"),
		},
		SourceDetails: core.InstanceSourceViaImageDetails{
			ImageId:             &spec.BootstrapParams.Image,
			BootVolumeSizeInGBs: &spec.BootVolumeSize,
		},
	}

	mockComputeClient.On("LaunchInstance", ctx, core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			CompartmentId:      &spec.CompartmentID,
			AvailabilityDomain: &spec.AvailabilityDomain,
			DisplayName:        &spec.BootstrapParams.Name,
			Shape:              &spec.BootstrapParams.Flavor,
			CreateVnicDetails: &core.CreateVnicDetails{
				SubnetId: &spec.SubnetID,
				NsgIds:   []string{spec.NsgID},
			},
			ShapeConfig: &core.LaunchInstanceShapeConfigDetails{
				Ocpus:       common.Float32(spec.Ocpus),
				MemoryInGBs: common.Float32(spec.MemoryInGBs),
			},
			FreeformTags: map[string]string{
				"Name":               spec.BootstrapParams.Name,
				"GARM_POOL_ID":       spec.BootstrapParams.PoolID,
				"OSType":             string(spec.BootstrapParams.OSType),
				"OSArch":             string(spec.BootstrapParams.OSArch),
				"GARM_CONTROLLER_ID": spec.ControllerID,
			},
			Metadata: map[string]string{
				"user_data":           spec.UserData,
				"ssh_authorized_keys": strings.Join(spec.SSHPublicKeys, "\n"),
			},
			SourceDetails: core.InstanceSourceViaImageDetails{
				ImageId:             &spec.BootstrapParams.Image,
				BootVolumeSizeInGBs: &spec.BootVolumeSize,
			},
		},
	}).Return(core.LaunchInstanceResponse{
		Instance: expectedInstance,
	}, nil)

	instance, err := ociCli.CreateInstance(ctx, &spec)

	assert.Nil(t, err)
	assert.Equal(t, expectedInstance, instance)

}

func TestGetInstanceWithName(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	inst := "instance"
	expectedInstance := core.Instance{
		Id:                 common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
		FreeformTags:       map[string]string{"Name": inst},
		AvailabilityDomain: &cfg.AvailabilityDomain,
		CompartmentId:      &cfg.CompartmentId,
		Region:             &cfg.Region,
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
		Instance: expectedInstance,
	}, nil)

	instance, err := ociCli.GetInstance(ctx, inst)

	assert.Nil(t, err)
	assert.Equal(t, expectedInstance, instance)
}

func TestGetInstanceWithId(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	inst := "ocid1.instance.oc1.iad.aaaaaaaamf7"
	expectedInstance := core.Instance{
		Id:                 &inst,
		DisplayName:        &inst,
		AvailabilityDomain: &cfg.AvailabilityDomain,
		CompartmentId:      &cfg.CompartmentId,
		Region:             &cfg.Region,
	}
	mockComputeClient.On("GetInstance", ctx, core.GetInstanceRequest{
		InstanceId: &inst,
	}).Return(core.GetInstanceResponse{
		Instance: expectedInstance,
	}, nil)

	instance, err := ociCli.GetInstance(ctx, inst)

	assert.Nil(t, err)
	assert.Equal(t, expectedInstance, instance)
}

func TestDeleteInstanceWithName(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	inst := "instance"
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

	err := ociCli.DeleteInstance(ctx, inst)

	assert.Nil(t, err)
}

func TestDeleteInstanceWithId(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	inst := "ocid1.instance.oc1.iad.aaaaaaaamf7"
	mockComputeClient.On("TerminateInstance", ctx, core.TerminateInstanceRequest{
		InstanceId: &inst,
	}).Return(core.TerminateInstanceResponse{}, nil)

	err := ociCli.DeleteInstance(ctx, inst)

	assert.Nil(t, err)
}

func TestListInstances(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	expectedInstances := []core.Instance{
		{
			Id:                 common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
			AvailabilityDomain: &cfg.AvailabilityDomain,
			CompartmentId:      &cfg.CompartmentId,
			Region:             &cfg.Region,
			FreeformTags:       map[string]string{"Name": "instance1"},
			LifecycleState:     core.InstanceLifecycleStateRunning,
		},
		{
			Id:                 common.String("ocid1.instance.oc1.iad.aaaaaaaamf8"),
			AvailabilityDomain: &cfg.AvailabilityDomain,
			CompartmentId:      &cfg.CompartmentId,
			Region:             &cfg.Region,
			FreeformTags:       map[string]string{"Name": "instance2"},
			LifecycleState:     core.InstanceLifecycleStateRunning,
		},
	}
	mockComputeClient.On("ListInstances", ctx, core.ListInstancesRequest{
		CompartmentId: &cfg.CompartmentId,
	}).Return(core.ListInstancesResponse{
		Items: expectedInstances,
	}, nil)

	instances, err := ociCli.ListInstances(ctx, "")

	assert.Nil(t, err)
	assert.Equal(t, expectedInstances, instances)
}

func TestStopInstance(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	inst := "ocid1.instance.oc1.iad.aaaaaaaamf7"
	mockComputeClient.On("InstanceAction", ctx, core.InstanceActionRequest{
		InstanceId: &inst,
		Action:     core.InstanceActionActionStop,
	}).Return(core.InstanceActionResponse{}, nil)

	err := ociCli.StopInstance(ctx, inst)

	assert.Nil(t, err)
}

func TestStartInstance(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	inst := "ocid1.instance.oc1.iad.aaaaaaaamf7"
	mockComputeClient.On("InstanceAction", ctx, core.InstanceActionRequest{
		InstanceId: &inst,
		Action:     core.InstanceActionActionStart,
	}).Return(core.InstanceActionResponse{}, nil)

	err := ociCli.StartInstance(ctx, inst)

	assert.Nil(t, err)
}

func TestFindInstanceByTags(t *testing.T) {
	ctx := context.Background()
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
	mockComputeClient := new(MockComputeClient)
	ociCli := &OciCli{
		computeClient: mockComputeClient,
		cfg:           cfg,
	}
	tags := map[string]string{
		"Name": "instance1",
	}
	expectedInstance := core.Instance{
		Id:                 common.String("ocid1.instance.oc1.iad.aaaaaaaamf7"),
		AvailabilityDomain: &cfg.AvailabilityDomain,
		CompartmentId:      &cfg.CompartmentId,
		Region:             &cfg.Region,
		FreeformTags:       tags,
		LifecycleState:     core.InstanceLifecycleStateRunning,
	}
	mockComputeClient.On("ListInstances", ctx, core.ListInstancesRequest{
		CompartmentId: &cfg.CompartmentId,
	}).Return(core.ListInstancesResponse{
		Items: []core.Instance{expectedInstance},
	}, nil)

	instance, err := ociCli.FindInstanceByTags(ctx, tags)

	assert.Nil(t, err)
	assert.Equal(t, &expectedInstance, instance)
}
