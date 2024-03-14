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

	"github.com/oracle/oci-go-sdk/v49/core"
	"github.com/stretchr/testify/mock"
)

type MockComputeClient struct {
	mock.Mock
}

func (m *MockComputeClient) LaunchInstance(ctx context.Context, request core.LaunchInstanceRequest) (core.LaunchInstanceResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(core.LaunchInstanceResponse), args.Error(1)
}

func (m *MockComputeClient) GetInstance(ctx context.Context, request core.GetInstanceRequest) (core.GetInstanceResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(core.GetInstanceResponse), args.Error(1)
}

func (m *MockComputeClient) TerminateInstance(ctx context.Context, request core.TerminateInstanceRequest) (core.TerminateInstanceResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(core.TerminateInstanceResponse), args.Error(1)
}

func (m *MockComputeClient) ListInstances(ctx context.Context, request core.ListInstancesRequest) (core.ListInstancesResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(core.ListInstancesResponse), args.Error(1)
}

func (m *MockComputeClient) InstanceAction(ctx context.Context, request core.InstanceActionRequest) (core.InstanceActionResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(core.InstanceActionResponse), args.Error(1)
}
