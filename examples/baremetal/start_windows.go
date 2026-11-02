// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import "os/exec"

// detach does nothing on windows, where the command keeps running
// after this process ends.
func detach(cmd *exec.Cmd) {}
