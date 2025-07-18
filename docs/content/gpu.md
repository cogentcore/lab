+++
Categories = ["Gosl"]
Title = "GPU"
Name = "GPU"
+++

The use of massively parallel _Graphical Processsing Unit_ (**GPU**) hardware has revolutionized machine learning and other fields, producing many factors of speedup relative to traditional _CPU_ (_Central Processing Unit_) computation. However, there are numerous challenges for supporting GPU-based computation, relative to the more flexible CPU coding.

The [[Gosl]] Go shader language operating within the broader [[Goal]] augmented version of the Go lanuage provides a solution to these challenges that enables the same Go-based code to work efficiently and reasonably naturally on both the GPU and CPU (i.e., standard Go execution).

Debugging code on the GPU is notoriously difficult because the usual tools are not directly available (not even print statements), so the ability to run exactly the same code on the CPU and GPU is invaluable, in addition to the benefits in portability across platforms without GPU hardware.

See the [[gosl]] documentation for the details on how to write code that works on the GPU. The remainder of this document provides an overview of the overall approach in relation to other related tools.

## Challenges for GPU computation

The two most important challenges for GPU-based programs are:

* The GPU _has its own separate memory space_ that needs to be synchronized explicitly and bidirectionally with the standard CPU memory (this is true programmatically even if at a hardware level there is shared memory).

* Computation must be organized into discrete chunks that can be computed efficiently in parallel, and each such chunk of computation lives in its own separate _kernel_ (_compute shader_) in the GPU, as an entirely separate, self-contained program, operating on _global variables_ that define the entire memory space of the computation.

To be maximally efficient, both of these factors must be optimized, such that:

* The bidirectional syncing of memory between CPU and GPU should be minimized, because such transfers incur a significant overhead.

* The overall computation should be broken down into the _largest possible chunks_ to minimize the number of discrete kernel runs, each of which incurs significant overhead.

Thus, it is unfortunately _highly inefficient_ to implement GPU-based computation by running each elemental vectorizable tensor operation (add, multiply, etc) as a separate GPU kernel, with its own separate bidirectional memory sync, even though that is a conceptually attractive and simple way to organize GPU computation, with minimal disruption relative to the CPU model.

The [JAX](https://github.com/jax-ml/jax) framework in Python provides one solution to this situation, optimized for neural network machine learning uses, by imposing strict _functional programming_ constraints on the code you write (i.e., all functions must be _read-only_), and leveraging those to automatically combine elemental computations into larger parallelizable chunks, using a "just in time" (_jit_) compiler.

We take a different approach, which is much simpler implementationally but requires a bit more work from the developer, which is to provide tools that allow _you_ to organize your computation into kernel-sized chunks according to your knowledge of the problem, and transparently turn that code into the final CPU and GPU programs.

In many cases, a human programmer can most likely out-perform the automatic compilation process, by knowing the full scope of what needs to be computed, and figuring out how to package it most efficiently per the above constraints. In the end, you get maximum efficiency and complete transparency about exactly what is being computed, perhaps with fewer "gotcha" bugs arising from all the magic happening under the hood, but it may take a bit more work to get there.

The role of [[Gosl]] and [[Goal]] is to allow you to express the full computation in the clear, simple, Go language, using intuitive data structures that minimize the need for additional boilerplate to run efficiently on CPU and GPU. This ability to write a single codebase that runs efficiently on CPU and GPU is similar to the [SYCL](https://en.wikipedia.org/wiki/SYCL) framework (and several others discussed on that wikipedia page), which builds on [OpenCL](https://en.wikipedia.org/wiki/OpenCL), both of which are based on the C / C++ programming language.

In addition to the critical differences between Go and C++ as languages, Gosl targets only one hardware platform: WebGPU (via the [core gpu](https://github.com/cogentcore/core/tree/main/gpu) package), so it is more specifically optimized for this use-case. Furthermore, SYCL and other approaches require you to write GPU-like code that can also run on the CPU (with lots of explicit fine-grained memory and compute management), whereas Goal provides a more natural CPU-like programming model, while imposing some stronger constraints that encourage more efficient implementations.

The bottom line is that the fantasy of being able to write CPU-native code and have it magically "just work" on the GPU with high levels of efficiency is just that: a fantasy. The reality is that code must be specifically structured and organized to work efficiently on the GPU. Goal just makes this process relatively clean and efficient and easy to read, with a minimum of extra boilerplate. The resulting code should be easily understood by anyone familiar with the Go language, even if that isn't the way you would have written it in the first place. The reward is that you can get highly efficient results with significant GPU-accelerated speedups that works on _any platform_, including the web and mobile phones, all with a single easy-to-read codebase.

## Kernel functions

First, we assume the scope is a single Go package that implements a set of computations on some number of associated data representations. The package will likely contain a lot of CPU-only Go code that manages all the surrounding infrastructure for the computations, in terms of creating and configuring the data in memory, visualization, i/o, etc.

The GPU-specific computation is organized into some (hopefully small) number of **kernel** functions, that are conceptually called using a **parallel for loop**, e.g., something like this:

```go
for i := range parallel(data) {
    Compute(i)
}
```

The `i` index effectively iterates over the range of the values of the `data` variable, with the GPU version launching kernels on the GPU for each different index value. The CPU version actually runs in parallel as well, using goroutines.

We assume that multiple kernels will in general be required, and that there is likely to be a significant amount of shared code infrastructure across these kernels. Thus, the kernel functions are typically relatively short, and call into a large body of code that is likely shared among the different kernel functions.

Even though the GPU kernels must each be compiled separately into a single distinct WGSL _shader_ file that is run under WebGPU, they can `import` a shared codebase of files, and thus replicate the same overall shared code structure as the CPU versions.

The GPU code can only handle a highly restricted _subset_ of Go code, with data structures having strict alignment requirements, and no `string` or other composite variable-length data structures (maps, slices etc). Thus, [[Gosl]] recognizes `//gosl:start` and `//gosl:end` comment directives surrounding the GPU-safe (and relevant) portions of the overall code. Any `.go` or `.goal` file can contribute GPU relevant code, including in other packages, and the gosl system automatically builds a shadow package-based set of `.wgsl` files accordingly.

> Each kernel function is marked with a `//gosl:kernel` directive, and the name of the function is used to create the name of the GPU shader file.

```go
// Compute does the main computation.
func Compute(i uint32) { //gosl:kernel
	Params[0].IntegFromRaw(&Data[i])
}
```

## Memory Organization

Perhaps the strongest constraints for GPU programming stem from the need to organize and synchronize all the memory buffers holding the data that the GPU kernel operates on. Furthermore, within a GPU kernel, the variables representing this data are _global variables_, which is sensible given the standalone nature of each kernel.

> To provide a common programming environment, all GPU-relevant variables must be Go global variables.

Thus, names must be chosen appropriately for these variables, given their global scope within the Go package. The specific _values_ for these variables can be dynamically set in an easy way, but the variables themselves are global.

Within the [core gpu](https://github.com/cogentcore/core/tree/main/gpu) framework, each `ComputeSystem` defines a specific organization of such GPU buffer variables, and maximum efficiency is achieved by minimizing the number of such compute systems, and associated memory buffers. Each system also encapsulates the associated kernel shaders that operate on the associated memory data, so

> Kernels and variables both must be defined within a specific system context.

### tensorfs mapping

TODO:

The grouped global variables can be mapped directly to a corresponding [tensorfs](../tensor/tensorfs) directory, which provides direct accessibility to this data within interactive Goal usage. Further, different sets of variable values can be easily managed by saving and loading different such directories.

```go
    gosl.ToDataFS("path/to/dir" [, system]) // set tensorfs items in given path to current global vars
    
    gosl.FromDataFS("path/to/dir" [,system]) // set global vars from given tensorfs path
```

These and all such `gosl` functions use the current system if none is explicitly specified, which is settable using the `gosl.SetSystem` call. Any given variable can use the `get` or `set` Goal math mode functions directly.

## Memory access

In general, all global GPU variables will be arrays (slices) or tensors, which are exposed to the GPU as an array of floats.

The tensor-based indexing syntax in Goal math mode transparently works across CPU and GPU modes, and is thus the preferred way to access tensor data.

It is critical to appreciate that none of the other convenient math-mode operations will work as you expect on the GPU, because:

> There is only one outer-loop, kernel-level parallel looping operation allowed at a time.

You cannot nest multiple such loops within each other. A kernel cannot launch another kernel. Therefore, as noted above, you must directly organize your computation to maximize the amount of parallel computation happening wthin each such kernel call.

> Therefore, tensor indexing on the GPU only supports direct index values, not ranges.

Furthermore:

> Pointer-based access of global variables is not supported in GPU mode.

You have to use _indexes_ into arrays exclusively. Thus, some of the data structures you may need to copy up to the GPU include index variables that determine how to access other variables.

## Examples

A large and complex biologically-based neural network simulation framework called [axon](https://github.com/emer/axon) has been implemented using `gosl`, allowing 1000's of lines of equations and data structures to run through standard Go on the CPU, and accelerated significantly on the GPU. This allows efficient debugging and unit testing of the code in Go, whereas debugging on the GPU is notoriously difficult.


