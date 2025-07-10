+++
Categories = ["Goal"]
+++

In general, the [[Goal]] shell mode behavior mimics that of `bash`.

The following documentation describes specific use-cases.

## Environment variables

* `set <var> <value>` (space delimited as in all shell mode, no equals)

## Output redirction

* Standard output redirect: `>` and `>&` (and `|`, `|&` if needed)

## Control flow

* Any error stops the script execution, except for statements wrapped in `[ ]`, indicating an "optional" statement, e.g.:

```sh
cd some; [mkdir sub]; cd sub
```

* `&` at the end of a statement runs in the background (as in bash) -- otherwise it waits until it completes before it continues.

* `jobs`, `fg`, `bg`, and `kill` builtin commands function as in usual bash.

## Shell functions (aliases)

Use the `command` keyword to define new functions for Shell mode execution, which can then be used like any other command, for example:

```sh
command list {
	ls -la args...
}
```

```sh
cd data
list *.tsv
```

The `command` is transpiled into a Go function that takes `args ...string` arguments. In the command function body, you can use the `args...` expression to pass all of the args, or `args[1]` etc to refer to specific positional indexes, as usual.

The command function name is registered so that the standard shell execution code can run the function, passing the args.  You can also call it directly from Go code using the standard parentheses expression.

## Script Files and Makefile-like functionality

As with most scripting languages, a file of goal code can be made directly executable by appending a "shebang" expression at the start of the file:

```sh
#!/usr/bin/env goal
```

When executed this way, any additional args passed on the command line invocation are available via the `goalrun.Args()` function, which can be passed to a command as follows:
```go
install {goalrun.Args()}
```
or by referring to specific arg indexes etc, as in `goalrun.Args()[0]`.  Note that these are not the same as the implicit `args` for command aliases, which are local function args in effect.

To make a script behave like a standard Makefile, you can define different `command`s for each of the make commands, and then add the following at the end of the file to use the args to run commands:

```go
goal.RunCommands(goalrun.Args())
```

See [make](cmd/goal/testdata/make) for an example, in `cmd/goal/testdata/make`, which can be run for example using:

```sh
./make build
```

Note that there is nothing special about the name `make` here, so this can be done with any file.

The `make` package defines a number of useful utility functions that accomplish the standard dependency and file timestamp checking functionality from the standard `make` command, as in the [magefile](https://magefile.org/dependencies/) system.  Note that the goal direct shell command syntax makes the resulting make files much closer to a standard bash-like Makefile, while still having all the benefits of Go control and expressions, compared to magefile.

TODO: implement and document above.

## SSH connections to remote hosts

Any number of active SSH connections can be maintained and used dynamically within a script, including simple ways of copying data among the different hosts (including the local host).  The Go mode execution is always on the local host in one running process, and only the shell commands are executed remotely, enabling a unique ability to easily coordinate and distribute processing and data across various hosts.

Each host maintains its own working directory and environment variables, which can be configured and re-used by default whenever using a given host.

* `gossh hostname.org [name]`  establishes a connection, using given optional name to refer to this connection.  If the name is not provided, a sequential number will be used, starting with 1, with 0 referring always to the local host.

* `@name` then refers to the given host in all subsequent commands, with `@0` referring to the local host where the goal script is running. 

* You can use a variable name for the server, like this (the explicit `$ $` shell mode is required because a line starting with `{` is not recognized as shell code):
```goal
server := "@myserver"
${server} ls$
```

### Explicit per-command specification of host

```sh
@name cd subdir; ls
```

### Default host

```sh
@name // or:
gossh @name
```

uses the given host for all subsequent commands (unless explicitly specified), until the default is changed.  Use `gossh @0` to return to localhost.

### Redirect input / output among hosts

The output of a remote host command can be sent to a file on the local host:
```sh
@name cat hostfile.tsv > @0:localfile.tsv
```
Note the use of the `:` colon delimiter after the host name here.  TODO: You cannot send output to a remote host file (e.g., `> @host:remotefile.tsv`) -- maybe with sftp?

The output of any command can also be piped to a remote host as its standard input:
```sh
ls *.tsv | @host cat > files.txt
```

### scp to copy files easily

The builtin `scp` function allows easy copying of files across hosts, using the persistent connections established with `gossh` instead of creating new connections as in the standard scp command.

`scp` is _always_ run from the local host, with the remote host filename specified as `@name:remotefile`

```sh
scp @name:hostfile.tsv localfile.tsv
```

Importantly, file wildcard globbing works as expected:
```sh
scp @name:*.tsv @0:data/
```

and entire directories can be copied, as in `cp -a` or `cp -r` (this behavior is automatic and does not require a flag).

### Close connections

```sh
gossh close
```

Will close all active connections and return the default host to @0.  All active connections are also automatically closed when the shell terminates.

## Other Utilties

** TODO: need a replacement for findnm -- very powerful but garbage..

## Rules for Go vs. Shell determination

These are the rules used to determine whether a line is Go vs. Shell (word = IDENT token):

* `$` at the start: Shell.
* Within Shell, `{}`: Go
* Within Go, `$ $`: Shell
* Line starts with `go` keyword: if no `( )` then Shell, else Go
* Line is one word: Shell
* Line starts with `path` expression (e.g., `./myexec`) : Shell
* Line starts with `"string"`: Shell
* Line starts with `word word`: Shell
* Line starts with `word {`: Shell
* Otherwise: Go

TODO: update above

## Multiple statements per line

* Multiple statements can be combined on one line, separated by `;` as in regular Go and shell languages.  Critically, the language determination for the first statement determines the language for the remaining statements; you cannot intermix the two on one line, when using `;` 

