# Implementational details of Go to SL translation process

Overall, there are three main steps:

1. Translate all the `.go` files in the current package, and all the files they `//gosl:import`, into corresponding `.wgsl` files, and put those in `shaders/imports`.  All these files will be pasted into the generated primary kernel files, that go in `shaders`, and are saved to disk for reference. All the key kernel, system, variable info is extracted from the package .go file directives during this phase.

2. Generate the `main` kernel `.wgsl` files, for each kernel function, which: a) declare the global buffer variables; b) include everything from imports; c) define the `main` function entry point. Each resulting file is pre-processed by `naga` to ensure it compiles, and to remove dead code not needed for this particular shader.

3. Generate the `gosl.go` file in the package directory, which contains generated Go code for configuring the gpu compute systems according to the vars.

## Go to SL translation

1. `files.go`: Get a list of all the .go files in the current directory that have a `//gosl:` tag (`ProjectFiles`) and all the `//gosl:import` package files that those files import, recursively.

2. `extract.go`: Extract the `//gosl:start` -> `end` regions from all the package and imported filees.

3. Save all these files as new `.go` files in `shaders/imports`. We manually append a simple go "main" package header with basic gosl imports for each file, which allows the go compiler to process them properly. This is then removed in the next step.

4. `translate.go:` Run `TranslateDir` on shaders/imports using the "golang.org/x/tools/go/packages" `Load` function, which gets `ast` and type information for all that code. Run the resulting `ast.File` for each file through the modified version of the Go stdlib `src/go/printer` code (`printer.go`, `nodes.go`, `gobuild.go`, `comment.go`), which prints out WGSL instead of Go code from the underlying `ast` representation of the Go files. This is what does the actual translation.

5. `sledits.go:` Do various forms of post-processing text replacement cleanup on the generated WGSL files, in `SLEdits` function. 

# Struct types and read-write vs. read-only:

1. In Go, all struct args are generally pointers, including method receivers.

2. In WGSL, pointers are strongly constrained, and you cannot use a pointer to a struct field within another struct. In general, structs are best used for static parameters, rather than writable data.

3. For a given method, it is not possible to know in advance based just on type data whether a given struct arg is read-only or read-write -- these are properties of the variables, not the types.

4. Therefore, we make the strong simplifying assumption that all struct args to a method are read-only, and thus the GPU code sets them to non-pointers. Any code that passes non-read-only struct args to such methods will generate a WGSL compilation / validation error.

5. There is one exception: a given method can be explicitly marked with `//gosl:pointer-receiver` to make the method receiver argument a pointer, and thus usable with read-write pointers.

