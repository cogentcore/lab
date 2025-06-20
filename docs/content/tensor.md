+++
Categories = ["Tensor"]
+++

The **tensor.Tensor** represents n-dimensional data of various types, providing similar functionality to the widely used [NumPy](https://numpy.org/doc/stable/index.html) libraries in Python, and the commercial MATLAB framework.

The [[Goal]] [[math]] mode operates on tensor data exclusively: see documentation there for convenient shortcut expressions for common tensor operations. This page documents the underlying Go language implementation of tensors. See [[doc:tensor]] for the basic Go API docs, and [[tensor math]] for basic math operations that can be performed on tensors.

A tensor can be constructed from a Go slice, and accessed using a 1D index into that slice:

```Goal
x := tensor.NewFromValues(0, 1, 2, 3)
val := x.Float1D(2)

core.NewText(b).SetText(fmt.Sprint(val))
```

Note that the type of the tensor is inferred from the values, using standard Go rules, so you would need to add a decimal to obtain floating-point numbers instead of `int`s:

```Goal
x := tensor.NewFromValues(0., 1., 2., 3.)
val := x.Float1D(2)

core.NewText(b).SetText(fmt.Sprintf("value: %v %T", val, val))
```

You can reshape the tensor by setting the number of values along any number of dimensions, preserving any values that are compatible with the new shape, and access values using n-dimensional indexes:

```Goal
x := tensor.NewFromValues(0, 1, 2, 3)
x.SetShapeSizes(2, 2)
val := x.Float(1, 0)

core.NewText(b).SetText(fmt.Sprint(val))
```

The dimensions are organized in _row major_ format (same as [NumPy](https://numpy.org/doc/stable/index.html)), so the number of rows comes first, then columns; the last dimension (i.e., columns in this case) is the _innermost_ dimension, so that each column represents a contiguous array of values in memory, while rows are _not_ contiguous.

You can create a tensor with a specified shape, and fill it with a single value:

```Goal
x := tensor.NewFloat32(2, 2)
tensor.SetAllFloat64(x, 1)

core.NewText(b).SetText(x.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

Note the detailed formatting available from the standard stringer `String()` method on any tensor, providing the shape sizes on the first line, with dimensional indexes for the values.

A given tensor can hold any standard Go value type, including `int`, `float32` and `float64`, and `string` values (using Go generics for the numerical types), and it provides accessor methods for the following "core" types:
* `Float` methods set and return `float64` values.
* `Int` methods set and return `int` values.
* `String` methods set and return `string` values.

For example, you can directly get a `string` representation of any value:

```Goal
x := tensor.NewFromValues(0, 1, 2, 3)
val := x.String1D(2)

core.NewText(b).SetText(val)
```

### Setting values

To set a value, you typically use a type-specific method most appropriate for the underlying data type:

```Goal
x := tensor.NewFloat32(2, 2)
tensor.SetAllFloat64(x, 1)

x.SetFloat(3.14, 0, 1) // value comes first, then the appropriate number of indexes as varargs...

core.NewText(b).SetText(x.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

There are also `Value`, `Value1D`, and `Set`, `Set1D` methods that use Generics to operate on the actual underlying data type:

```Goal
x := tensor.NewFloat32(2, 2)
tensor.SetAllFloat64(x, 1)

x.Set(3.1415, 0, 1)

val := x.Value(0, 1)
v1d := x.Value1D(1)

core.NewText(b).SetText(x.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
core.NewText(b).SetText(fmt.Sprintf("val: %v %T", val, val))
core.NewText(b).SetText(fmt.Sprintf("v1d: %v", v1d))
```


## Views and values

The abstract [[doc:tensor.Tensor]] interface is implemented (and extended) by the concrete [[doc:tensor.Values]] types, which are what we've been getting in the above examples, and directly manage an underlying Go slice of values. These can be reshaped and appended to, like a Go slice.

In addition, there are various _View_ types that wrap other tensors and provide more flexible ways of accessing the tensor values, and provide all of the same core functionality present in [NumPy](https://numpy.org/doc/stable/index.html).

### Sliced

First, this is the starting Values tensor, as a 3x4 matrix:

```Goal
x := tensor.NewFloat64(3, 4)
x.CopyFrom(tensor.NewIntRange(12))

core.NewText(b).SetText(x.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

Using the [[doc:tensor.Reslice]] function, you can extract any subset from this 2D matrix, for example the values in a given row or column:

```Goal
x := tensor.NewFloat64(3, 4)
x.CopyFrom(tensor.NewIntRange(12))
row1 := tensor.Reslice(x, 1) // row is first index; column index is unspecified = all
col1 := tensor.Reslice(x, tensor.FullAxis, 1) // explicitly request all rows

core.NewText(b).SetText("row1: " + row1.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
core.NewText(b).SetText("col1: " + col1.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

Note that the column values got turned into a 1D tensor in this process -- to keep it as a column vector, you need to add a "blank" row dimension, which can be done using the `tensor.NewAxis` value:

```Goal
x := tensor.NewFloat64(3, 4)
x.CopyFrom(tensor.NewIntRange(12))
col1 := tensor.Reslice(x, tensor.NewAxis, tensor.FullAxis, 1)

core.NewText(b).SetText("col1: " + col1.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

You can also specify sub-ranges along each dimension, or even reorder the values, by using a [[doc:tensor.Slice]] element that has `Start`, `Stop` and `Step` values, like those of a standard Go `for` loop expression, with sensible default behavior for zero values:

```Goal
x := tensor.NewFloat64(3, 4)
x.CopyFrom(tensor.NewIntRange(12))
col1 := tensor.Reslice(x, tensor.Slice{Step: -1}, 1)

core.NewText(b).SetText("col1: " + col1.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

You can use `tensor.Ellipsis` to specify `FullAxis` for all the dimensions up to those specified, to flexibly focus on the innermost dimensions:

```Goal
x := tensor.NewFloat64(3, 2, 2)
x.CopyFrom(tensor.NewIntRange(12))
last1 := tensor.Reslice(x, tensor.Ellipsis, 1)

core.NewText(b).SetText("values: " + x.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
core.NewText(b).SetText("last1: " + last1.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

As in [NumPy](https://numpy.org/doc/stable/index.html), the [[doc:tensor.Sliced]] view wraps the original source tensor, so that if you change a value in that original source, _the value automatically changes in the view_ as well. Use the `AsValues()` method on a view to get a new concrete [[doc:tensor.Values]] representation of the view (equivalent to the NumPy `copy` function).

```Goal
x := tensor.NewFloat64(3, 2, 2)
x.CopyFrom(tensor.NewIntRange(12))
last1 := tensor.Reslice(x, tensor.Ellipsis, 1)

core.NewText(b).SetText("values: " + x.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
core.NewText(b).SetText("last1: " + last1.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})

values := last1.AsValues()
x.Set(3.14, 1, 0, 1)

core.NewText(b).SetText("values: " + x.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
core.NewText(b).SetText("last1: " + last1.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
core.NewText(b).SetText("values: " + values.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

### Masked by boolean values

You can apply a boolean mask to a tensor, to extract arbitrary values where the boolean value is true:

```Goal
x := tensor.NewFloat64(3, 4)
x.CopyFrom(tensor.NewIntRange(12))

m := tensor.NewMasked(x).Filter(func(tsr tensor.Tensor, idx int) bool {
	return tsr.Float1D(idx) >= 6
})
vals := m.AsValues()

core.NewText(b).SetText("masked: " + m.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
core.NewText(b).SetText("vals: " + vals.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

Note that missing values are encoded as `NaN`, which allows the resulting [[doc:tensor.Masked]] view to retain the shape of the original, and all of the other math functions operating on tensors properly treat `NaN` as a missing value that is ignored. You can also get the concrete values as shown, but this reduces the shape to 1D by default.

### Arbirary indexed values

You can extract arbitrary values from a tensor using a list of indexes (as a tensor), where the shape of that list then determines the shape of the resulting view:

```Goal
x := tensor.NewFloat64(3, 4)
x.CopyFrom(tensor.NewIntRange(12))

ixs := tensor.NewIntFromValues(
	0, 1,
	0, 1,
	0, 2,
	0, 2,
	1, 1,
	1, 1,
	2, 2,
	2, 2)
ixs.SetShapeSizes(2,4,2) // note: last 2 is the number of indexes into source

ix := tensor.NewIndexed(x, ixs)

core.NewText(b).SetText(ix.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

You can also feed Masked indexes into the [[doc:tensor.Indexed]] view to get a reshaped view:

```Goal
x := tensor.NewFloat64(3, 4)
x.CopyFrom(tensor.NewIntRange(12))

m := tensor.NewMasked(x).Filter(func(tsr tensor.Tensor, idx int) bool {
	return tsr.Float1D(idx) >= 6
})
ixs := m.SourceIndexes(true)
ixs.SetShapeSizes(2,3,2)
ix := tensor.NewIndexed(x, ixs)

core.NewText(b).SetText("masked: " + ix.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

### Differences from NumPy

[NumPy](https://numpy.org/doc/stable/index.html) is somewhat confusing with respect to the distinction between _basic indexing_ (using a single index or sliced ranges of indexes along each dimension) versus _advanced indexing_ (using an array of indexes or bools). Basic indexing returns a _view_ into the original data (where changes to the view directly affect the underlying type), while advanced indexing returns a _copy_.

However, rather confusingly (per this [stack overflow question](https://stackoverflow.com/questions/15691740/does-assignment-with-advanced-indexing-copy-array-data)), you can do direct assignment through advanced indexing (more on this below):
```Python
a[np.array([1,2])] = 5  # or:
a[a > 0.5] = 1          # boolean advanced indexing
```

In the tensor package, all of the View types ([[doc:tensor.Sliced]], [[doc:tensor.Reshaped]], [[doc:tensor.Masked]], and [[doc:tensor.Indexed]]) are unambiguously wrappers around a source tensor, and their values change when the source changes. Use `.AsValues()` to break that connection and get the view as a new set of concrete values.

### Row, Cell access

The [[doc:tensor.RowMajor]] interface provides a convenient set of methods to access tensors where the first, outermost dimension is a row, and there may be multiple remaining dimensions after that. All concrete [[doc:tensor.Values]] tensors implement this interface.

For example, you can easily get a `SubSpace` tensor that contains the values within a given row, and set values within a row tensor using a flat 1D "cell" index that applies to the values within a row:

```Goal
x := tensor.NewFloat64(3, 2, 2)
x.CopyFrom(tensor.NewIntRange(12))

x.SetFloatRow(3.14, 1, 2) // set 1D cell 2 in row 1
row1 := x.RowTensor(1)

core.NewText(b).SetText("values: " + x.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
core.NewText(b).SetText("row1: " + row1.String()).Styler(func(s *styles.Style) {
    s.Text.WhiteSpace = text.WhiteSpacePre
    s.Font.Family = rich.Monospace
})
```

## Tensor pages

