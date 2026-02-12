module cogentcore.org/lab

go 1.25.6

// after go mod tidy, you need to run:
// go get google.golang.org/genproto@latest
// otherwise there will be an ambiguous import warning when building baremetal
// https://github.com/googleapis/go-genproto/issues/1015

require (
	cogentcore.org/core v0.3.17
	github.com/cogentcore/readline v0.1.3
	github.com/cogentcore/yaegi v0.0.0-20260116172027-700fbf8949f3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nsf/termbox-go v1.1.1
	github.com/stretchr/testify v1.11.1
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96
	golang.org/x/tools v0.41.0
	gonum.org/v1/gonum v0.17.0
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
)

require (
	codeberg.org/go-pdf/fpdf v0.11.0 // indirect
	github.com/Bios-Marcel/wastebasket/v2 v2.0.3 // indirect
	github.com/Masterminds/vcs v1.13.3 // indirect
	github.com/adrg/strutil v0.3.1 // indirect
	github.com/alecthomas/chroma/v2 v2.23.0 // indirect
	github.com/anthonynsimon/bild v0.14.0 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bramvdbogaerde/go-scp v1.6.0 // indirect
	github.com/chewxy/math32 v1.11.1 // indirect
	github.com/cogentcore/webgpu v0.23.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/ericchiang/css v1.4.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-fonts/latin-modern v0.3.3 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20250301202403-da16c1255728 // indirect
	github.com/go-text/typesetting v0.3.1-0.20250707124828-58cd3ef91251 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gomarkdown/markdown v0.0.0-20250810172220-2e2c11897d1a // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hack-pad/go-indexeddb v0.3.2 // indirect
	github.com/hack-pad/hackpadfs v0.2.4 // indirect
	github.com/hack-pad/safejs v0.1.1 // indirect
	github.com/jeandeaual/go-locale v0.0.0-20250612000132-0ef82f21eade // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/tdewolff/parse/v2 v2.8.5 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/image v0.35.0 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto v0.0.0-20260209200024-4cfbd4190f57 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/knuth v0.5.4 // indirect
	modernc.org/token v1.1.0 // indirect
	star-tex.org/x/tex v0.7.1 // indirect
)
