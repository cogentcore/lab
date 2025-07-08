+++
Categories = ["Tensorfs"]
+++

**tensorfs** provides a virtual filesystem for [[tensor]] data, which can be accessed for example in [[Goal]] [[math]] mode expressions, like the variable storage system in [IPython / Jupyter](https://ipython.readthedocs.io/en/stable/interactive/tutorial.html), with the advantage that the hierarchical structure of a filesystem allows data to be organized in more intuitive and effective ways. For example, data at different time scales can be put into different directories, or multiple different statistics computed on a given set of data can be put into a subdirectory. [[stats#Groups]] creates pivot-table style groups of values as directories, for example.

`tensorfs` implements the Go [fs](https://pkg.go.dev/io/fs) interface, and can be accessed using fs-general tools, including the cogent core `filetree` and the [[Goal]] shell. 

There are two main APIs, one for direct usage within Go, and another that is used by the [[Goal]] framework for interactive shell-based access, which always operates relative to a current working directory.

## Go API

There are type-specific accessor methods for the standard high-frequency data types: `Float64`, `Float32`, `Int`, and `StringValue` (`String` is taken by the stringer interface):

```Goal
dir, _ := tensorfs.NewDir("root")
x := dir.Float64("data", 3, 3)

fmt.Println(dir.ListLong(true, 2))
fmt.Println(x)
```

Which are wrappers around the underlying Generic `Value` method:

```go
x := tensorfs.Value[float64](dir, "data", 3, 3)
```

These methods create the given tensor if it does not yet exist, and otherwise return it, providing a robust order-independent way of accessing / constructing the relevant data.

For efficiency, _there are no checks_ on the existing value relative to the arguments passed, so if you end up using the same name for two different things, that will cause problems that will hopefully become evident. If you want to ensure that the size is correct, you should use an explicit `tensor.SetShapeSizes` call, which is still quite efficient if the size is the same. You can also have an initial call to `Value` that has no size args, and then set the size later -- that works fine.

There are also a few other variants of the `Value` functionality:
* `Scalar` calls `Value` with a size of 1.
* `Values` makes multiple tensor values of the same shape, with a final variadic list of names.
* `ValueType` takes a `reflect.Kind` arg for the data type, which can then be a variable.
* `SetTensor` sets a tensor to a node of given name, creating the node if needed. This is also available as the `Set` method on a directory node.

## DirTable and tar files

[[doc:tensorfs.DirTable]] returns a [[table]] with all the tensors under a given directory node, which can then be used for making plots or doing other forms of data analysis. This works best when each tensor has the same outer-most row dimension. The table is persistent and very efficient, using direct pointers to the underlying tensor values.

Use [[doc:tensorfs.DirFromTable]] to set the contents of a directory from a table. This will also use any slashes in column names to recreate the hierarchical structure of directories and subdirectories, but note that the `DirTable` command only uses the last two levels of the path name for naming columns (i.e., the leaf name and its immediate parent).

Use [[doc:tensorfs.Tar]] and [[doc:tensorfs.Untar]] if you want to save and reload a full directory structure in an efficient manner (also doesn't depend on row alignment).

## Directories

A given [[doc:tensorfs.Node]] can either have a [[tensor]] value or be a _subdirectory_ containing a list of other node lements.

To make a new subdirectory:

```Goal
dir, _ := tensorfs.NewDir("root")
subdir := dir.Dir("sub")
x := subdir.Float64("data", 3, 3)

fmt.Println(dir.ListLong(true, 2))
fmt.Println(x)
```

If the subdirectory doesn't exist yet, it will be made, and otherwise it is returned. Any errors will be logged and a nil returned, likely causing a panic unless you expect it to fail and check for that.

## Operating over values across directories

The `ValuesFunc` method on a directory node allows you to easily extract a list of values across any number of subdirectories (it only returns the final value "leaves" of the filetree):

```Goal
dir, _ := tensorfs.NewDir("root")
subdir := dir.Dir("sub")
x := subdir.Float64("x", 3, 3)
subsub := subdir.Dir("stats")
y := subsub.Float64("y", 1)
z := subsub.Float64("z", 1)

fmt.Println(dir.ListLong(true, 2))

vals := dir.ValuesFunc(nil) // nil = get everything
for _, v := range vals {
	fmt.Println(v)
}
```

Thus, even if you have statistics or other data nested down deep, this will "flatten" the hierarchy and allow you to process it. Here's a version that actually filters the nodes:

```Goal
dir, _ := tensorfs.NewDir("root")
subdir := dir.Dir("sub")
x := subdir.Float64("x", 5, 5)
subsub := subdir.Dir("stats")
y := subsub.Float64("y", 1)
z := subsub.Float64("z", 1)

fmt.Println(dir.ListLong(true, 2))

vals := dir.ValuesFunc(func(n *tensorfs.Node) bool {
    if n.IsDir() { // can filter by dirs here too (get to see everything)
        return true
    }
    return n.Tensor.NumDims() == 1
})
for _, v := range vals {
	fmt.Println(v)
}
```

There are parallel `Node` and `Value` access methods for directory nodes, with the Value ones being:

* `tsr := dir.Value("name")` returns tensor directly, will panic if not valid
* `tsrs, err := dir.Values("name1", "name2")` returns a slice of tensor values within directory by name. a plain `.Values()` returns all values.
* `tsrs := dir.ValuesFunc(<filter func>)` walks down directories (unless filtered) and returns a flat list of all tensors found. Goes in "directory order" = order nodes were added.
* `tsrs := dir.ValuesAlphaFunc(<filter func>)` is like `ValuesFunc` but traverses in alpha order at each node.

## 
