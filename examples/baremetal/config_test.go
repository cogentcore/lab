// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig tests that a Config is opened from and saved to the format
// of the config.toml file, which BareMetal loads its servers from.
func TestConfig(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "BareMetal", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0777))
	require.NoError(t, os.WriteFile(file, []byte(`[[Servers.Values]]
	Name = "h100"
	SSH = "user@h100.example.com"
	NGPUs = 8
`), 0666))
	cf := &Config{}
	require.NoError(t, cf.Open(file))
	require.Len(t, cf.Servers, 1)
	assert.Equal(t, Server{Name: "h100", SSH: "user@h100.example.com", NGPUs: 8}, *cf.Servers[0])

	cf.Servers = append(cf.Servers, &Server{Name: "a100", SSH: "user@a100.example.com", NGPUs: 4})
	require.NoError(t, cf.Save(file))
	bm := NewBareMetal()
	bm.DataDir = dir
	assert.Equal(t, file, bm.ConfigFile())
	bm.config()
	require.Equal(t, []string{"h100", "a100"}, bm.Servers.Keys)
	sv := bm.Servers.At("a100")
	assert.Equal(t, "user@a100.example.com", sv.SSH)
	assert.Equal(t, 4, sv.NGPUs)
}
