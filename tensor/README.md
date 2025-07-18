# Tensor

Tensor and related sub-packages provide a simple yet powerful framework for representing n-dimensional data of various types, providing similar functionality to the widely used [NumPy](https://numpy.org/doc/stable/index.html) and [pandas](https://pandas.pydata.org/) libraries in Python, and the commercial MATLAB framework.

The [Goal](../goal) augmented version of the _Go_ language directly supports NumPy-like operations on tensors. A `Tensor` is comparable to the NumPy `ndarray` type, and it provides the universal representation of a homogenous data type throughout all the packages here, from scalar to vector, matrix and beyond. All functions take and return `Tensor` arguments.

See the [Cogent Lab Docs](https://cogentcore.org/lab/tensor) for full documentation.

## Design discussion

The `Tensor` interface is implemented at the basic level with n-dimensional indexing into flat Go slices of any numeric data type (by `Number`), along with `String`, and `Bool` (which uses [bitslice](bitslice) for maximum efficiency). These implementations satisfy the `Values` sub-interface of Tensor, which supports the most direct and efficient operations on contiguous memory data. The `Shape` type provides all the n-dimensional indexing with arbitrary strides to allow any ordering, although _row major_ is the default and other orders have to be manually imposed.

In addition, there are five important "view" implementations of `Tensor` that wrap another "source" Tensor to provide more flexible and efficient access to the data, consistent with the NumPy functionality.  See [Basic and Advanced Indexing](#basic-and-advanced-indexing) below for more info.

* `Sliced` provides a sub-sliced view into the wrapped `Tensor` source, using an indexed list along each dimension. Thus, it can provide a reordered and filtered view onto the raw data, and it has a well-defined shape in terms of the number of indexes per dimension. This corresponds to the NumPy basic sliced indexing model.

* `Masked` provides a `Bool` masked view onto each element in the wrapped `Tensor`, where the two maintain the same shape).  Any cell with a `false` value in the bool mask returns a `NaN` (missing data), and `Set` functions are no-ops. The [stats](stats) packages treat `NaN` as missing data, but [tmath](tmath), [vector](vector), and [matrix](matrix) packages do not, so it is best to call `.AsValues()` on masked data prior to operating on it, in a basic math context (i.e., `copy` in Goal).

* `Indexed` has a tensor of indexes into the source data, where the final, innermost dimension of the indexes is the same size as the number of dimensions in the wrapped source tensor. The overall shape of this view is that of the remaining outer dimensions of the Indexes tensor, and like other views, assignment and return values are taken from the corresponding indexed value in the wrapped source tensor.

* `Reshaped` applies a different `Shape` to the source tensor, with the constraint that the new shape has the same length of total elements as the source tensor. It is particularly useful for aligning different tensors binary operation between them produces the desired results, for example by adding a new axis or collapsing multiple dimensions into one.

* `Rows` is a specialized version of `Sliced` that provides a row index-based view, with the `Indexes` applying to the outermost _row_ dimension, which allows sorting and filtering to operate only on the indexes, leaving the underlying Tensor unchanged. This view is returned by the [table](table) data table, which organizes multiple heterogenous Tensor columns along a common outer row dimension, and provides similar functionality to pandas and particularly [xarray](http://xarray.pydata.org/en/stable/) in Python. 

Note that any view can be "stacked" on top of another, to produce more complex net views.

Each view type implements the `AsValues` method to create a concrete "rendered" version of the view (as a `Values` tensor) where the actual underlying data is organized as it appears in the view. This is like the `copy` function in NumPy, disconnecting the view from the original source data. Note that unlike NumPy, `Masked` and `Indexed` remain views into the underlying source data -- see [Basic and Advanced Indexing](#basic-and-advanced-indexing) below.

The `float64` ("Float"), `int` ("Int"), and `string` ("String") types are used as universal input / output types, and for intermediate computation in the math functions. Any performance-critical code can be optimized for a specific data type, but these universal interfaces are suitable for misc ad-hoc data analysis.

There is also a `RowMajor` sub-interface for tensors (implemented by the `Values` and `Rows` types), which supports `[Set]FloatRow[Cell]` methods that provide optimized access to row major data. See [Standard shapes](#standard-shapes) for more info.

The `Vectorize` function and its variants provide a universal "apply function to tensor data" mechanism (often called a "map" function, but that name is already taken in Go). It takes an `N` function that determines how many indexes to iterate over (and this function can also do any initialization prior to iterating), a compute function that gets the current index value, and a varargs list of tensors. In general it is completely up to the compute function how to interpret the index, although we also support the "broadcasting" principles from NumPy for binary functions operating on two tensors, as discussed below. There is a Threaded version of this for parallelizable functions, and a GPU version in the [gosl](../gpu/gosl) Go-as-a-shading-language package.

To support the best possible performance in compute-intensive code, we have written all the core tensor functions in an `Out` suffixed version that takes the output tensor as an additional input argument (it must be a `Values` type), which allows an appropriately sized tensor to be used to hold the outputs on repeated function calls, instead of requiring new memory allocations every time. These versions are used in other calls where appropriate. The function without the `Out` suffix just wraps the `Out` version, and is what is called directly by Goal, where the output return value is essential for proper chaining of operations.

To support proper argument handling for tensor functions, the [goal](../goal) transpiler registers all tensor package functions into the global name-to-function map (`tensor.Funcs`), which is used to retrieve the function by name, along with relevant arg metadata. This registry is also key for enum sets of functions, in the `stats` and `metrics` packages, for example, to be able to call the corresponding function. Goal uses symbols collected in the [yaegicore](../yaegicore) package to populate the Funcs, but enums should directly add themselves to ensure they are always available even outside of Goal.

* [table](table) organizes multiple Tensors as columns in a data `Table`, aligned by a common outer row dimension. Because the columns are tensors, each cell (value associated with a given row) can also be n-dimensional, allowing efficient representation of patterns and other high-dimensional data. Furthermore, the entire column is organized as a single contiguous slice of data, so it can be efficiently processed. A `Table` automatically supplies a shared list of row Indexes for its `Indexed` columns, efficiently allowing all the heterogeneous data columns to be sorted and filtered together.

    Data that is encoded as a slice of `struct`s can be bidirectionally converted to / from a Table, which then provides more powerful sorting, filtering and other functionality, including [plot/plotcore](../plot/plotcore).

* [tensorfs](tensorfs) provides a virtual filesystem (FS) for organizing arbitrary collections of data, supporting interactive, ad-hoc (notebook style) as well as systematic data processing. Interactive [goal](../goal) shell commands (`cd`, `ls`, `mkdir` etc) can be used to navigate the data space, with numerical expressions immediately available to operate on the data and save results back to the filesystem. Furthermore, the data can be directly copied to / from the OS filesystem to persist it, and `goal` can transparently access data on remote systems through ssh. Furthermore, the [databrowser](databrowser) provides a fully interactive GUI for inspecting and plotting data.

* [tensorcore](tensorcore) provides core widgets for graphically displaying the `Tensor` and `Table` data, which are used in `tensorfs`.

* [tmath](tmath) implements all standard math functions on `tensor.Indexed` data, including the standard `+, -, *, /` operators. `goal` then calls these functions.

* [plot/plotcore](../plot/plotcore) supports interactive plotting of `Table` data.

* [bitslice](bitslice) is a Go slice of bytes `[]byte` that has methods for setting individual bits, as if it was a slice of bools, while being 8x more memory efficient. This is used for encoding null entries in  `etensor`, and as a Tensor of bool / bits there as well, and is generally very useful for binary (boolean) data.

* [stats](stats) implements a number of different ways of analyzing tensor and table data, including:
    - [cluster](cluster) implements agglomerative clustering of items based on [metric](metric) distance / similarity matrix data.
    - [convolve](convolve) convolves data (e.g., for smoothing).
    - [glm](glm) fits a general linear model for one or more dependent variables as a function of one or more independent variables. This encompasses all forms of regression.
    - [histogram](histogram) bins data into groups and reports the frequency of elements in the bins.
    - [metric](metric) computes similarity / distance metrics for comparing two tensors, and associated distance / similarity matrix functions, including PCA and SVD analysis functions that operate on a covariance matrix.
    - [stats](stats) provides a set of standard summary statistics on a range of different data types, including basic slices of floats, to tensor and table data. It also includes the ability to extract Groups of values and generate statistics for each group, as in a "pivot table" in a spreadsheet.

# Standard shapes

There are various standard shapes of tensor data that different functions expect, listed below. The two most general-purpose functions for shaping and slicing any tensor to get it into the right shape for a given computation are:

* `Reshape` returns a `Reshaped` view with the same total length as the source tensor, functioning like the NumPy `reshape` function.

* `Reslice` returns a re-sliced view of a tensor, extracting or rearranging dimenstions. It supports the full NumPy [basic indexing](https://numpy.org/doc/stable/user/basics.indexing.html#basic-indexing) syntax. It also does reshaping as needed, including processing the `NewAxis` option.

* **Flat, 1D**: this is the simplest data shape, and any tensor can be turned into a flat 1D list using `NewReshaped(-1)` or the `As1D` function, which either returns the tensor itself it is already 1D, or a `Reshaped` 1D view. The [stats](stats) functions for example report summary statistics across the outermost row dimension, so converting data to this 1D view gives stats across all the data.

* **Row, Cell 2D**: This is the natural shape for tabular data, and the `RowMajor` type and `Rows` view provide methods for efficiently accessing data in this way. In addition, the [stats](stats) and [metric](metric) packages automatically compute statistics across the outermost row dimension, aggregating results across rows for each cell. Thus, you end up with the "average cell-wise pattern" when you do `stats.Mean` for example. The `NewRowCellsView` function returns a `Reshaped` view of any tensor organized into this 2D shape, with the row vs. cell split specified at any point in the list of dimensions, which can be useful in obtaining the desired results.

* **Matrix 2D**: For matrix algebra functions, a 2D tensor is treated as a standard row-major 2D matrix, which can be processed using `gonum` based matrix and vector operations, as in the [matrix](matrix) package.

* **Matrix 3+D**: For functions that specifically process 2D matricies, a 3+D shape can be used as well, which iterates over the outer dimensions to process the inner 2D matricies.

## Dynamic row sizing (e.g., for logs)

The `SetNumRows` function can be used to progressively increase the number of rows to fit more data, as is typically the case when logging data (often using a [table](table)). You can set the row dimension to 0 to start -- that is (now) safe. However, for greatest efficiency, it is best to set the number of rows to the largest expected size first, and _then_ set it back to 0. The underlying slice of data retains its capacity when sized back down. During incremental increasing of the slice size, if it runs out of capacity, all the elements need to be copied, so it is more efficient to establish the capacity up front instead of having multiple incremental re-allocations.

# Printing format

The following are examples of tensor printing via the `Sprintf` function, which is used with default values for the `String()` stringer method on tensors. It does a 2D projection of higher-dimensional tensors, using the `Projection2D` set of functions, which assume a row-wise outermost dimension in general, and pack even sets of inner dimensions into 2D row x col shapes (see examples below).

1D (vector): goes column-wise, and wraps around as needed, e.g., length = 4:
```
[4] 0 1 2 3 
```
and 40:
```
[40]  0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 
     25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 
```

2D matrix:
```
[4 3]
    [0] [1] [2] 
[0]   0   1   2 
[1]  10  11  12 
[2]  20  21  22 
[3]  30  31  32 
```
and a column vector (2nd dimension is 1):
```
[4 1]
[0] 0 
[1] 1 
[2] 2 
[3] 3 
```

3D tensor, shape = `[4 3 2]` -- note the `[r r c]` legend below the shape which indicates which dimensions are shown on the row (`r`) vs column (`c`) axis, so you know how to interpret the indexes:
```
[4 3 2]
[r r c] [0] [1] 
[0 0]     0   1 
[0 1]    10  11 
[0 2]    20  21 
[1 0]   100 101 
[1 1]   110 111 
[1 2]   120 121 
[2 0]   200 201 
[2 1]   210 211 
[2 2]   220 221 
[3 0]   300 301 
[3 1]   310 311 
[3 2]   320 321 
```

4D tensor: note how the row, column dimensions alternate, resulting in a 2D layout of the outer 2 dimensions, with another 2D layout of the inner 2D dimensions inlaid within that:
```
[5 4 3 2]
[r c r c] [0 0] [0 1] [1 0] [1 1] [2 0] [2 1] [3 0] [3 1] 
[0 0]         0     1   100   101   200   201   300   301 
[0 1]        10    11   110   111   210   211   310   311 
[0 2]        20    21   120   121   220   221   320   321 
[1 0]      1000  1001  1100  1101  1200  1201  1300  1301 
[1 1]      1010  1011  1110  1111  1210  1211  1310  1311 
[1 2]      1020  1021  1120  1121  1220  1221  1320  1321 
[2 0]      2000  2001  2100  2101  2200  2201  2300  2301 
[2 1]      2010  2011  2110  2111  2210  2211  2310  2311 
[2 2]      2020  2021  2120  2121  2220  2221  2320  2321 
[3 0]      3000  3001  3100  3101  3200  3201  3300  3301 
[3 1]      3010  3011  3110  3111  3210  3211  3310  3311 
[3 2]      3020  3021  3120  3121  3220  3221  3320  3321 
[4 0]      4000  4001  4100  4101  4200  4201  4300  4301 
[4 1]      4010  4011  4110  4111  4210  4211  4310  4311 
[4 2]      4020  4021  4120  4121  4220  4221  4320  4321 
```

5D tensor: is treated like a 4D with the outermost dimension being an additional row dimension:
```
[6 5 4 3 2]
[r r c r c] [0 0] [0 1] [1 0] [1 1] [2 0] [2 1] [3 0] [3 1] 
[0 0 0]         0     1   100   101   200   201   300   301 
[0 0 1]        10    11   110   111   210   211   310   311 
[0 0 2]        20    21   120   121   220   221   320   321 
[0 1 0]      1000  1001  1100  1101  1200  1201  1300  1301 
[0 1 1]      1010  1011  1110  1111  1210  1211  1310  1311 
[0 1 2]      1020  1021  1120  1121  1220  1221  1320  1321 
[0 2 0]      2000  2001  2100  2101  2200  2201  2300  2301 
[0 2 1]      2010  2011  2110  2111  2210  2211  2310  2311 
[0 2 2]      2020  2021  2120  2121  2220  2221  2320  2321 
[0 3 0]      3000  3001  3100  3101  3200  3201  3300  3301 
[0 3 1]      3010  3011  3110  3111  3210  3211  3310  3311 
[0 3 2]      3020  3021  3120  3121  3220  3221  3320  3321 
[0 4 0]      4000  4001  4100  4101  4200  4201  4300  4301 
[0 4 1]      4010  4011  4110  4111  4210  4211  4310  4311 
[0 4 2]      4020  4021  4120  4121  4220  4221  4320  4321 
[1 0 0]     10000 10001 10100 10101 10200 10201 10300 10301 
[1 0 1]     10010 10011 10110 10111 10210 10211 10310 10311 
[1 0 2]     10020 10021 10120 10121 10220 10221 10320 10321 
[1 1 0]     11000 11001 11100 11101 11200 11201 11300 11301 
[1 1 1]     11010 11011 11110 11111 11210 11211 11310 11311 
...
```

# History

This package was originally developed as [etable](https://github.com/emer/etable) as part of the _emergent_ software framework. It always depended on the GUI framework that became Cogent Core, and having it integrated within the Core monorepo makes it easier to integrate updates, and also makes it easier to build advanced data management and visualization applications. For example, the [plot/plotcore](../plot/plotcore) package uses the `Table` to support flexible and powerful plotting functionality.

It was completely rewritten in Sept 2024 to use a single data type (`tensor.Indexed`) and call signature for compute functions taking these args, to provide a simple and efficient data processing framework that greatly simplified the code and enables the [goal](../goal) language to directly transpile simplified math expressions into corresponding tensor compute code.


