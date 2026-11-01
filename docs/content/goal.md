**Goal** is the _Go augmented language_ with support for two additional modes, in addition to standard Go:

* [[shell|$ shell mode $]] that operates like a standard command-line shell (e.g., `bash`), with space-separated elements and standard shell functionality including input / output redirection. Goal automatically detects most instances of shell mode based on the syntax of the line, but it can always be explicitly indicated with surrounding `$`s.

* [[math|# math mode #]] that supports Python-like concise mathematical expressions operating on [[tensor]] elements.

Here is an example of shell-mode mixing Go and shell code:

```goal
for i, f := range goalib.SplitLines($ls -la$) {  // ls executes, returns string
    echo {i} {strings.ToLower(f)}              // {} surrounds Go within shell
}
```

where Go-code is explicitly indicated by the `{}` braces.

Here is an example of math-mode:
```goal
# x := 1. / (1. + exp(-wts[:, :, :n] * acts[:]))
```

You can also intermix math within Go code:
```goal
for _, x := range #[1,2,3]# {
    fmt.Println(#x**2#)
}
```

Goal can be used in an interpreted mode by using the [mvm](https://github.com/cogentcore/mvm) Go interpreter (and can be used as your shell executable in a terminal), and it can also replace the standard `go` compiler in command-line mode, to build compiled executables using the extended Goal syntax.

A key design feature of Goal is that it always _transpiles directly to Go_ in a purely syntactically driven way, so the output of Goal is pure Go code.

Goal can also be used in conjunction with [[gosl]] to build programs that transparently run on GPU hardware in addition to standard CPUs (as standard Go programs).

## Install

```shell
go install cogentcore.org/lab/goal/cmd/goal@latest
```

## Editor support

The `goalimports` tool (in [goalimports](goal/cmd/goalimports)) reformats a Goal file in the same way that the [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) tool does, fixing the imports section and reformatting the code into the standard Go format style. You can configure your editor to run this on `.goal` files. It uses the offical Go version, run on the transpiled Go code, and projects any changes back to the .goal file, thus avoiding the need for a full Goal reformatting / parsing system.

The `goalls` is a version of the [gopls](https://pkg.go.dev/golang.org/x/tools/gopls) language server implementing the [LSP](https://microsoft.github.io/language-server-protocol/) language server protocol for Goal files. It directly uses gopls, operating on the transpiled Go code, and projects everything back into the Goal source, again avoiding the need for a full Goal language version. It should be updated whenever gopls is updated, and can be used as a replacement for gopls because it handless both Go and Goal files.

To install:

```sh
go install cogentcore.org/lab/goal/cmd/goalimports@latest
go install cogentcore.org/lab/goal/cmd/goalls@latest
```

These are also installed by the `core` tool in the [CogentCore](https://cogentcore.org/core) system.

## Goal pages

