# GoSL: Go as a shader language

**GoSL** (via the `gosl` executable) allows you to write Go programs that run on [[GPU]] hardware, by transpiling Go into the WGSL shader language used by [WebGPU](https://www.w3.org/TR/webgpu/), thereby establishing the _Go shader language_.

GoSL uses the [core gpu](https://github.com/cogentcore/core/tree/main/gpu) compute shader system, and can take advantage of the [[Goal]] transpiler to provide a more natural tensor indexing syntax.

Functionally, GoSL is similar to [NVIDIA warp](https://github.com/NVIDIA/warp) --  [docs](https://nvidia.github.io/warp/basics.html), which uses python as the original source and converts it to either C++ or CUDA code. In GoSL, the original code is directly Go, so we just need to do the WGSL part. Unlike warp, WGSL runs on all GPU platforms, including the web (warp only runs on NVIDIA GPUs, on desktop).

See [examples/basic](examples/basic) and [rand](examples/rand) for complete working examples.

See the [Cogent Lab Docs](https://cogentcore.org/lab/gosl) for full documentation.

