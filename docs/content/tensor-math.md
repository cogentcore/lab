+++
Categories = ["Tensor"]
+++

The [[doc:tensor/tmath]] package implements most of the standard library [math](https://pkg.go.dev/math) functions, including basic arithmetic operations, for [[tensor]]s.

For example:

```Goal
x := tensor.NewFromValues(0., 1., 2., 3.)
add := tmath.Add(x, x)
sub := tmath.Sub(x, x)
mul := tmath.Mul(x, x)
div := tmath.Div(x, x)

fmt.Println("add:", add)
fmt.Println("sub:", sub)
fmt.Println("mul:", mul)
fmt.Println("div:", div)
```

As you can see, the operations are performed element-wise; see the [[matrix]] package for 2D matrix multiplication and related operations.

Math functions can be performed:

```Goal
x := tensor.NewFromValues(0., 1., 2., 3.)
sin := tmath.Sin(x)
atan := tmath.Atan2(x, tensor.NewFromValues(3.0))
pow := tmath.Pow(x, tensor.NewFromValues(2.0))

fmt.Println("sin:", sin)
fmt.Println("atan:", atan)
fmt.Println("pow:", pow)
```

See the info below on [[#Alignment of shapes]] for the rules governing the way that different-shaped tensors are aligned for these computations.

Parallel goroutines will be used for implementing these computations if the tensors are sufficiently large to make it generally beneficial to do so.

There are also `*Out` versions of each function, which take an additional output tensor to store the results into, instead of creating a new one. For computationally-intensive pipelines, it can be significantly more efficient to re-use pre-allocated outputs (which are automatically and efficiently resized to the proper capacity if not already).

## Alignment of shapes

The NumPy concept of [broadcasting](https://numpy.org/doc/stable/user/basics.broadcasting.html) is critical for flexibly defining the semantics for how functions taking two n-dimensional Tensor arguments behave when they have different shapes. Ultimately, the computation operates by iterating over the length of the longest tensor, and the question is how to _align_ the shapes so that a meaningful computation results from this.

If both tensors are 1D and the same length, then a simple matched iteration over both can take place. However, the broadcasting logic defines what happens when there is a systematic relationship between the two, enabling powerful (but sometimes difficult to understand) computations to be specified.

The following examples demonstrate the logic:

Innermost dimensions that match in dimension are iterated over as you'd expect:
```
Image  (3d array): 256 x 256 x 3
Scale  (1d array):             3
Result (3d array): 256 x 256 x 3
```

Anything with a dimension size of 1 (a "singleton") will match against any other sized dimension:
```
A      (4d array):  8 x 1 x 6 x 1
B      (3d array):      7 x 1 x 5
Result (4d array):  8 x 7 x 6 x 5
```
In the innermost dimension here, the single value in A acts like a "scalar" in relationship to the 5 values in B along that same dimension, operating on each one in turn. Likewise for the singleton second-to-last dimension in B.

Any non-1 mismatch represents an error:
```
A      (2d array):      2 x 1
B      (3d array):  8 x 4 x 3 # second from last dimensions mismatched
```

The `AlignShapes` function performs this shape alignment logic, and the `WrapIndex1D` function is used to compute a 1D index into a given shape, based on the total output shape sizes, wrapping any singleton dimensions around as needed. These are used in the [tmath](tmath) package for example to implement the basic binary math operators.

