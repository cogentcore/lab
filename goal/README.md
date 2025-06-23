# Goal: Go augmented language

Goal is an augmented version of the Go language, which combines the best parts of Go, `bash`, and Python, to provide and integrated shell and numerical expression processing experience, which can be combined with the [yaegi](https://github.com/traefik/yaegi) interpreter to provide an interactive "REPL" (read, evaluate, print loop).

See the [Cogent Lab Docs](https://cogentcore.org/lab/goal) for full documentation.

## Design discussion

Goal transpiles directly into Go, so it automatically leverages all the great features of Go, and remains fully compatible with it. The augmentation is designed to overcome some of the limitations of Go in specific domains:

* Shell scripting, where you want to be able to directly call other executable programs with arguments, without having to navigate all the complexity of the standard [os.exec](https://pkg.go.dev/os/exec) package.

* Numerical / math / data processing, where you want to be able to write simple mathematical expressions operating on vectors, matricies and other more powerful data types, without having to constantly worry about numerical type conversions, and advanced n-dimensional indexing and slicing expressions are critical. Python is the dominant language here precisely because it lets you ignore type information and write such expressions, using operator overloading.

* GPU-based parallel computation, which can greatly speed up some types of parallelizable computations by effectively running many instances of the same code in parallel across a large array of data.  The [gosl](gosl) package (automatically run in `goal build` mode) allows you to run the same Go-based code on a GPU or CPU (using parallel goroutines). See the [GPU](GPU.md) docs for an overview and comparison to other approaches to GPU computation.

The main goal of Goal is to achieve a "best of both worlds" solution that retains all the type safety and explicitness of Go for all the surrounding control flow and large-scale application logic, while also allowing for a more relaxed syntax in specific, well-defined domains where the Go language has been a barrier.  Thus, unlike Python where there are various weak attempts to try to encourage better coding habits, Goal retains in its Go foundation a fundamentally scalable, "industrial strength" language that has already proven its worth in countless real-world applications.

For the shell scripting aspect of Goal, the simple idea is that each line of code is either Go or shell commands, determined in a fairly intuitive way mostly by the content at the start of the line (formal rules below). If a line starts off with something like `ls -la...` then it is clear that it is not valid Go code, and it is therefore processed as a shell command.

You can intermix Go within a shell line by wrapping an expression with `{ }` braces, and a Go expression can contain shell code by using `$`.  Here's an example:
```go
for i, f := range goalib.SplitLines($ls -la$) {  // ls executes, returns string
    echo {i} {strings.ToLower(f)}              // {} surrounds Go within shell
}
```
where `goalib.SplitLines` is a function that runs `strings.Split(arg, "\n")`, defined in the `goalib` standard library of such frequently-used helper functions.

For cases where most of the code is standard Go with relatively infrequent use of shell expressions, or in the rare cases where the default interpretation doesn't work, you can explicitly tag a line as shell code using `$`:

```go
$ chmod +x *.goal
```

For mathematical expressions, we use `#` symbols (`#` = number) to demarcate such expressions. Often you will write entire lines of such expressions:
```go
# x := 1. / (1. + exp(-wts[:, :, :n] * acts[:]))
```
You can also intermix within Go code:
```go
for _, x := range #[1,2,3]# {
    fmt.Println(#x^2#)
}
```
Note that you cannot enter math mode directly from shell mode, which is unlikely to be useful anyway (you can wrap in go mode `{ }` if really needed).

In general, the math mode syntax in Goal is designed to be as compatible with Python NumPy / SciPy syntax as possible, while also adding a few Go-specific additions as well: see the [Math mode](#math-mode) section for details.  All elements of a Goal math expression are [tensors](../tensor), which can represent everything from a scalar to an n-dimenstional tensor.  These are called "ndarray" in NumPy terms.

The one special form of tensor processing that is available in regular Go code is _n dimensional indexing_, e.g., `tsr[1,2]`.  This kind of expression with square brackets `[ ]` and a comma is illegal according to standard Go syntax, so when we detect it, we know that it is being used on a tensor object, and can transpile it into the corresponding `tensor.Value` or `tensor.Set*` expression. This is particularly convenient for [gosl](gosl) GPU code that has special support for tensor data. Note that for this GPU use-case, you actually do _not_ want to use math mode, because that engages a different, more complex form of indexing that does _not_ work on the GPU.

The rationale and mnemonics for using `$` and `#` are as follows:

* These are two of the three common ASCII keyboard symbols that are not part of standard Go syntax (`@` being the other).

* `$` can be thought of as "S" in _S_hell, and is often used for a `bash` prompt, and many bash examples use it as a prefix. Furthermore, in bash, `$( )` is used to wrap shell expressions.

* `#` is commonly used to refer to numbers. It is also often used as a comment syntax, but on balance the number semantics and uniqueness relative to Go syntax outweigh that issue.

