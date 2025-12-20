# Physics engine for virtual reality

The `physics` engine is a 3D physics simulator for creating virtual environments, which can run on the GPU or CPU using [GoSL](https://cogentcore.org/lab/gosl).

See [physics docs](https://cogentcore.org/lab/physics) for the main docs.

The [world](world) visualization sub-package manages a `View` element that links to physics bodies and generates an [xyz](https://cogentcore.org/core/xyz) 3D scenegraph based on the physics bodies, and updates this visualization efficiently as the physics is updated.

## TODO

* Muscles: https://mujoco.readthedocs.io/en/stable/modeling.html#muscles

