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
	"fmt"

	"github.com/cloudbase/garm-provider-common/execution"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-oci/config"
	"github.com/cloudbase/garm-provider-oci/internal/client"
	"github.com/cloudbase/garm-provider-oci/internal/spec"
	"github.com/cloudbase/garm-provider-oci/internal/util"
)

var _ execution.ExternalProvider = &OciProvider{}

func NewOciProvider(ctx context.Context, cfgFile string, controllerID string) (*OciProvider, error) {
	conf, err := config.NewConfig(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	ociCli, err := client.NewOciCli(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("error creating oci client: %w", err)
	}
	return &OciProvider{
		cfg:          conf,
		ociCli:       ociCli,
		controllerID: controllerID,
	}, nil
}

type OciProvider struct {
	cfg          *config.Config
	ociCli       *client.OciCli
	controllerID string
}

func (o *OciProvider) CreateInstance(ctx context.Context, bootstrapParams params.BootstrapInstance) (params.ProviderInstance, error) {
	spec, err := spec.GetRunnerSpecFromBootstrapParams(o.cfg, bootstrapParams, o.controllerID)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error getting runner spec: %w", err)
	}

	ociInstance, err := o.ociCli.CreateInstance(ctx, spec)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error creating instance: %w", err)
	}
	instance := params.ProviderInstance{
		ProviderID: *ociInstance.Id,
		Name:       spec.BootstrapParams.Name,
		OSType:     spec.BootstrapParams.OSType,
		OSArch:     spec.BootstrapParams.OSArch,
		Status:     "running",
	}
	return instance, nil
}

func (o *OciProvider) GetInstance(ctx context.Context, instanceID string) (params.ProviderInstance, error) {
	ociInstance, err := o.ociCli.GetInstance(ctx, instanceID)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error getting instance: %w", err)
	}
	providerInstance := util.OciInstanceToProviderInstance(ociInstance)
	return providerInstance, nil
}

func (o *OciProvider) DeleteInstance(ctx context.Context, instanceID string) error {
	return o.ociCli.DeleteInstance(ctx, instanceID)
}

func (o *OciProvider) ListInstances(ctx context.Context, poolID string) ([]params.ProviderInstance, error) {
	ociInstances, err := o.ociCli.ListInstances(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("error listing instances: %w", err)
	}
	var providerInstances []params.ProviderInstance
	for _, ociInstance := range ociInstances {
		providerInstances = append(providerInstances, util.OciInstanceToProviderInstance(ociInstance))
	}
	return providerInstances, nil
}

func (o *OciProvider) RemoveAllInstances(ctx context.Context) error {
	return nil
}

func (o *OciProvider) Stop(ctx context.Context, instance string, force bool) error {
	return o.ociCli.StopInstance(ctx, instance)
}

func (o *OciProvider) Start(ctx context.Context, instance string) error {
	return o.ociCli.StartInstance(ctx, instance)
}
