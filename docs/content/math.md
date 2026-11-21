+++
Categories = ["Goal"]
+++

**Math** mode in [[Goal]] provides a [NumPy](https://numpy.org/doc/stable/index.html)-like language for mathematical expressions involving [[tensor]] data elements (see [[#Reference tables]]), which are transpiled into Go code for compilation or interactive interpretation. It is activated by a `#` character on a line, which is otherwise not a recognized symbol in Go, and `##` starts and stops multi-line blocks of math mode.

Variables defined with `:=` in math mode are tensors, and its operators work on tensors. Plain Go values, such as a `float64` defined outside of math mode, are converted to tensors where needed, and can also be converted explicitly using the `array` function. All of Go can be used in math mode as well; see [[#Go values and code]].

```Goal
fv := 42.0 // this is a float64

// now we enter math mode:
##
tv := 42.0 // this is a tensor.Float64
tfv := array(fv) // as is this
##

fmt.Printf("fv: %v %T\n", fv, fv)
fmt.Printf("tv: %v %T\n", tv, tv)
fmt.Printf("tfv: %v %T\n", tfv, tfv)
```

## Go values and code

Math mode transpiles each line into Go on its own, based only on its syntax, so it does not know whether a variable that it did not define is a tensor or a Go value. Instead, it uses functions that work for both, so that tensors and Go values can be mixed freely:

* Go numbers, bools and strings, and slices of them, can be used in math expressions with tensors: they are converted to tensors where needed (with `tensor.ToTensor`).
* Assignments, including assignment operators like `+=`, work for variables that are either tensors or Go values.
* Conversions get Go values from tensors: `float64(x)`, `int(x)`, `bool(x)` and `string(x)` work for both tensors and Go values.
* Conditions of `if`, `for` and `switch` statements can be tensors or Go `bool`s.

All of Go can be used in math mode, including `for` loops, `if` and `switch` statements, `var` declarations, and composite and function literals:

* The variables of `for` loops are Go values. In `for i := range n` and `for i, v := range x`, `i` is an `int`, and `v` is a `float64` if `x` is a tensor, while `n` and `x` can be either tensors or Go values. The init and post statements of a `for i := 0; i < n; i++` loop are Go.
* A `switch` with a tag compares Go values, so a tensor tag must be converted, as in `switch int(x)`.

Operators on Go values that math mode knows about are Go operators, which are much faster than tensor operations. These are the variables of `for` loops, variables declared with `var`, variables defined with `:=` as a conversion such as `int(x)` or as an expression of other such Go values, and function parameters of basic types, as well as variables defined as a number outside of math mode, as in `n := 10`. As in Go, the types must match: an `int` times a `float64` is done with tensors (and `float64(i) * x` with Go).

```Goal
tau := 0.5 // a float64
##
a := zeros(5)
for i := range 5 {
	a[i] = float64(i) * tau
}
sum := 0.0
for _, v := range a {
	if v > 1 {
		sum += v
	}
}
n := int(sum) // a Go int
switch {
case n > 2:
	fmt.Println("big:", n)
default:
	fmt.Println("small:", n)
}
##
fmt.Println(a, sum)
```

## Basics

Here's how you can create, inspect, and manipulate tensor data:

```Goal
##
a := [[1., 2., 3.], [4., 5., 6.]] 
aShape := a.shape  // tensor of sizes of each dim
b := zeros(aShape) // which can be used to make a new one
c := a.reshape(3,2) // preserves data while reshaping
d := arange(1, 11)  // like a for loop from 1..11 (exclusive max)
e := linspace(1., 3., 11, true) // floats over a range: final arg = inclusive
##

fmt.Println("a:", a)
fmt.Println("aShape:", aShape)
fmt.Println("b:", b)
fmt.Println("c:", c)
fmt.Println("d:", d)
fmt.Println("e:", e)
```

(go ahead and play around with any of the above expressions to explore the effects!)

Note that, as in Python, you do need to add a decimal point to have a number treated as a floating point value in most cases -- otherwise it will be an int.

You can perform math operations directly using standard operators:

```Goal
##
a := [[1., 2., 3.], [4., 5., 6.]] 
b := a * a
c := sin(a)
d := a * [3., 2., 1.] // smaller dims apply repeatedly
##

fmt.Println("a:", a)
fmt.Println("b:", b)
fmt.Println("c:", c)
fmt.Println("d:", d)
```

See [[tensor math#Alignment of shapes]] for more details on [[tensor math]] operations, using the NumPy [broadcasting](https://numpy.org/doc/stable/user/basics.broadcasting.html) logic.

### TensorFS

In an interactive Goal shell (which we simulate here in the docs), variables in math mode are automatically saved to the [[tensorfs]] virtual data filesystem:

```Goal
##
// make a new tensorfs directory for this example
mkdir tfs
cd tfs
a := [[1., 2., 3.], [4., 5., 6.]]
setcp("a_copy", a) // this will preserve values

b := a * a
a += a

// list the current directory:
ls -l
// go back to root (all vars from this page are there!)
cd ..
// (add another ls -l here to see them all..)
##

fmt.Println("a:", a)
fmt.Println("a get:", # get("tfs/a") #)
fmt.Println("a copy:", # get("tfs/a_copy") #)
fmt.Println("b:", b)
```

Note that the filesystem variables are pointers to the live variables, so they always reflect the latest changes, so the `setcp` command is useful for saving a copy that does not get further updated.

### Tables

The columns of a [[table]] are indexed by name, as in pandas: `dt["mass"]` is the column named `mass`, and assigning to a column by name, as in `dt["ratio"] = dt["mass"] / dt["size"]`, sets its values, adding the column if the table does not have it. The functions `unique`, `isin`, and `where` work well with the columns of tables (see [[#Selecting and counting values]]), and a boolean mask selects the rows of a view of a table with `FilterMask`. See the [[planets]] tutorial for a full example.

```Goal
dt := table.New("sizes")
dt.AddStringColumn("name")
dt.AddFloat64Column("size")
dt.SetNumRows(4)
##
dt["name"] = ["small", "big", "small", "medium"]
dt["size"] = [1., 10., 2., 5.]
dt["double"] = dt["size"] * 2
names := unique(dt["name"])
big := dt["size"] > 3
##
fmt.Println("double:", dt.Column("double"))
fmt.Println("names:", names)
bigOnes, _ := table.NewView(dt).FilterMask(big)
fmt.Println("big:", bigOnes.Column("name"))
```

### Reductions, sorting, and stacking

As in NumPy, `sum(a)`, `max(a)`, `mean(a)`, and so on reduce all of the values of a tensor to one value, and with an axis, as in `a.max(0)` or `sum(a, 1)`, the values along that axis. `sort` and `argsort` sort along the last axis by default, and `vstack`, `hstack`, `concatenate`, `stack`, and `tile` make new tensors from others (see [[#Reference tables]] for all of the functions):

```Goal
##
a := [[3., 1., 2.], [9., 7., 8.]]
total := sum(a)
colMax := a.max(0)
rowSum := sum(a, 1)
sorted := sort(a)
byFirst := a[argsort(-a[:, 0]), :]
more := vstack(a, [0., 0., 0.])
xs, ys := meshgrid([1., 2., 3.], [10., 20.])
##
fmt.Println("total:", total, "colMax:", colMax, "rowSum:", rowSum)
fmt.Println("sorted:", sorted)
fmt.Println("byFirst:", byFirst)
fmt.Println("more:", more)
fmt.Println("xs:", xs, "ys:", ys)
```

## Slicing and indexing

Math mode provides NumPy-like ways of extracting data from a tensor (examples follow Go versions in [[tensor#Views and values]]).

```Goal
##
x := linspace(0., 12., 12., false).reshape(3,4)

row1 := x[1]
col1 := x[:,1]
##

fmt.Println("x:", x)
fmt.Println("row1:", row1)
fmt.Println("col:", col1)
```

Where `:` is an empty `slice` expression that indicates all values in that dimension.

To get the column as a column vector, use `newaxis`:

```Goal
##
x := linspace(0., 12., 12., false).reshape(3,4)

col1 := x[:, 1, newaxis]
##

fmt.Println("col1:", col1)
```

To get ranges within each dimension, or reorder, use `slice` expressions similar to those used in accessing Go slices, but with a 3rd `Step` value, as in a standard Go `for` loop:

```Goal
##
x := linspace(0., 12., 12., false).reshape(3,4)

row1 := x[1, 1:3]  // only a subset of columns
col1 := x[::-1, 1] // reverse order of rows dimension
##

fmt.Println("row1:", row1)
fmt.Println("col1:", col1)
```

Ellipsis (`...`) makes it easy to get the last dimension(s):

```Goal
##
x := linspace(0., 12., 12., false).reshape(3,2,2)

last1 := x[..., 1]
##

fmt.Println("x", x)
fmt.Println("last1:", last1)
```

As in [NumPy](https://numpy.org/doc/stable/index.html) (and standard Go slices), indexed subsets of a tensor are _views_ onto the original tensor, so that changes to the original values are immediately seen in these views. Use `copy` to make a new separate copy of the values to break this connection.

```Goal
##
x := linspace(0., 12., 12., false).reshape(3,2,2)

last1 := x[..., 1]
cp1 := copy(last1)

x[..., 1, 1] = 3.14 // note how we assign to all rows
##

fmt.Println("x", x)
fmt.Println("last1:", last1)
fmt.Println("cp1:", cp1)
```

### Masked by booleans

```Goal
##
x := linspace(0., 12., 12., false).reshape(3,4)

m := x[x>=6]
mi := x >= 6
##

fmt.Println("m:", m)
fmt.Println("mi:", mi)
```

### Arbitrary indexes

```Goal
##
x := linspace(0., 12., 12., false).reshape(3,4)

ixs := [[0, 1], [0, 1], [0, 2], [0, 2], [1, 1], [1, 1], [2, 2], [2, 2]].reshape(2,4,2)

ix := x[ixs]
##

fmt.Println("ix:", ix)
fmt.Println("ixs:", ixs)
```


## Reference tables

The following tables summarize Goal math-mode syntax in terms of [NumPy](https://numpy.org/doc/stable/index.html) and the underlying Go code generated. For MATLAB equivalents, see [numpy-for-matlab-users](https://numpy.org/doc/stable/user/numpy-for-matlab-users.html).

* The _same:_ in Goal means that the same NumPy syntax works in Goal, minus the `np.` prefix, and likewise for _or:_ (where Goal also has additional syntax).
* In the `tensor.Go` code, we sometimes just write a scalar number for simplicity, but these are actually `tensor.NewFloat64Scalar` etc.
* Goal also has support for `string` tensors, e.g., for labels, and operators such as addition that make sense for strings are supported. Otherwise, strings are automatically converted to numbers using the `tensor.Float` interface. If you have any doubt about whether you've got a `tensor.Float64` when you expect one, use `tensor.AsFloat64Tensor` which makes sure.

### Tensor shape

| `tensor` Go  |   Goal      | NumPy   | Notes            |
| ------------ | ----------- | ------  | ---------------- |
| `a.NumDim()` | `ndim(a)` or `a.ndim` | `np.ndim(a)` or `a.ndim`   | number of dimensions of tensor `a` |
| `a.Len()`    | `len(a)` or `a.len` or: | `np.size(a)` or `a.size`   | number of elements of tensor `a` |
| `a.Shape().Sizes` | same: | `np.shape(a)` or `a.shape` | "size" of each dimension in a; `shape` returns a 1D `int` tensor |
| `a.Shape().Sizes[1]` | same: | `a.shape[1]` | the number of elements of the 2nd dimension of tensor `a` |
| `tensor.Reshape(a, 10, 2)` | same except no `a.shape = (10,2)`: | `a.reshape(10, 2)` or `np.reshape(a, 10, 2)` or `a.shape = (10,2)` | set the shape of `a` to a new shape that has the same total number of values (len or size); No option to change order in Goal: always row major; Goal does _not_ support direct shape assignment version. |
| `tensor.Reshape(a, tensor.AsIntSlice(sh)...)` | same: | `a.reshape(10, sh)` or `np.reshape(a, sh)` | set shape based on list of dimension sizes in tensor `sh` |
| `tensor.Reshape(a, -1)` or `tensor.As1D(a)` | same: | `a.reshape(-1)` or `np.reshape(a, -1)` | a 1D vector view of `a`; Goal does not support `ravel`, which is nearly identical. |
| `tensor.Flatten(a)` | same: | `b = a.flatten()`   | returns a 1D copy of a |
| `b := tensor.Clone(a)` | `b := copy(a)` or: | `b = a.copy()` | direct assignment `b = a` in Goal or NumPy just makes variable b point to tensor a; `copy` is needed to generate new underlying values (MATLAB always makes a copy) |
| `tensor.Squeeze(a)` | same: |`a.squeeze()` | remove singleton dimensions of tensor `a`. |


### Constructing

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `tensor.NewFloat64FromValues(` `1, 2, 3)` | `[1., 2., 3.]` | `np.array([1., 2., 3.])` | define a 1D tensor |
| (reshape) | `[[1., 2., 3.], [4., 5., 6.]]` or: | `(np.array([[1., 2., 3.], [4., 5., 6.]])` | define a 2x3 2D tensor |
| `tensor.VStack(` `tensor.HStack(a, b),` `tensor.HStack(c, d))` | `vstack(hstack(a, b), hstack(c, d))` | `np.block([[a, b], [c, d]])` | construct a matrix from blocks `a`, `b`, `c`, and `d` |
| `tensor.NewFloat64(3,4)` | `zeros(3,4)` | `np.zeros((3, 4))` | 3x4 2D tensor of float64 zeros; Goal does not use "tuple" so no double parens |
| `tensor.NewFloat64(3,4,5)` | `zeros(3, 4, 5)` | `np.zeros((3, 4, 5))` | 3x4x5 three-dimensional tensor of float64 zeros |
| `tensor.NewFloat64Ones(3,4)` | `ones(3, 4)`  | `np.ones((3, 4))` | 3x4 2D tensor of 64-bit floating point ones |
| `tensor.NewFloat64Full(5.5, 3,4)` | `full(5.5, 3, 4)` | `np.full((3, 4), 5.5)` | 3x4 2D tensor of 5.5; Goal variadic arg structure requires value to come first |
| `tensor.NewFloat64Rand(3,4)` | `rand(3, 4)` | `rng.random(3, 4)` | 3x4 2D float64 tensor with uniform random 0..1 elements, using the current Go `rand` source |
| `tensor.Concatenate(1, a, b)` | `concatenate(1, a, b)` or `hstack(a, b)` or `column_stack(a, b)` |`np.concatenate((a,b),1)` or `np.hstack((a,b))` or `np.column_stack((a,b))` or `np.c_[a,b]` | concatenate columns of a and b; Goal has no tuples, so the axis comes first, as the value does in `full`; `column_stack` makes 1D tensors columns |
| `tensor.Concatenate(0, a, b)` | `concatenate(0, a, b)` or `vstack(a, b)` |`np.concatenate((a,b))` or `np.vstack((a,b))` or `np.r_[a,b]` | concatenate rows of a and b; `vstack` makes 1D tensors rows |
| `tensor.Stack(0, a, b)` | `stack(0, a, b)` |`np.stack((a,b))` | stack tensors of the same shape along a new axis (here the first) |
| `tensor.Tile(a, m, n)` | `tile(a, m, n)` |`np.tile(a, (m, n))`   | create m by n copies of a |
| `tensor.VStack(a,` `tensor.AnySlice(a, 0))` | `vstack(a, a[0])` |`a[np.r_[:len(a),0]]`  | `a` with copy of the first row appended to the end |

### Ranges and grids

See [NumPy](https://numpy.org/doc/stable/user/how-to-partition.html) docs for details.

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `tensor.NewIntRange(1, 11)` | same: |`np.arange(1., 11.)` or `np.r_[1.:11.]` or `np.r_[1:10:10j]` | create an increasing vector; `arange` in goal is always ints; use `linspace` or `tensor.AsFloat64` for floats |
| . | same: |`np.arange(10.)` or `np.r_[:10.]` or `np.r_[:9:10j]` | create an increasing vector; 1 arg is the stop value in a slice |
| . | `arange(1, 11)[:, newaxis]` |`np.arange(1.,11.)` `[:, np.newaxis]` | create a column vector |
| `t.NewFloat64` `SpacedLinear(` `1, 3, 4, true)` | `linspace(1,3,4,true)` |`np.linspace(1,3,4)` | 4 equally spaced samples between 1 and 3, inclusive of end (use `false` at end for exclusive) |
| `xs, ys := tensor.Meshgrid(x, y)` | `xs, ys := meshgrid(x, y)` |`np.meshgrid(x, y)` or `np.mgrid[0:9.,0:6.]` | two 2D tensors of the points of a grid: one of x values (varying along the columns), the other of y values (varying along the rows), with a row for each y and a column for each x |
| . | `x[newaxis, :]` and `y[:, newaxis]` |`np.ogrid[0:9.,0:6.]` or `np.ix_(x, y)` | a row of x values and a column of y values, which broadcasting makes into a grid in element-wise operations, as in `x[newaxis, :] * y[:, newaxis]`: the best way to eval functions on a grid |

### Basic indexing

See [NumPy basic indexing](https://numpy.org/doc/stable/user/basics.indexing.html#basic-indexing). Tensor Go uses the `Reslice` function for all cases (repeated `tensor.` prefix replaced with `t.` to take less space). Here you can clearly see the advantage of Goal in allowing significantly more succinct expressions to be written for accomplishing critical tensor functionality.

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `t.Reslice(a, 1, 4)` | same: |`a[1, 4]` | access element in second row, fifth column in 2D tensor `a` |
| `t.Reslice(a, -1)` | same: |`a[-1]` | access last element |
| `t.Reslice(a,` `1, t.FullAxis)` | same: |`a[1]` or `a[1, :]` | entire second row of 2D tensor `a`; unspecified dimensions are equivalent to `:` (could omit second arg in Reslice too) |
| `t.Reslice(a,` `Slice{Stop:5})` | same: |`a[0:5]` or `a[:5]` or `a[0:5, :]` | 0..4 rows of `a`; uses same Go slice ranging here: (start:stop) where stop is _exclusive_ |
| `t.Reslice(a,` `Slice{Start:-5})` | same: |`a[-5:]` | last 5 rows of 2D tensor `a` |
| `t.Reslice(a,` `t.NewAxis,` `Slice{Start:-5})` | same: |`a[newaxis, -5:]` | last 5 rows of 2D tensor `a`, as a column vector |
| `t.Reslice(a,` `Slice{Stop:3},` `Slice{Start:4, Stop:9})` | same: |`a[0:3, 4:9]` | The first through third rows and fifth through ninth columns of a 2D tensor, `a`. |
| `t.Reslice(a,` `Slice{Start:2,` `Stop:25,` `Step:2}, t.FullAxis)` | same: |`a[2:21:2,:]` | every other row of `a`, starting with the third and going to the twenty-first |
| `t.Reslice(a,` `Slice{Step:2},` `t.FullAxis)` | same: |`a[::2, :]`  | every other row of `a`, starting with the first |
| `t.Reslice(a,`, `Slice{Step:-1},` `t.FullAxis)` | same: |`a[::-1,:]`  | `a` with rows in reverse order |
| `t.Clone(t.Reslice(a,` `1, t.FullAxis))` | `b = copy(a[1, :])` or: | b = a[1, :].copy()` | without the copy, `y` would point to a view of values in `x`; `copy` creates distinct values, in this case of _only_ the 2nd row of `x` -- i.e., it "concretizes" a given view into a literal, memory-continuous set of values for that view. |
| `tmath.Assign(` `t.Reslice(a,` `Slice{Stop:5}),` `t.NewIntScalar(2))` | same: |`a[:5] = 2` | assign the value 2 to 0..4 rows of `a` |
| (you get the idea) | same: |`a[:5] = b[:, :5]` | assign the values in the first 5 columns of `b` to the first 5 rows of `a` |

### Boolean tensors and indexing

See [NumPy boolean indexing](https://numpy.org/doc/stable/user/basics.indexing.html#boolean-array-indexing).

Note that Goal only supports boolean logical operators (`&&` and `||`) on boolean tensors, not the single bitwise operators `&` and `|`.

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `tmath.Greater(a, 0.5)` | same: | `(a > 0.5)` | `bool` tensor of shape `a` with elements `(v > 0.5)` |
| `tmath.And(a, b)` | `a && b` | `logical_and(a,b)` | element-wise AND operator on `bool` tensors |
| `tmath.Or(a, b)` | `a \|\| b` | `np.logical_or(a,b)` | element-wise OR operator on `bool` tensors | 
| `tmath.Negate(a)` | `!a` | ? | element-wise negation on `bool` tensors | 
| `tmath.Assign(` `tensor.Mask(a,` `tmath.Less(a, 0.5),` `0)` | same: |`a[a < 0.5]=0` | `a` with elements less than 0.5 zeroed out |
| `tensor.Flatten(` `tensor.Mask(a,` `tmath.Less(a, 0.5)))` | same: |`a[a < 0.5].flatten()` | a 1D list of the elements of `a` < 0.5 (as a copy, not a view) |
| `tensor.Mul(a,` `tmath.Greater(a, 0.5))` | same: |`a * (a > 0.5)` | `a` with elements less than 0.5 zeroed out |

### Advanced index-based indexing

See [NumPy integer indexing](https://numpy.org/doc/stable/user/basics.indexing.html#integer-array-indexing).  Note that the current NumPy version of indexed is rather complex and difficult for many people to understand, as articulated in this [NEP 21 proposal](https://numpy.org/neps/nep-0021-advanced-indexing.html). 

In Goal, a list of indexes (a tensor of more than one value) for one dimension selects those indexes of that dimension, in that order, and a `bool` tensor for one dimension selects the indexes where it is true. Lists of indexes for more than one dimension select all of the combinations of them, as `np.ix_` does in NumPy (unlike NumPy without `ix_`).

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `t.Reslice(a,` `t.NewIntFromValues(1, 3, 4),` `t.NewIntFromValues(0, 2))` | `a[[1, 3, 4], [0, 2]]` |`a[np.ix_([1, 3, 4], [0, 2])]` | rows 2,4 and 5 and columns 1 and 3. |
| `t.ArgWhere(` `tmath.Greater(a, 0.5))` | `argwhere(a > 0.5)` |`np.argwhere(a > 0.5)` or `np.nonzero(a > 0.5)` | find the indices where (a > 0.5), as a 2D tensor with a row of the index in each dimension of each value, which are the arbitrary indexes of Goal: `a[argwhere(a > 0.5)]` |
| `t.Reslice(a,` `t.FullAxis,` `tmath.Greater(v, 0.5))` | `a[:, v > 0.5]` |`a[:, v.T > 0.5]` or `a[:,np.nonzero(v > 0.5)[0]]` | extract the columns of `a` where vector `v` > 0.5 |
| `tmath.Assign(` `t.Reslice(a,` `t.FullAxis), 3)` | same: |`a[:] = 3` | set all values to the same scalar value |
| `t.Sorted(a)` | `sort(a)` |`np.sort(a)` or `a.sort()` | sort each row of a 2D tensor, `a` (along its last axis); `NaN` values are last |
| `t.Sorted(a, 0)` | `sort(a, 0)` |`np.sort(a, axis=0)` or `a.sort(axis=0)` | sort each column of 2D tensor, `a` |
| `t.ArgSort(a)` | `argsort(a)` |`np.argsort(a)` | the indexes that sort each row of `a` (along its last axis), or along the given axis |
| `t.Reslice(a,` `t.ArgSort(...),` `t.FullAxis)` | `b := a[argsort(a[:, 0]), :]` |`I = np.argsort(a[:, 0]); b = a[I,:]` | save the tensor `a` as tensor `b` with rows sorted by the first column |

### Selecting and counting values

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `tensor.Unique(a)` | `unique(a)` | `np.unique(a)` | a sorted vector of the distinct values of tensor `a`, of its type, without `NaN` (missing data) |
| `tensor.UniqueCounts(a)` | `unique_counts(a)` | `np.unique_counts(a).counts` | the number of times that each value of `unique(a)` occurs |
| `tmath.IsIn(a, v)` | `isin(a, v)` | `np.isin(a, v)` | `bool` tensor of shape `a`, true where the value of `a` is one of the values of `v` |
| `tmath.Where(c, a, b)` | `where(c, a, b)` | `np.where(c, a, b)` | the values of `a` where the `bool` tensor `c` is true, and of `b` where it is false, with broadcasting (e.g., a scalar `b`); strings if `a` or `b` has strings |
| `dt.Column("mass")` | `dt["mass"]` | `df["mass"]` | the column of table `dt` named `mass` (pandas DataFrame `df`) |
| `dt.SetColumn("x", v)` | `dt["x"] = v` | `df["x"] = v` | set the values of column `x` of table `dt`, adding it if needed |
| `dt.FilterMask(m)` | same: | `df[m]` | select the rows of table `dt` where the `bool` tensor `m` is true (on a view: `table.NewView(dt).FilterMask(m)`) |
| `m, rl, cl := tensor.Pivot(` `r, c, v)` | same: | `df.pivot(index="r",` `columns="c", values="v")` | the values `v` of each pair of labels in `r` and `c` (e.g., the columns of the statistics of `stats.GroupCombined`), as a matrix with a row for each unique value of `r` and a column for each of `c`, which are returned as the labels `rl` and `cl`; `NaN` where a pair has no value |
| `tmath.Clip(a, lo, hi)` | `clip(a, lo, hi)` | `np.clip(a, lo, hi)` | the values of `a` limited to the range from `lo` to `hi` |

### Basic math operations (add, multiply, etc)

In Goal and NumPy, the standard `+, -, *, /` operators perform _element-wise_ operations because those are well-defined for all dimensionalities and are consistent across the different operators, whereas matrix multiplication is specifically used in a 2D linear algebra context, and is not well defined for the other operators.

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `tmath.Add(a,b)` | same: |`a + b` | element-wise addition; Goal does this string-wise for string tensors |
| `tmath.Mul(a,b)` | same: |`a * b` | element-wise multiply |
| `tmath.Div(a,b)` | same: |`a/b`   | element-wise divide. _important:_ this always produces a floating point result. |
| `tmath.Mod(a,b)` | same: |`a%b`   | element-wise modulous (works for float and int) |
| `tmath.Pow(a,3)` | same: | `a**3`  | element-wise exponentiation |
| `tmath.Cos(a)`   | same: | `cos(a)` | element-wise function application |
| `tmath.CumSum(a)` | `cumsum(a)` or `cumsum(a, 0)` | `np.cumsum(a)` or `np.cumsum(a, axis=0)` | cumulative sums of all of the values (1D), or along the given axis; likewise `cumprod` |
| `tmath.Diff(a)` | `diff(a)` or `diff(a, 0)` | `np.diff(a)` or `np.diff(a, axis=0)` | differences between successive values along the last axis, or the given axis |

### 2D Matrix Linear Algebra

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `matrix.Mul(a,b)` | same: |`a @ b` | matrix multiply |
| `tensor.Transpose(a)` | or `a.T` |`a.transpose()` or `a.T` | transpose of `a` |
| TODO: | . |`a.conj().transpose() or a.conj().T` | conjugate transpose of `a` |
| `matrix.Det(a)` | `linalg.det(a)` or `matrix.Det(a)` | `np.linalg.det(a)` | determinant of `a` |
| `matrix.Identity(3)` | `eye(3)` or `identity(3)` |`np.eye(3)` | 3x3 identity matrix |
| `matrix.Diag(a)` | `diag(a)` |`np.diag(a)` | returns a vector of the diagonal elements of 2D tensor, `a`; `matrix.Diagonal(a)` is a read / write view of them. |
| `matrix.Diag(v)` | `diag(v)` |`np.diag(v, 0)` | returns a square diagonal matrix whose nonzero values are the elements of vector, v |
| `matrix.Trace(a)` | `trace(a)` |`np.trace(a)` | returns the sum of the elements along the diagonal of `a`. |
| `matrix.Tri(3)` | `tri(3)` |`np.tri(3)` | returns a new 2D Float64 matrix with 1s in the lower triangular region (including the diagonal) and the remaining upper triangular elements zero |
| `matrix.TriL(a)` | `tril(a)` |`np.tril(a)` | returns a copy of `a` with the lower triangular elements (including the diagonal) from `a` and the remaining upper triangular elements zeroed out |
| `matrix.TriU(a)` | `triu(a)` |`np.triu(a)` | returns a copy of `a` with the upper triangular elements (including the diagonal) from `a` and the remaining lower triangular elements zeroed out |
| `matrix.Inverse(a)` | `linalg.inv(a)` |`linalg.inv(a)` | inverse of square 2D tensor a |
| `matrix.PseudoInverse(a)` | `linalg.pinv(a)` |`linalg.pinv(a)` | pseudo-inverse of 2D tensor a |
| `matrix.Rank(a)` | `linalg.matrix_rank(a)` |`np.linalg.matrix_rank(a)` | matrix rank of a 2D tensor a |
| `matrix.Solve(a, b)` | `linalg.solve(a, b)` or `linalg.lstsq(a, b)` |`linalg.solve(a, b)` if `a` is square; `linalg.lstsq(a, b)` otherwise | solution of `a x = b` for x; the least squares solution if `a` is not square (Goal's `lstsq` returns only `x`) |
| . | . |Solve `a.T x.T = b.T` instead | solution of x a = b for x |
| `u, s, vh := matrix.SVDFactors(a)` | `u, s, vh := linalg.svd(a)` |`U, S, Vh = linalg.svd(a, full_matrices=False); V = Vh.T` | singular value decomposition of a; `matrix.SVD` is that of a symmetric matrix, as for principal components analysis |
| `matrix.Cholesky(a)` | `linalg.cholesky(a)` |`linalg.cholesky(a)` | Cholesky factorization of a 2D tensor: the lower triangular `L` where `a == L @ L.T` |
| `vecs, vals := matrix.Eig(a)` | `vecs, vals := linalg.eig(a)` |`D,V = linalg.eig(a)` | eigenvalues and eigenvectors of `a`; note that Goal returns the vectors first, and `linalg.eigh` (`matrix.EigSym`) is that of a symmetric matrix |
| . | . |`D,V = eigs(a, k=3)`  | find the k=3 largest eigenvalues and eigenvectors of 2D tensor, a |
| `q, r := matrix.QR(a)` | `q, r := linalg.qr(a)` |`Q,R = linalg.qr(a)`  | QR decomposition |
| . | . |`P,L,U = linalg.lu(a)` where `a == P@L@U` | LU decomposition with partial pivoting (note: P(MATLAB) == transpose(P(NumPy))) | 
| `matrix.Solve(Z, y)` | `x := linalg.lstsq(Z, y)` |`x = linalg.lstsq(Z, y)` | perform a linear regression of the form `Z x = y` |

### Statistics

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| `stats.Reduce(a,` `stats.StatMax)` | `a.max()` or `max(a)` | `a.max()` or `np.nanmax(a)` | maximum element of `a`, Goal always ignores `NaN` as missing data; likewise for `min`, `sum`, `prod`, `mean`, `std`, `median` and `count` (the number of values that are not `NaN`) |
| `stats.Reduce(a,` `stats.StatMax, 0)` | `a.max(0)` |`a.max(0)` | maximum element of each column of tensor `a`; also `sum(a, 0)` etc, except for `max` and `min` (see below). The Go stats functions such as `stats.Max(a)` compute the statistic of each cell over the rows (the first axis), as for the columns of a table, as `a.max(0)` does |
| `stats.Reduce(a,` `stats.StatMax, 1)` | `a.max(1)` |`a.max(1)` | maximum element of each row of tensor `a` |
| `tmath.Max(a, b)` | `max(a, b)` or `maximum(a, b)` |`np.maximum(a, b)` | compares a and b element-wise, and returns the maximum value from each pair; likewise `min` and `minimum` |
| `matrix.Norm(v)` | `linalg.norm(v)` | `np.sqrt(v @ v)` or `np.linalg.norm(v)` | L2 norm of vector v |
| . | . |`cg`  | conjugate gradients solver |

### FFT and complex numbers

todo: huge amount of work needed to support complex numbers throughout!

| `tensor` Go  |   Goal      | NumPy  | Notes            |
| ------------ | ----------- | ------ | ---------------- |
| . | . |`np.fft.fft(a)` | Fourier transform of `a` |
| . | . |`np.fft.ifft(a)` | inverse Fourier transform of `a` |
| . | . |`signal.resample(x, np.ceil(len(x)/q))` |  downsample with low-pass filtering |

### TensorFS

The [[tensorfs]] data filesystem provides a global filesystem-like workspace for storing tensor data, and [[Goal]] has special commands and functions to facilitate interacting with it.

In an interactive `goal` shell, when you do `##` to switch into math mode, the prompt changes to show your current directory in the tensorfs, not the regular OS filesystem, and the final prompt character turns into a `#`.

Use `get` and `set` (aliases for `tensorfs.Get` and `tensorfs.Set`) to retrieve and store data in the tensorfs:

* `x := get("path/to/item")` retrieves the tensor data value at given path, which can then be used directly in an expression or saved to a new variable as in this example.

* `set("path/to/item", x)` saves tensor data to given path, overwriting any existing value for that item if it already exists, and creating a new one if not. `x` can be any data expression.

You can use the standard shell commands to navigate around the data filesystem:

* `cd <dir>` to change the current working directory. By default, new variables created in the shell are also recorded into the current working directory for later access.

* `ls [-l,r] [dir]` list the contents of a directory; without arguments, it shows the current directory. The `-l` option shows each element on a separate line with its shape. `-r` does a recursive list through subdirectories.

* `mkdir <dir>` makes a new subdirectory.

