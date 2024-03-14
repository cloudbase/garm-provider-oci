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
	"errors"
	"fmt"
	"strings"

	garmErrors "github.com/cloudbase/garm-provider-common/errors"
	"github.com/cloudbase/garm-provider-oci/config"
	"github.com/cloudbase/garm-provider-oci/internal/spec"
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/core"
)

func NewOciCli(ctx context.Context, cfg *config.Config) (*OciCli, error) {
	privateKey, err := cfg.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("error getting private key: %w", err)
	}
	confProvider := common.NewRawConfigurationProvider(
		cfg.TenancyID,
		cfg.UserID,
		cfg.Region,
		cfg.Fingerprint,
		privateKey,
		common.String(cfg.PrivateKeyPassword),
	)
	computeClient, err := core.NewComputeClientWithConfigurationProvider(confProvider)
	if err != nil {
		return nil, fmt.Errorf("error creating compute client: %w", err)
	}
	return &OciCli{
		computeClient: computeClient,
		cfg:           cfg,
	}, nil
}

type ClientInterface interface {
	LaunchInstance(ctx context.Context, request core.LaunchInstanceRequest) (core.LaunchInstanceResponse, error)
	GetInstance(ctx context.Context, request core.GetInstanceRequest) (core.GetInstanceResponse, error)
	TerminateInstance(ctx context.Context, request core.TerminateInstanceRequest) (core.TerminateInstanceResponse, error)
	ListInstances(ctx context.Context, request core.ListInstancesRequest) (core.ListInstancesResponse, error)
	InstanceAction(ctx context.Context, request core.InstanceActionRequest) (core.InstanceActionResponse, error)
}

type OciCli struct {
	cfg           *config.Config
	computeClient ClientInterface
}

func (o *OciCli) Config() *config.Config {
	return o.cfg
}

func (o *OciCli) ComputeClient() ClientInterface {
	return o.computeClient
}

func (o *OciCli) SetConfig(cfg *config.Config) {
	o.cfg = cfg
}

func (o *OciCli) SetComputeClient(computeClient ClientInterface) {
	o.computeClient = computeClient
}

func (o *OciCli) CreateInstance(ctx context.Context, spec *spec.RunnerSpec) (core.Instance, error) {
	req := core.LaunchInstanceRequest{
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
	}
	response, err := o.computeClient.LaunchInstance(ctx, req)
	if err != nil {
		return core.Instance{}, fmt.Errorf("error creating instance: %w", err)
	}
	return response.Instance, nil
}

func (o *OciCli) GetInstance(ctx context.Context, instanceID string) (core.Instance, error) {
	var inst string
	if strings.HasPrefix(instanceID, "ocid1.instance") {
		inst = instanceID
	} else {
		tags := map[string]string{
			"Name": instanceID,
		}

		tmp, err := o.FindInstanceByTags(ctx, tags)
		if err != nil {
			if errors.Is(err, garmErrors.ErrNotFound) {
				return core.Instance{}, fmt.Errorf("instance not found")
			}
			return core.Instance{}, fmt.Errorf("failed to determine instance: %w", err)
		}
		inst = *tmp.Id
	}
	req := core.GetInstanceRequest{
		InstanceId: &inst,
	}
	resp, err := o.computeClient.GetInstance(ctx, req)
	if err != nil {
		return core.Instance{}, fmt.Errorf("error getting instance: %w", err)
	}

	return resp.Instance, nil
}

func (o *OciCli) DeleteInstance(ctx context.Context, instanceID string) error {
	var inst string
	if strings.HasPrefix(instanceID, "ocid1.instance") {
		inst = instanceID
	} else {
		tags := map[string]string{
			"Name": instanceID,
		}

		tmp, err := o.FindInstanceByTags(ctx, tags)
		if err != nil {
			if errors.Is(err, garmErrors.ErrNotFound) {
				return nil
			}
			return fmt.Errorf("failed to determine instance: %w", err)
		}
		inst = *tmp.Id
	}

	request := core.TerminateInstanceRequest{
		InstanceId: &inst,
	}

	_, err := o.computeClient.TerminateInstance(ctx, request)
	if err != nil {
		return fmt.Errorf("error terminating instance: %w", err)
	}
	return nil
}

func (o *OciCli) ListInstances(ctx context.Context, poolID string) ([]core.Instance, error) {
	request := core.ListInstancesRequest{
		CompartmentId: &o.cfg.CompartmentId,
	}
	computeInstances, err := o.computeClient.ListInstances(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing instances: %w", err)
	}
	instances := []core.Instance{}
	for _, instance := range computeInstances.Items {
		if instance.FreeformTags["GARM_POOL_ID"] == poolID && instance.LifecycleState != core.InstanceLifecycleStateTerminated {
			instances = append(instances, instance)
		}
	}
	return instances, nil
}

func (o *OciCli) StopInstance(ctx context.Context, instanceID string) error {
	req := core.InstanceActionRequest{
		Action:     core.InstanceActionActionStop,
		InstanceId: &instanceID,
	}
	_, err := o.computeClient.InstanceAction(ctx, req)
	if err != nil {
		return fmt.Errorf("error stopping instance: %w", err)
	}
	return nil
}

func (o *OciCli) StartInstance(ctx context.Context, instanceID string) error {
	req := core.InstanceActionRequest{
		Action:     core.InstanceActionActionStart,
		InstanceId: &instanceID,
	}
	_, err := o.computeClient.InstanceAction(ctx, req)
	if err != nil {
		return fmt.Errorf("error starting instance: %w", err)
	}
	return nil
}

func (o *OciCli) FindInstanceByTags(ctx context.Context, tags map[string]string) (*core.Instance, error) {
	request := core.ListInstancesRequest{
		CompartmentId: &o.cfg.CompartmentId,
	}
	computeInstances, err := o.computeClient.ListInstances(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing instances: %w", err)
	}
	for _, instance := range computeInstances.Items {
		if instance.LifecycleState != core.InstanceLifecycleStateTerminated {
			for key, value := range tags {
				if instance.FreeformTags[key] != value {
					return nil, nil
				}
			}
			return &instance, nil
		}
	}
	return nil, nil
}
