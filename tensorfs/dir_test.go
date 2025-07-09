// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorfs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDir(t *testing.T) {
	dir, err := NewDir("root")
	assert.NoError(t, err)

	mdir := dir.Dir("multi/path/deep")
	ls := dir.ListAll()
	// fmt.Println(ls)
	lsc :=
		`multi/
	path/
		deep/
`
	assert.Equal(t, lsc, ls)
	assert.Equal(t, "deep", mdir.Name())
}

func TestDirTable(t *testing.T) {
	dir, err := NewDir("root")
	assert.NoError(t, err)

	mdir := dir.Dir("multi/path/deep")
	mdir.Float64("data", 3, 3)
	bdir := dir.Dir("multi/path/next")
	bdir.Float64("dat", 3)

	ls := dir.ListAll()
	// fmt.Println(ls)
	lsc :=
		`multi/
	path/
		deep/
			data [3 3]
		next/
			dat [3]
`
	assert.Equal(t, lsc, ls)

	ts := `#deep/data[1:0]<1:3>	#deep/data[1:1]	#deep/data[1:2]	#next/dat
0	0	0	0
0	0	0	0
0	0	0	0
`
	mpd := dir.Dir("multi/path")

	dt := DirTable(mpd, nil)
	ds := dt.String()
	// fmt.Println(ds)
	assert.Equal(t, ts, ds)

	ndir, err := NewDir("root")
	assert.NoError(t, err)

	ll :=
		`deep/
	data [3 3]
next/
	dat [3]
`

	DirFromTable(ndir, dt)
	nls := ndir.ListAll()
	// fmt.Println(nls)
	assert.Equal(t, ll, nls)
}

func TestDirTar(t *testing.T) {
	dir, err := NewDir("root")
	assert.NoError(t, err)

	mdir := dir.Dir("multi/path/deep")
	mdir.Float64("data", 3, 3)
	bdir := dir.Dir("multi/path/next")
	bdir.Float64("dat", 3)

	ls := dir.ListAll()
	// fmt.Println(ls)
	lsc :=
		`multi/
	path/
		deep/
			data [3 3]
		next/
			dat [3]
`
	assert.Equal(t, lsc, ls)

	gz := true

	var b bytes.Buffer
	err = Tar(&b, dir, gz, nil)
	assert.NoError(t, err)

	// fmt.Println(b.Len())

	ndir, err := NewDir("root")
	assert.NoError(t, err)

	err = Untar(&b, ndir, gz)
	assert.NoError(t, err)

	nls := ndir.ListAll()
	// fmt.Println(nls)
	assert.Equal(t, lsc, nls)
}
