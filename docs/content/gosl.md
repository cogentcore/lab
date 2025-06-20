+++
Categories = ["Gosl"]
+++

**Gosl** allows you to write Go programs that run on [[GPU]] hardware, by transpiling Go into the WGSL shader language used by [WebGPU](https://www.w3.org/TR/webgpu/), thereby establishing the _Go shader language_.

Gosl uses the [core gpu](https://github.com/cogentcore/core/tree/main/gpu) compute shader system, and operates within the overall [[Goal]] framework of an augmented version of the Go language.

The relevant regions of Go code to be run on the GPU are tagged using the `//gosl:start` and `//gosl:end` comment directives, and this code must only use basic expressions and concrete types that will compile correctly in a GPU shader (see [[#Restrictions]] below). Method functions and pass-by-reference pointer arguments to `struct` types are supported and incur no additional compute cost due to inlining (see notes below for more detail).

See [[doc:examples/basic]] and [[doc:examples/rand]] for complete working examples.

Typically, `gosl` is called from a `go generate` command, e.g., by including this comment directive:

```
//go:generate gosl 
```

To install the `gosl` command:
```bash
$ go install cogentcore.org/lab/gosl/@latest
```

It is also strongly recommended to install the `naga` WGSL compiler from [wgpu](https://github.com/gfx-rs/wgpu) and the `tint` compiler from [dawn](https://dawn.googlesource.com/dawn/) Both of these are used if available to validate the generated GPU shader code. It is much faster to fix the issues at generation time rather than when trying to run the app later. Once code passes validation in both of these compilers, it should load fine in your app, and if the Go version runs correctly, there is a good chance of at least some reasonable behavior on the GPU.

## Usage

There are two key elements for GPU-enabled code:

1. One or more [[#Kernels]] compute functions that take an _index_ argument and perform computations for that specific index of data, _in parallel_. **GPU computation is effectively just a parallel `for` loop**. On the GPU, each such kernel is implemented by its own separate compute shader code, and one of the main functions of `gosl` is to generate this code from the Go sources, in the automatically created `shaders/` directory.

2. [[#Global variables]] on which the kernel functions _exclusively_ operate: all relevant data must be specifically copied from the CPU to the GPU and back. As explained in the [[GPU]] docs, each GPU compute shader is effectively a _standalone_ program operating on these global variables. To replicate this environment on the CPU, so the code works in both contexts, we need to make these variables global in the CPU (Go) environment as well.

`gosl` generates a file named `gosl.go` in your package directory that initializes the GPU with all of the global variables, and functions for running the kernels and syncing the gobal variable data back and forth between the CPu and GPU.

## Kernels

Each distinct compute kernel must be tagged with a `//gosl:kernel` comment directive, as in this example (from `examples/basic`):
```go
// Compute does the main computation.
func Compute(i uint32) { //gosl:kernel
	Params[0].IntegFromRaw(int(i))
}
```

The kernel functions receive a `uint32` index argument, and use this to index into the global variables containing the relevant data. Typically the kernel code itself just calls other relevant function(s) using the index, as in the above example. Critically, _all_ of the data that a kernel function ultimately depends on must be contained with the global variables, and these variables must have been sync'd up to the GPU from the CPU prior to running the kernel (more on this below).

In the CPU mode, the kernel is effectively run in a `for` loop like this:
```go
	for i := range n {
		Compute(uint32(i))
	}
```
A parallel goroutine-based mechanism is actually used, but conceptually this is what it does, on both the CPU and the GPU. To reiterate: **GPU computation is effectively just a parallel for loop**.

## Global variables

The global variables on which the kernels operate are declared in the usual Go manner, as a single `var` block, which is marked at the top using the `//gosl:vars` comment directive:

```go
//gosl:vars
var (
	// Params are the parameters for the computation.
	//gosl:read-only
	Params []ParamStruct

	// Data is the data on which the computation operates.
	// 2D: outer index is data, inner index is: Raw, Integ, Exp vars.
	//gosl:dims 2
	Data *tensor.Float32
)
```

All such variables must be either:
1. A `slice` of GPU-alignment compatible `struct` types, such as `ParamStruct` in the above example. In general such structs should be marked as `//gosl:read-only` due to various challenges associated with writing to structs, detailed below.
2. A `tensor` of a GPU-compatible elemental data type (`float32`, `uint32`, or `int32`), with the number of dimensions indicated by the `//gosl:dims <n>` tag as shown above. This is the preferred type for writable data.

You can also just declare a slice of elemental GPU-compatible data values such as `float32`, but it is generally preferable to use the tensor instead, because it has built-in support for higher-dimensional indexing in a way that is transparent between CPU and GPU.

### Tensor data

On the GPU, the tensor data is represented using a simple flat array of the basic data type. To index into this array, the _strides_ for each dimension are encoded in a special `TensorStrides` tensor that is managed by `gosl`, in the generated `gosl.go` file. `gosl` automatically generates the appropriate indexing code using these strides (which is why the number of dimensions is needed).

Whenever the strides of any tensor variable change, and at least once at initialization, your code must call the function that copies the current strides up to the GPU:
```go
	ToGPUTensorStrides()
```

### Multiple tensor variables for large data

The size of each memory buffer is limited by the GPU, to a maximum of at most 4GB on modern GPU hardware. Therefore, if you need to have any single tensor that holds more than this amount of data, then a bank of multiple vars are required. `gosl` provides helper functions to make this relatively straightforward.

TODO: this could be encoded in the TensorStrides. It will always be the outer-most index that determines when it gets over threshold, which all can be pre-computed.

### Systems and Groups

Each kernel belongs to a `gpu.ComputeSystem`, and each such system has one specific configuration of memory variables. In general, it is best to use a single set of global variables, and perform as much of the computation as possible on this set of variables, to minimize the number of memory transfers. However, if necessary, multiple systems can be defined, using an optional additional system name argument for the `args` and `kernel` tags.

In addition, the vars can be organized into _groups_, which generally should have similar memory syncing behavior, as documented in the [core gpu](https://github.com/cogentcore/core/tree/main/gpu) system.

Here's an example with multiple groups:
```go
//gosl:vars [system name]
var (
    // Layer-level parameters
    //gosl:group -uniform Params
    Layers   []LayerParam // note: struct with appropriate memory alignment

    // Path-level parameters
    Paths    []PathParam  

    // Unit state values
    //gosl:group Units
    Units    tensor.Float32
    
    // Synapse weight state values
    Weights  tensor.Float32
)
```

## Memory syncing

Each global variable gets an automatically-generated `*Var` enum (e.g., `DataVar` for global variable named `Data`), that used for the memory syncing functions, to make it easy to specify any number of such variables to sync, which is by far the most efficient. All of this is in the generated `gosl.go` file. For example:

```go
	ToGPU(ParamsVar, DataVar)
```

Specifies that the current contents of `Params` and `Data` are to be copied up to the GPU, which is guaranteed to complete by the time the next kernel run starts, within a given system.

## Kernel running

As with memory transfers, it is much more efficient to run multiple kernels in sequence, all operating on the current data variables, followed by a single sync of the updated global variable data that has been computed. Thus, there are separate functions for specifying the kernels to run, followed by a single "Done" function that actually submits the entire batch of kernels, along with memory sync commands to get the data back from the GPU. For example:

```go
    RunCompute1(n)
    RunCompute2(n)
    ...
    RunDone(Data1Var, Data2Var) // launch all kernels and get data back to given vars
```

For CPU mode, `RunDone` is a no-op, and it just runs each kernel during each `Run` command.

It is absolutely essential to understand that _all data must already be on the GPU_ at the start of the first Run command, and that any CPU-based computation between these calls is completely irrelevant for the GPU. Thus, it typically makes sense to just have a sequence of Run commands grouped together into a logical unit, with the relevant `ToGPU` calls at the start and the final `RunDone` grabs everything of relevance back from the GPU.

## GPU relevant code taggng

In a large GPU-based application, you should organize your code as you normally would in any standard Go application, distributing it across different files and packages. The GPU-relevant parts of each of those files can be tagged with the gosl tags:
```
//gosl:start

< Go code to be translated >

//gosl:end
```
to make this code available to all of the shaders that are generated.

Use the `//gosl:import "package/path"` directive to import GPU-relevant code from other packages, similar to the standard Go import directive. It is assumed that many other Go imports are not GPU relevant, so this separate directive is required.

If any `enums` variables are defined, pass the `-gosl` flag to the `core generate` command to ensure that the `N` value is tagged with `//gosl:start` and `//gosl:end` tags.

**IMPORTANT:** all `.go` and `.wgsl` files are removed from the `shaders` directory prior to processing to ensure everything there is current -- always specify a different source location for any custom `.wgsl` files that are included.

# Command line usage

```
gosl [flags] 
```
    
The flags are:
```
  -debug
    	enable debugging messages while running
  -exclude string
    	comma-separated list of names of functions to exclude from exporting to WGSL (default "Update,Defaults")
  -keep
    	keep temporary converted versions of the source files, for debugging
  -out string
    	output directory for shader code, relative to where gosl is invoked -- must not be an empty string (default "shaders")
```

`gosl` always operates on the current directory, looking for all files with `//gosl:` tags, and accumulating all the `import` files that they include, etc.
  
Any `struct` types encountered will be checked for 16-byte alignment of sub-types and overall sizes as an even multiple of 16 bytes (4 `float32` or `int32` values), which is the alignment used in WGSL and glsl shader languages, and the underlying GPU hardware presumably.  Look for error messages on the output from the gosl run.  This ensures that direct byte-wise copies of data between CPU and GPU will be successful.  The fact that `gosl` operates directly on the original CPU-side Go code uniquely enables it to perform these alignment checks, which are otherwise a major source of difficult-to-diagnose bugs.

# Restrictions    

In general shader code should be simple mathematical expressions and data types, with minimal control logic via `if`, `for` statements, and only using the subset of Go that is consistent with C.  Here are specific restrictions:

* Can only use `float32`, `[u]int32` for basic types (`int` is converted to `int32` automatically), and `struct` types composed of these same types -- no other Go types (i.e., `map`, slices, `string`, etc) are compatible.  There are strict alignment restrictions on 16 byte (e.g., 4 `float32`'s) intervals that are enforced via the `alignsl` sub-package.

* WGSL does _not_ support 64 bit float or int.

* Use `slbool.Bool` instead of `bool` -- it defines a Go-friendly interface based on a `int32` basic type.

* Alignment and padding of `struct` fields is key -- this is automatically checked by `gosl`.

* WGSL does not support enum types, but standard go `const` declarations will be converted.  Use an `int32` or `uint32` data type.  It will automatically deal with the simple incrementing `iota` values, but not more complex cases.  Also, for bitflags, define explicitly, not using `bitflags` package, and use `0x01`, `0x02`, `0x04` etc instead of `1<<2` -- in theory the latter should be ok but in practice it complains.

* Cannot use multiple return values, or multiple assignment of variables in a single `=` expression.

* *Can* use multiple variable names with the same type (e.g., `min, max float32`) -- this will be properly converted to the more redundant form with the type repeated, for WGSL.

* `switch` `case` statements are _purely_ self-contained -- no `fallthrough` allowed!  does support multiple items per `case` however. Every `switch` _must_ have a `default` case.

* WGSL does specify that new variables are initialized to 0, like Go, but also somehow discourages that use-case.  It is safer to initialize directly:
```go
    val := float32(0) // guaranteed 0 value
    var val float32 // ok but generally avoid
```    

* Use the automatically-generated `GetX` methods to get a local variable to a slice of structs:
```go
    ctx := GetCtx(0)
```
This automatically does the right thing on GPU while returning a pointer to the indexed struct on CPU.

* tensor variables can only be used in `storage` (not `uniform`) memory, due to restrictions on dynamic sizing and alignment. Aside from this constraint, it is possible to designate a group of variables to use uniform memory, with the `-uniform` argument as the first item in the `//gosl:group` comment directive.

## Other language features

* [tour-of-wgsl](https://google.github.io/tour-of-wgsl/types/pointers/passing_pointers/) is a good reference to explain things more directly than the spec.

* `ptr<function,MyStruct>` provides a pointer arg
* `private` scope = within the shader code "module", i.e., one thread.  
* `function` = within the function, not outside it.
* `workgroup` = shared across workgroup -- coudl be powerful (but slow!) -- need to learn more.

## Atomic access

WGSL adopts the Metal (lowest common denominator) strong constraint of imposing a _type_ level restriction on atomic operations: you can only do atomic operations on variables that have been declared atomic, as in:

```
var<storage, read_write> PathGBuf: array<atomic<i32>>;
...
atomicAdd(&PathGBuf[idx], val);
```

This also unfortunately has the side-effect that you cannot do _non-atomic_ operations on atomic variables, as discussed extensively here: https://github.com/gpuweb/gpuweb/issues/2377  Gosl automatically detects the use of atomic functions on GPU variables, and tags them as atomic. 

## Random numbers: slrand

See [[doc:gosl/slrand]] for a shader-optimized random number generation package, which is supported by `gosl` -- it will convert `slrand` calls into appropriate WGSL named function calls.  `gosl` will also copy the `slrand.wgsl` file, which contains the full source code for the RNG, into the destination `shaders` directory, so it can be included with a simple local path:

```go
//gosl:wgsl mycode
// #include "slrand.wgsl"
//gosl:end mycode
```

## Performance

With sufficiently large N, and ignoring the data copying setup time, around ~80x speedup is typical on a Macbook Pro with M1 processor.  The `rand` example produces a 175x speedup!


