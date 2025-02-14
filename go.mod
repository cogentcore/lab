module cogentcore.org/lab

go 1.22.7

// important: after running go mod tidy, you must then run this:
// go get google.golang.org/genproto@latest
// otherwise there will be an ambiguous import warning when building baremetal
// https://github.com/googleapis/go-genproto/issues/1015

require (
	cogentcore.org/core v0.3.9-0.20250204223236-67d60a912a4b
	github.com/cogentcore/readline v0.1.3
	github.com/cogentcore/yaegi v0.0.0-20240724064145-e32a03faad56
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nsf/termbox-go v1.1.1
	github.com/stretchr/testify v1.9.0
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
	golang.org/x/tools v0.23.0
	gonum.org/v1/gonum v0.15.0
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.4
)

require (
	github.com/Bios-Marcel/wastebasket v0.0.4-0.20240213135800-f26f1ae0a7c4 // indirect
	github.com/Masterminds/vcs v1.13.3 // indirect
	github.com/adrg/strutil v0.3.1 // indirect
	github.com/alecthomas/chroma/v2 v2.13.0 // indirect
	github.com/anthonynsimon/bild v0.13.0 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bramvdbogaerde/go-scp v1.4.0 // indirect
	github.com/chewxy/math32 v1.10.1 // indirect
	github.com/cogentcore/webgpu v0.0.0-20250118183535-3dd1436165cf // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/ericchiang/css v1.3.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20240506104042-037f3cc74f2a // indirect
	github.com/goki/freetype v1.0.5 // indirect
	github.com/gomarkdown/markdown v0.0.0-20240930133441-72d49d9543d8 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hack-pad/go-indexeddb v0.3.2 // indirect
	github.com/hack-pad/hackpadfs v0.2.1 // indirect
	github.com/hack-pad/safejs v0.1.1 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/pelletier/go-toml/v2 v2.1.2-0.20240227203013-2b69615b5d55 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/image v0.18.0 // indirect
	golang.org/x/mod v0.19.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto v0.0.0-20250212204824-5a70512c5d8b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250124145028-65684f501c47 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
