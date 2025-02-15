// Copyright (c) 2020, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

func main() {
	err := termbox.Init()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer termbox.Close()

	TheFiles.Open(os.Args[1:])

	nf := len(TheFiles)
	if nf == 0 {
		fmt.Printf("usage: etail <filename>...  (space separated)\n")
		return
	}

	if nf > 1 {
		TheTerm.ShowFName = true
	}

	err = TheTerm.ToggleTail() // start in tail mode
	if err != nil {
		log.Println(err)
		panic(err)
	}

	Tailer := time.NewTicker(time.Duration(500) * time.Millisecond)
	go func() {
		for {
			<-Tailer.C
			TheTerm.TailCheck()
		}
	}()

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch {
			case ev.Key == termbox.KeyEsc || ev.Char == 'Q' || ev.Char == 'q':
				break loop
			case ev.Char == ' ' || ev.Char == 'n' || ev.Char == 'N' || ev.Key == termbox.KeyPgdn || ev.Key == termbox.KeySpace:
				TheTerm.NextPage()
			case ev.Char == 'p' || ev.Char == 'P' || ev.Key == termbox.KeyPgup:
				TheTerm.PrevPage()
			case ev.Key == termbox.KeyArrowDown:
				TheTerm.NextLine()
			case ev.Key == termbox.KeyArrowUp:
				TheTerm.PrevLine()
			case ev.Char == 'f' || ev.Char == 'F' || ev.Key == termbox.KeyArrowRight:
				TheTerm.ScrollRight()
			case ev.Char == 'b' || ev.Char == 'B' || ev.Key == termbox.KeyArrowLeft:
				TheTerm.ScrollLeft()
			case ev.Char == 'a' || ev.Char == 'A' || ev.Key == termbox.KeyHome:
				TheTerm.Top()
			case ev.Char == 'e' || ev.Char == 'E' || ev.Key == termbox.KeyEnd:
				TheTerm.End()
			case ev.Char == 'w' || ev.Char == 'W':
				TheTerm.FixRight()
			case ev.Char == 's' || ev.Char == 'S':
				TheTerm.FixLeft()
			case ev.Char == 'v' || ev.Char == 'V':
				TheTerm.FilesNext()
			case ev.Char == 'u' || ev.Char == 'U':
				TheTerm.FilesPrev()
			case ev.Char == 'm' || ev.Char == 'M':
				TheTerm.MoreMinLines()
			case ev.Char == 'l' || ev.Char == 'L':
				TheTerm.LessMinLines()
			case ev.Char == 'd' || ev.Char == 'D':
				TheTerm.ToggleNames()
			case ev.Char == 't' || ev.Char == 'T':
				TheTerm.ToggleTail()
			case ev.Char == 'c' || ev.Char == 'C':
				TheTerm.ToggleColNums()
			case ev.Char == 'h' || ev.Char == 'H':
				TheTerm.Help()
			}
		case termbox.EventResize:
			TheTerm.Draw()
		}
	}
}
