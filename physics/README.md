# Physics engine for virtual reality

The `physics` engine is a 3D physics simulator for creating virtual environments, which can run on the GPU or CPU using [GoSL](https://cogentcore.org/lab/gosl).

See [physics docs](https://cogentcore.org/lab/physics) for the main docs.

The [phyxyz](phyxyz) ("physics") visualization sub-package manages a `Skin` element that links to physics bodies and generates an [xyz](https://cogentcore.org/core/xyz) 3D scenegraph based on the physics bodies, and updates this visualization efficiently as the physics is updated. There is an `Editor` widget that makes it easy to explore physics Models.

The [builder](builder) package provides a structured, hierarchical description of a  `physics.Model` that supports replicating worlds for parallel world execution, and easier manipulation of objects as collections of bodies (e.g., an entire object can be moved and re-oriented in one call).

## TODO

* sphere-sphere collision definitely not right: sometimes too much and sometimes not at all. do all the tests..

* pendula: if starting in vertical with 4+ links, it gets unstable when target pos is at 0, even with 0 stiff!

* Muscles: https://mujoco.readthedocs.io/en/stable/modeling.html#muscles

* fix basic issues in restitution: needs a more thorough approach. Basically need to integrate during entire time of penetration and then compute escape velocity based on the saved incoming velocity just prior to impact.

