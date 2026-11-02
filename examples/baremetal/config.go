// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"os"

	"cogentcore.org/core/base/iox/tomlx"
)

// Config is the configuration of baremetal in its config.toml file
// (see [BareMetal.ConfigFile]), for editing it.
type Config struct {
	// Servers are the server machines, in order of use preference.
	Servers []*Server
}

// configFile is the format of the config.toml file, which has a
// [[Servers.Values]] header for each server, as [BareMetal.Servers].
type configFile struct {
	Servers struct {
		Values []*Server
	}
}

// Open opens the config from the given file.
func (cf *Config) Open(file string) error {
	f := &configFile{}
	if err := tomlx.Open(f, file); err != nil {
		return err
	}
	cf.Servers = f.Servers.Values
	return nil
}

// Bytes returns the config in the format of the config.toml file.
func (cf *Config) Bytes() ([]byte, error) {
	f := &configFile{}
	f.Servers.Values = cf.Servers
	return tomlx.WriteBytes(f)
}

// Save saves the config to the given file.
func (cf *Config) Save(file string) error {
	b, err := cf.Bytes()
	if err != nil {
		return err
	}
	return os.WriteFile(file, b, 0666)
}
