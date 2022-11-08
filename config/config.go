/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	exConfig "github.com/deislabs/ratify/pkg/executor/config"
	"github.com/deislabs/ratify/pkg/homedir"
	"github.com/deislabs/ratify/pkg/policyprovider"
	pcConfig "github.com/deislabs/ratify/pkg/policyprovider/config"
	pf "github.com/deislabs/ratify/pkg/policyprovider/factory"
	"github.com/deislabs/ratify/pkg/referrerstore"
	rsConfig "github.com/deislabs/ratify/pkg/referrerstore/config"
	sf "github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/deislabs/ratify/pkg/verifier"
	vfConfig "github.com/deislabs/ratify/pkg/verifier/config"
	vf "github.com/deislabs/ratify/pkg/verifier/factory"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	ConfigFileName = "config.json"
	ConfigFileDir  = ".ratify"
	PluginsFolder  = "plugins"
)

type Config struct {
	StoresConfig    rsConfig.StoresConfig    `json:"store,omitempty"`
	PoliciesConfig  pcConfig.PoliciesConfig  `json:"policy,omitempty"`
	VerifiersConfig vfConfig.VerifiersConfig `json:"verifier,omitempty"`
	ExecutorConfig  exConfig.ExecutorConfig  `json:"executor,omitempty"`
	fileHash        string                   `json:"-"`
}

var (
	initConfigDir         = new(sync.Once)
	homeDir               string
	configDir             string
	defaultConfigFilePath string
	defaultPluginsPath    string
)

func InitDefaultPaths() {
	if configDir != "" {
		return
	}
	configDir = os.Getenv("RATIFY_CONFIG")
	if configDir == "" {
		configDir = filepath.Join(getHomeDir(), ConfigFileDir)

	}
	defaultPluginsPath = filepath.Join(configDir, PluginsFolder)
	defaultConfigFilePath = filepath.Join(configDir, ConfigFileName)
}

func getHomeDir() string {
	if homeDir == "" {
		homeDir = homedir.Get()
	}
	return homeDir
}

// Returns created referer store, verifier, policyprovider objects from config
func CreateFromConfig(cf Config) ([]referrerstore.ReferrerStore, []verifier.ReferenceVerifier, policyprovider.PolicyProvider, error) {
	stores, err := sf.CreateStoresFromConfig(cf.StoresConfig, GetDefaultPluginPath())

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to load store from config")
	}
	logrus.Infof("stores successfully created. number of stores %d", len(stores))

	verifiers, err := vf.CreateVerifiersFromConfig(cf.VerifiersConfig, GetDefaultPluginPath())

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to load verifiers from config")
	}

	logrus.Infof("verifiers successfully created. number of verifiers %d", len(verifiers))

	policyEnforcer, err := pf.CreatePolicyProviderFromConfig(cf.PoliciesConfig)

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to load policy provider from config")
	}

	logrus.Infof("policies successfully created.")

	return stores, verifiers, policyEnforcer, nil
}

// Load the config from file path provided, read from default path if configFilePath is empty
func Load(configFilePath string) (Config, error) {

	config := Config{}

	body, readErr := ioutil.ReadFile(getConfigurationFile(configFilePath))

	if readErr != nil {
		return config, fmt.Errorf("unable to read config file at path %s", readErr)
	}

	err := json.Unmarshal(body, &config)
	if err != nil {
		return config, fmt.Errorf("unable to unmarshal config body: %v", err)
	}

	config.fileHash, err = getFileHash(body)
	if err != nil {
		return config, fmt.Errorf("error getting configuration file hash error %v", err)
	}

	return config, nil
}

func GetDefaultPluginPath() string {
	if defaultPluginsPath == "" {
		initConfigDir.Do(InitDefaultPaths)
	}
	return defaultPluginsPath
}

// if configFilePath is empty, return configuration path from environment variable
func getConfigurationFile(configFilePath string) string {
	if configFilePath == "" {
		if configDir == "" {
			initConfigDir.Do(InitDefaultPaths)
		}

		return defaultConfigFilePath
	}
	return configFilePath
}

func getFileHash(file []byte) (fileHash string, err error) {
	hash := sha256.New()

	length, err := hash.Write(file)
	if err != nil || length == 0 {
		return "", fmt.Errorf("unable to generate hash for configuration file, err '%v', hash length %v", err, length)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
