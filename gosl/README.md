# gosl: Go as a shader language

`gosl` implements _Go as a shader language_ for GPU compute shaders (using [WebGPU](https://www.w3.org/TR/webgpu/)), **enabling standard Go code to run on the GPU**.

`gosl` converts Go code to WGSL which can then be loaded directly into a WebGPU compute shader, using the [core gpu](https://github.com/cogentcore/core/tree/main/gpu) compute shader system. It operates within the overall [Goal](../goal/README.md) framework of an augmented version of the Go language.

See [examples/basic](examples/basic) and [rand](examples/rand) for complete working examples.

See the [Cogent Lab Docs](https://cogentcore.org/lab/gosl) for full documentation.

