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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test.toml")
	require.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tempFile.Name())

	// Write some dummy TOML data to the temp file
	dummyTOML := `
		availability_domain = "mQqX:US-ASHBURN-AD-2"
		compartment_id = "ocid1.compartment.oc1...fsbq"
		subnet_id = "ocid1.subnet.oc1.iad....feoplaka"
		network_security_group_id = "ocid1.networksecuritygroup....pfzya"
		tenancy_id = "ocid1.tenancy.oc1..aaaaaaaajds7tbqbvrcaiavm2uk34t7wke7jg75aemsacljymbjxcio227oq"
		user_id = "ocid1.user.oc1...ug6l37u6a"
		region = "us-ashburn-1"
		fingerprint = "38...6f:bb"
		private_key_path = "/home/ubuntu/.oci/private_key.pem"
		private_key_password = ""
	`

	_, err = tempFile.Write([]byte(dummyTOML))
	require.NoError(t, err, "Failed to write to temp file")

	err = tempFile.Close()
	require.NoError(t, err, "Failed to close temp file")

	// Test case for successful read
	t.Run("success", func(t *testing.T) {
		got, err := NewConfig(tempFile.Name())
		require.NoError(t, err, "NewConfig() should not have returned an error")
		require.Equal(t, &Config{
			AvailabilityDomain: "mQqX:US-ASHBURN-AD-2",
			CompartmentId:      "ocid1.compartment.oc1...fsbq",
			SubnetID:           "ocid1.subnet.oc1.iad....feoplaka",
			NsgID:              "ocid1.networksecuritygroup....pfzya",
			TenancyID:          "ocid1.tenancy.oc1..aaaaaaaajds7tbqbvrcaiavm2uk34t7wke7jg75aemsacljymbjxcio227oq",
			UserID:             "ocid1.user.oc1...ug6l37u6a",
			Region:             "us-ashburn-1",
			Fingerprint:        "38...6f:bb",
			PrivateKeyPath:     "/home/ubuntu/.oci/private_key.pem",
			PrivateKeyPassword: "",
		}, got, "NewConfig() returned unexpected content")
	})

	// Test case for failed read (file does not exist)
	t.Run("fail", func(t *testing.T) {
		_, err := NewConfig("nonexistent.toml")
		require.Error(t, err, "NewConfig() expected an error, got none")
	})

	// Test case for failed read (invalid TOML)
	t.Run("fail", func(t *testing.T) {
		// Create a temporary file
		tempFile, err := os.CreateTemp("", "test.toml")
		require.NoError(t, err, "Failed to create temp file")
		defer os.Remove(tempFile.Name())

		// Write some invalid TOML data to the temp file
		invalidTOML := "invalid TOML"
		_, err = tempFile.Write([]byte(invalidTOML))
		require.NoError(t, err, "Failed to write to temp file")

		err = tempFile.Close()
		require.NoError(t, err, "Failed to close temp file")

		_, err = NewConfig(tempFile.Name())
		require.Error(t, err, "NewConfig() expected an error, got none")
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		errString error
	}{
		{
			name: "valid config",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				UserID:             "user",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: nil,
		},
		{
			name: "missing availability domain",
			config: &Config{
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				UserID:             "user",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("availability_domain is required"),
		},
		{
			name: "missing compartment id",
			config: &Config{
				AvailabilityDomain: "ad",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				UserID:             "user",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("compartment_id is required"),
		},
		{
			name: "missing subnet id",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				UserID:             "user",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("subnet_id is required"),
		},
		{
			name: "missing nsg id",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				TenancyID:          "tenancy",
				UserID:             "user",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("ngs_id is required"),
		},
		{
			name: "missing tenancy id",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				UserID:             "user",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("tenancy_id is required"),
		},
		{
			name: "missing user id",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("user_id is required"),
		},
		{
			name: "missing region",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				UserID:             "user",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("region is required"),
		},
		{
			name: "missing fingerprint",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				UserID:             "user",
				Region:             "region",
				PrivateKeyPath:     "path",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("fingerprint is required"),
		},
		{
			name: "missing private key path",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				UserID:             "user",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPassword: "password",
			},
			errString: fmt.Errorf("private_key_path is required"),
		},
		{
			name: "valid config with empty private key password",
			config: &Config{
				AvailabilityDomain: "ad",
				CompartmentId:      "compartment",
				SubnetID:           "subnet",
				NsgID:              "nsg",
				TenancyID:          "tenancy",
				UserID:             "user",
				Region:             "region",
				Fingerprint:        "fingerprint",
				PrivateKeyPath:     "path",
			},
			errString: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			require.Equal(t, tt.errString, err)
		})
	}

}

func TestGetPrivateKey(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test.pem")
	require.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tempFile.Name())

	// Write some dummy PEM data to the temp file
	dummyPEM := "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqh...\n-----END PRIVATE KEY-----"
	_, err = tempFile.Write([]byte(dummyPEM))
	require.NoError(t, err, "Failed to write to temp file")

	err = tempFile.Close()
	require.NoError(t, err, "Failed to close temp file")

	// Test case for successful read
	t.Run("success", func(t *testing.T) {
		c := Config{PrivateKeyPath: tempFile.Name()}
		got, err := c.GetPrivateKey()
		require.NoError(t, err, "GetPrivateKey() should not have returned an error")
		require.Equal(t, dummyPEM, got, "GetPrivateKey() returned unexpected content")
	})

	// Test case for failed read (file does not exist)
	t.Run("fail", func(t *testing.T) {
		c := Config{PrivateKeyPath: "nonexistent.pem"}
		_, err := c.GetPrivateKey()
		require.Error(t, err, "GetPrivateKey() expected an error, got none")
	})
}
