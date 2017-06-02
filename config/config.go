// Package config loads OpenStack configuration from a clouds.yaml file.
package config

import (
	"errors"
	"github.com/gophercloud/gophercloud"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/user"
	"path/filepath"
)

// Config represents configuration data for all clouds defined in clouds.yaml.
// Its methods are safe for concurrent use by multiple goroutines.
type Config interface {
	// TODO: Change Get to Cloud, and GetAll to AllClouds

	// Get returns configuration for one cloud by name. If the cloud is not
	// defined, this returns an error.
	Get(name string) (gophercloud.AuthOptions, error)

	// GetAll returns a map of all cloud configurations keyed by name. If
	// no clouds are defined, this returns nil.
	GetAll() map[string]gophercloud.AuthOptions
}

// configImpl implements the Config interface.
type configImpl struct {
	clouds map[string]gophercloud.AuthOptions
}

// Get satisfies the Config interface.
func (c *configImpl) Get(name string) (gophercloud.AuthOptions, error) {
	if v, ok := c.clouds[name]; ok {
		return v, nil
	}
	err := errors.New("config: cloud `" + name + "` not found")
	return gophercloud.AuthOptions{}, err
}

// GetAll satisfies the Config interface.
func (c *configImpl) GetAll() map[string]gophercloud.AuthOptions {
	if len(c.clouds) == 0 {
		return nil
	}
	cs := map[string]gophercloud.AuthOptions{}
	for k, v := range c.clouds {
		cs[k] = v
	}
	return cs
}

// New returns an initialized *Config.
//
// This searches for a clouds.yaml file in the following directories:
//
//        1) current directory
//        2) ~/.config/openstack
//        3) /etc/openstack
//
// The first valid clouds.yaml file found wins. (See the documentation at
// http://docs.openstack.org/developer/os-client-config/)
//
// New returns an error if a suitable clouds.yaml file is not found.
//
// To specify a file directly rather than searching known paths, use FromFile.
func New() (Config, error) {
	paths, err := getDefaultPaths()
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		conf, err := FromFile(p)
		if err == nil {
			return conf, nil
		}
		// Return an error if cloud.yaml is not well-formed; otherwise,
		// just continue to the next file.
		if parseErr, ok := err.(*ParseError); ok {
			return nil, parseErr
		}
	}
	return nil, errors.New("config: no usable clouds.yaml file found")
}

// FromFile returns an initialized *Config from a given clouds.yaml file. This
// returns an error if the file cannot be read or is in an invalid format.
func FromFile(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("config: " + err.Error())
	}

	type authYAML struct {
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		TenantName string `yaml:"tenant_name"`
		TenantID   string `yaml:"tenant_id"`
		AuthURL    string `yaml:"auth_url"`
	}
	y := map[string]map[string]map[string]*authYAML{}
	if err := yaml.Unmarshal(b, y); err != nil {
		return nil, &ParseError{path, err}
	}
	if len(y["clouds"]) == 0 {
		return nil, &ParseError{path, errors.New("config is empty")}
	}

	clouds := map[string]gophercloud.AuthOptions{}
	for k, v := range y["clouds"] {
		if a, ok := v["auth"]; ok {
			clouds[k] = gophercloud.AuthOptions{
				IdentityEndpoint: a.AuthURL,
				Password:         a.Password,
				TenantID:         a.TenantID,
				TenantName:       a.TenantName,
				Username:         a.Username,
			}
		}
	}
	return &configImpl{clouds: clouds}, nil
}

// getDefaultPaths returns a list of directories that OpenStack searches by
// default for clouds.yaml files. This returns an error if the user’s home
// directory cannot be discovered.
func getDefaultPaths() ([]string, error) {
	u, err := user.Current()
	if err != nil {
		s := "config: cannot find home directory: " + err.Error()
		return nil, errors.New(s)
	}
	homeDir := u.HomeDir
	if homeDir == "" {
		return nil, errors.New("config: $HOME env var not set")
	}
	f := "clouds.yaml"
	return []string{
		filepath.Join("./", f),
		filepath.Join(homeDir, ".config/openstack", f),
		filepath.Join("/etc/openstack", f),
	}, nil
}

// ParseError represents an error parsing a clouds.yaml file.
type ParseError struct {
	File string
	Err  error
}

func (e *ParseError) Error() string {
	msg := "config: cannot parse " + e.File
	if e.Err != nil {
		return msg + ": " + e.Err.Error()
	}
	return msg
}
