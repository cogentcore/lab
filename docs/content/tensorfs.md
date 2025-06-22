+++
+++

**tensorfs** provides a virtual filesystem for [[tensor]] data, which can be accessed for example in [[Goal]] [[math]] mode expressions, like the variable storage system in [IPython / Jupyter](https://ipython.readthedocs.io/en/stable/interactive/tutorial.html), with the advantage that the hierarchical structure of a filesystem allows data to be organized in more intuitive and effective ways. For example, data at different time scales can be put into different directories, or multiple different statistics computed on a given set of data can be put into a subdirectory. [[stats#Groups]] creates pivot-table style groups of values as directories, for example.

`tensorfs` implements the Go [fs](https://pkg.go.dev/io/fs) interface, and can be accessed using fs-general tools, including the cogent core `filetree` and the [[Goal]] shell. 

There are two main APIs, one for direct usage within Go, and another that is used by the [[Goal]] framework for interactive shell-based access, which always operates relative to a current working directory.

## Go API

There are type-specific accessor methods for the standard high-frequency data types: `Float64`, `Float32`, `Int`, and `StringValue` (`String` is taken by the stringer interface):

```Goal
dir, _ := tensorfs.NewDir("root")
x := dir.Float64("data", 5, 5)

fmt.Println(dir.ListLong(true, 2))
fmt.Println(x.String())
```

Which are wrappers around the underlying Generic `Value` method:

```go
x := tensorfs.Value[float64](dir, "data", 5, 5)
```

These methods create the given tensor if it does not yet exist, and otherwise return it, providing a robust order-independent way of accessing / constructing the relevant data.

For efficiency, _there are no checks_ on the existing value relative to the arguments passed, so if you end up using the same name for two different things, that will cause problems that will hopefully become evident. If you want to ensure that the size is correct, you should use an explicit `tensor.SetShapeSizes` call, which is still quite efficient if the size is the same. You can also have an initial call to `Value` that has no size args, and then set the size later -- that works fine.

There are also a few other variants of the `Value` functionality:
* `Scalar` calls `Value` with a size of 1.
* `Values` makes multiple tensor values of the same shape, with a final variadic list of names.
* `ValueType` takes a `reflect.Kind` arg for the data type, which can then be a variable.
* `SetTensor` sets a tensor to a node of given name, creating the node if needed. This is also available as the `Set` method on a directory node.

`DirTable` returns a [[table]] with all the tensors under a given directory node, which can then be used for making plots or doing other forms of data analysis. This works best when each tensor has the same outer-most row dimension. The table is persistent and very efficient, using direct pointers to the underlying tensor values.

## Directories

A given [[doc:tensorfs.Node]] can either have a [[tensor]] value or be a _subdirectory_ containing a list of other node lements.

To make a new subdirectory:

```Goal
dir, _ := tensorfs.NewDir("root")
subdir := dir.Dir("sub")
x := subdir.Float64("data", 5, 5)

fmt.Println(dir.ListLong(true, 4))
fmt.Println(x.String())
```

If the subdirectory doesn't exist yet, it will be made, and otherwise it is returned. Any errors will be logged and a nil returned, likely causing a panic unless you expect it to fail and check for that.

There are parallel `Node` and `Value` access methods for directory nodes, with the Value ones being:

* `tsr := dir.Value("name")` returns tensor directly, will panic if not valid
* `tsrs, err := dir.Values("name1", "name2")` returns a slice of tensor values within directory by name. a plain `.Values()` returns all values.
* `tsrs := dir.ValuesFunc(<filter func>)` walks down directories (unless filtered) and returns a flat list of all tensors found. Goes in "directory order" = order nodes were added.
* `tsrs := dir.ValuesAlphaFunc(<filter func>)` is like `ValuesFunc` but traverses in alpha order at each node.

## Existing items and unique names

As in a real filesystem, names must be unique within each directory, which creates issues for how to manage conflicts between existing and new items. To make the overall framework maximally robust and eliminate the need for a controlled initialization-then-access ordering, we generally adopt the "Recycle" logic:

* _Return an existing item of the same name, or make a new one._

In addition, if you really need to know if there is an existing item, you can use the `Node` method to check for yourself -- it will return `nil` if no node of that name exists. Furthermore, the global `NewDir` function returns an `fs.ErrExist` error for existing items (e.g., use `errors.Is(fs.ErrExist)`), as used in various `os` package functions.

## Goal Command API

The following shell command style functions always operate relative to the global `CurDir` current directory and `CurRoot` root, and `goal` in [[math]] mode exposes these methods directly (see [[math#Tensorfs]]).

* `Chdir("subdir")` change current directory to subdir.
* `Mkdir("subdir")` make a new directory.
* `List()` print a list of nodes.
* `tsr := Get("mydata")` get tensor value at "mydata" node.
* `Set("mydata", tsr)` set tensor to "mydata" node (node points to tensor, is updated when tensor is).
* `SetCopy("mydata", tsr)` set tensor to a copy "mydata" node (data is preserved from time of call).

A given `Node` in the file system is either:
* A _Value_, with a tensor encoding its value. These are terminal "leaves" in the hierarchical data tree, equivalent to "files" in a standard filesystem.
* A _Directory_, with an ordered map of other Node nodes under it.

Each Node has a name which must be unique within the directory. The nodes in a directory are processed in the order of its ordered map list, which initially reflects the order added, and can be re-ordered as needed. An alphabetical sort is also available with the `Alpha` versions of methods, and is the default sort for standard FS operations.


