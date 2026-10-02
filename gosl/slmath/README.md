# slmath

`slmath` defines special math functions that operate on the [math32](https://github.com/cogentcore/core/tree/main/math32) vector, matrix and quaternion types. These must be called as functions, not methods, and be outside of math32 itself so that the `math32.Vector3` -> `vec3<f32>` replacement operates correctly. Must explicitly import this package into gosl using:

```go
	//gosl:import "cogentcore.org/lab/gosl/slmath"
```

WGSL has native support for the basic operations on vectors (`Add`, `Sub`, `Mul`, `Div`, `MulScalar`, `DivScalar`), which gosl transpiles into the equivalent operators, so there are no `slmath` functions for those. Everything else in math32 that makes sense on the GPU is here.

## Naming

Go has no overloading, so each type gets a distinct name:

| math32 type | slmath name | example |
| --- | --- | --- |
| `Vector2` | numeric `2` suffix | `slmath.Length2(v)` |
| `Vector3` | numeric `3` suffix | `slmath.Length3(v)` |
| `Vector4` | numeric `4` suffix | `slmath.Length4(v)` |
| `Quat` | `Quat` prefix | `slmath.QuatSlerp(q, o, t)` |
| `Matrix2` | `Mat2` prefix | `slmath.Mat2Inverse(m)` |
| `Matrix3` | `Mat3` prefix | `slmath.Mat3Determinant(m)` |
| `Matrix4` | `Mat4` prefix | `slmath.Mat4MulVector3(m, v)` |

Where a math32 method mutates through a pointer receiver, the slmath function takes and returns values instead: `math32.Quat.Normalize` becomes `slmath.QuatNormalize(q) math32.Quat`.

## What is not here

Methods that have no WGSL equivalent are omitted: `String`, `GenGoSet`, anything using `image.Point`, `fixed.Point26_6`, `[]float32` slices or returning an `error`; the CPU-side `Matrix4` camera methods (`LookAt`, `SetPerspective`, `SetFrustum`, `SetOrthographic`, and the Vk variants); and methods with multiple return values (`Decompose`, `Eigen`, `ExtractBasis`, `RandomTangents`), since WGSL functions return a single value. Functions that would need those can use pointer out-parameters, as `MulSpatialTransforms` does.

Where math32 silently produces a different result on the GPU, slmath makes the Go code explicit so that CPU and GPU agree:

- `Normal2`/`Normal3`/`Normal4` return the zero vector for a zero-length input. math32 gets this from the zero check inside `DivScalar`, which gosl transpiles to a bare `/`.
- `Mat2Inverse`/`Mat3Inverse`/`Mat4Inverse` return the identity matrix when the determinant is 0.
- `Mat4ScaleCols` scales the columns of the given matrix. `math32.Matrix4.ScaleCols` returns a zero matrix because it does not copy the source first; this matches `math32.Matrix4.SetScaleCols` and `math32.Matrix3.ScaleCols` instead.

## Matrices

`math32.Matrix3` and `math32.Matrix4` are flat Go arrays stored column-wise, so element `k` is column `k/N`, row `k%N`. They map to the WGSL `mat3x3f` and `mat4x4f` types, which are indexed `m[col][row]`, and gosl translates a flat Go index into that index pair. `math32.Matrix2` maps to `mat2x3f`, and gosl translates its named fields: `XX`, `XY`, `X0` are column 0 and `YX`, `YY`, `Y0` are column 1.

Because the WGSL constructors take their values column-wise, which is not the order of the `Matrix2` struct fields, matrices are built here with `Mat2Set` and `Mat4Set` rather than composite literals. `math32.Mat3` already takes its 9 values column-wise, so it is used directly.

The matrix methods on the math32 types mostly cannot be used in gosl code, so call the `slmath` functions instead. gosl translates the ones that are a plain WGSL operator (`Matrix3.Mul`, `Matrix3.MulVector3`, `Matrix3.MulScalar` and `Matrix4.Mul`) and reports an error for the rest.

`Mat2Mul`, `Mat3Mul` and `Mat4Mul` are all the standard product `a * b`: element (row i, column j) of the result is row i of `a` dotted with column j of `b`, the same as the WGSL `*` operator. Every `Mul` function multiplies the vector or point on the right, so a chain of them applies its transforms right to left, as in `math32.Matrix2`. `math32.Quat.Mul` is likewise an error, because the WGSL `*` on the `vec4<f32>` it maps to is a component-wise multiply, not a quaternion multiply: use `slmath.MulQuats`.

## Tests

`*_test.go` check every slmath function against the equivalent math32 method on the CPU. `gosl_test.go` runs gosl over `testdata/slmathtest.go`, which calls every slmath function, compares the generated WGSL against `testdata/SLMathTest.golden`, and has the WGSL compiler validate it. To update the golden after an intended change:

```sh
cd testdata && gosl && cp shaders/SLMathTest.wgsl SLMathTest.golden
```
