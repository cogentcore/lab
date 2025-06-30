# tensorfs: a virtual filesystem for tensor data

`tensorfs` is a virtual file system that implements the Go `fs` interface, and can be accessed using fs-general tools, including the cogent core `filetree` and the `goal` shell.

See the [Cogent Lab Docs](https://cogentcore.org/lab/tensorfs) for full documentation.

## Design discussion

Values are represented using the [tensor] package universal data type: the `tensor.Tensor`, which can represent everything from a single scalar value up to n-dimensional collections of patterns, in a range of data types.

A given `Node` in the file system is either:
* A _Value_, with a tensor encoding its value. These are terminal "leaves" in the hierarchical data tree, equivalent to "files" in a standard filesystem.
* A _Directory_, with an ordered map of other Node nodes under it.

Each Node has a name which must be unique within the directory. The nodes in a directory are processed in the order of its ordered map list, which initially reflects the order added, and can be re-ordered as needed. An alphabetical sort is also available with the `Alpha` versions of methods, and is the default sort for standard FS operations.

The uniqueness constraint of names within each directory creates issues for how to manage conflicts between existing and new items. To make the overall framework maximally robust and eliminate the need for a controlled initialization-then-access ordering, we generally adopt the "Recycle" logic:

* _Return an existing item of the same name, or make a new one._

In addition, if you really need to know if there is an existing item, you can use the `Node` method to check for yourself -- it will return `nil` if no node of that name exists. Furthermore, the global `NewDir` function returns an `fs.ErrExist` error for existing items (e.g., use `errors.Is(fs.ErrExist)`), as used in various `os` package functions.


