# slmath

`slmath` defines special math functions that operate on vector and quaternion types. These must be called as functions, not methods, and be outside of math32 itself so that the `math32.Vector3` -> `vec3<f32>` replacement operates correctly. Must explicitly import this package into gosl using:

```go
	//gosl:import "cogentcore.org/lab/gosl/slmath"
```

