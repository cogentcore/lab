// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cogentcore.org/core/base/sshclient"
)

// scpProgress returns a function that shows the progress of a copy of
// [Goal.Scp] on the given writer, if it is a terminal, as a line that it
// draws again at intervals, and clears at the end, or nil if it is not
// a terminal.
func scpProgress(w io.Writer) func(p sshclient.Progress) {
	f, ok := w.(*os.File)
	if !ok || !isTerminal(f) {
		return nil
	}
	width := termWidth()
	start := time.Now()
	var last time.Time
	file := 0
	return func(p sshclient.Progress) {
		if p.Finished {
			if !last.IsZero() {
				fmt.Fprint(f, "\r\033[K")
			}
			return
		}
		// at most 10 times a second, and when a file starts
		now := time.Now()
		if now.Sub(last) < 100*time.Millisecond && p.FileNum == file {
			return
		}
		last, file = now, p.FileNum
		fmt.Fprint(f, "\r"+progressLine(p, now.Sub(start), width)+"\033[K")
	}
}

// progressLine returns the line that shows the given progress of a copy,
// after the given time, in the given width: the number and name of the
// file being copied, and a bar, percent, bytes, rate and time remaining
// for all of the files, as
//
//	[2/5] name.tsv [=========>          ]  45%  12.3 MB/27.0 MB  8.1 MB/s  0:02
func progressLine(p sshclient.Progress, elapsed time.Duration, width int) string {
	frac := 1.0
	if p.Total > 0 {
		frac = float64(p.TotalDone) / float64(p.Total)
	}
	const barWidth = 20
	bar := strings.Repeat("=", int(frac*barWidth))
	if len(bar) < barWidth {
		bar += ">" + strings.Repeat(" ", barWidth-len(bar)-1)
	}
	rate := 0.0
	if s := elapsed.Seconds(); s > 0 {
		rate = float64(p.TotalDone) / s
	}
	eta := "-:--"
	if rate > 0 {
		left := time.Duration(float64(p.Total-p.TotalDone) / rate * float64(time.Second))
		eta = fmt.Sprintf("%d:%02d", int(left.Minutes()), int(left.Seconds())%60)
	}
	count := ""
	if p.NFiles > 1 {
		count = fmt.Sprintf("[%d/%d] ", p.FileNum, p.NFiles)
	}
	pct := fmt.Sprintf(" %3d%%", int(frac*100))
	bytes := fmt.Sprintf("  %s/%s", formatBytes(p.TotalDone), formatBytes(p.Total))
	speed := fmt.Sprintf("  %s/s", formatBytes(int64(rate)))
	// What fits in the width, with room for the name: on a narrow terminal,
	// without the bytes, then the rate, then the bar.
	name := filepath.Base(p.File)
	var rest string
	for _, r := range []string{
		" [" + bar + "]" + pct + bytes + speed + "  " + eta,
		" [" + bar + "]" + pct + speed + "  " + eta,
		" [" + bar + "]" + pct + "  " + eta,
		pct + "  " + eta,
	} {
		rest = r
		if width-1-len(count)-len(r) >= min(len(name), 12) {
			break
		}
	}
	if room := width - 1 - len(count) - len(rest); len(name) > room {
		name = name[:max(room-3, 0)] + "..."
	}
	return count + name + rest
}

// formatBytes returns the given number of bytes in units of 1024 bytes,
// such as 12.3 MB.
func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	v, units := float64(n), "KMGT"
	for i := range units {
		v /= unit
		if v < unit || i == len(units)-1 {
			return fmt.Sprintf("%.1f %cB", v, units[i])
		}
	}
	return ""
}

// isTerminal returns whether the given file is a terminal.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// termWidth returns the width of the terminal: that of the COLUMNS
// environment variable if it is set, and otherwise 80.
func termWidth() int {
	if w, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && w > 20 {
		return w
	}
	return 80
}
