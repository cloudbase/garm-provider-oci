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

package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func NewConfig(cfgFile string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(cfgFile, &config); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}
	return &config, nil
}

type Config struct {
	AvailabilityDomain string `toml:"availability_domain"`
	CompartmentId      string `toml:"compartment_id"`
	SubnetID           string `toml:"subnet_id"`
	NsgID              string `toml:"network_security_group_id"`
	TenancyID          string `toml:"tenancy_id"`
	UserID             string `toml:"user_id"`
	Region             string `toml:"region"`
	Fingerprint        string `toml:"fingerprint"`
	PrivateKeyPath     string `toml:"private_key_path"`
	PrivateKeyPassword string `toml:"private_key_password"`
}

func (c *Config) Validate() error {
	if c.AvailabilityDomain == "" {
		return fmt.Errorf("availability_domain is required")
	}
	if c.CompartmentId == "" {
		return fmt.Errorf("compartment_id is required")
	}
	if c.SubnetID == "" {
		return fmt.Errorf("subnet_id is required")
	}
	if c.NsgID == "" {
		return fmt.Errorf("ngs_id is required")
	}
	if c.TenancyID == "" {
		return fmt.Errorf("tenancy_id is required")
	}
	if c.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if c.Region == "" {
		return fmt.Errorf("region is required")
	}
	if c.Fingerprint == "" {
		return fmt.Errorf("fingerprint is required")
	}
	if c.PrivateKeyPath == "" {
		return fmt.Errorf("private_key_path is required")
	}
	return nil
}

func (c *Config) GetPrivateKey() (string, error) {
	pemFileContent, err := os.ReadFile(c.PrivateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read the .pem file: %v", err)
	}
	return string(pemFileContent), nil
}
