// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

// todo: need significant updates to do testing here.
//
// func TestLocal(t *testing.T) {
// 	bm := NewBareMetal()
// 	bm.Config()
//
// 	assert.Equal(t, 1, len(bm.Servers.Values))
// 	assert.Equal(t, "hpc2", bm.Servers.Values[0].Name)
// 	assert.Equal(t, 2, bm.Servers.Values[0].NGPUs)
//
// 	bm.InitGoal()
// 	bm.InitServers()
// 	// todo: set log to file
//
// 	pwd, err := os.Getwd()
// 	assert.NoError(t, err)
// 	pwd, err = filepath.Abs(pwd)
// 	assert.NoError(t, err)
//
// 	td := filepath.Join(pwd, "testdata")
//
// 	var b bytes.Buffer
// 	err = TarFiles(&b, td, true, "script.sh")
// 	assert.NoError(t, err)
//
// 	job := bm.Submit("test", "tmp/bare/test", "script.sh", "*.tsv", b.Bytes())
// 	nrun, err := bm.RunPendingJobs()
// 	assert.NoError(t, err)
// 	assert.Equal(t, 1, nrun)
//
// 	_ = job
// 	// todo: query job
//
// 	for {
// 		nfin, err := bm.PollJobs()
// 		assert.NoError(t, err)
// 		if nfin == 1 {
// 			break
// 		}
// 	}
//
// }
