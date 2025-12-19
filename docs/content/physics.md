+++
bibfile = "ccnlab.json"
+++

**Physics** is a 3D physics simulator for creating virtual environments, including simulated robots and animals, which can run on the GPU or CPU using [[GoSL]]. See [[doc:physics]] for the API docs. The [xyz](https://cogentcore.org/core/xyz) 3D visualization framework can be used to view the physics using the [[doc:physics/world]] package, including grabbing first-person views from the perspective of a body element within the virtual world. It is actively used for simulating motor learning in [axon](https://github.com/emergent/axon).

It is based on the design and algorithms from the [newton-physics](https://newton-physics.github.io/newton/guide/overview.html) framework developed by Disney Research, Google DeepMind, and NVIDIA, and implemented in the [NVIDIA Warp](https://nvidia.github.io/warp/basics.html) framework, which is conceptually similar to GoSL.

The XPBD (eXtended Position-Based Dynamics) solver is exclusively implemented ([[@MacklinMullerChentanez16]] and [[@MullerMacklinChentanezEtAl20]]), which has impressive capabilities as shown in this [YouTube video](https://www.youtube.com/watch?v=CPq87E1vD8k). The key idea is to avoid working in the world of higher-order derivatives (accelerations) and use robust position-based updates, as discussed in [[#Physics solver algorithms]].

Consistent with [xyz](https://cogentcore.org/core/xyz) and [gpu](https://cogentcore.org/core/gpu), the default coordinate system has `Y` as the up axis, and `Z` is the depth axis, consistent with the [USD](https://openusd.org/release/index.html) standard. Newton-physics uses `Z` up by default, which is the robotics standard.

## GoSL infrastructure

As discussed in [[GoSL]], to run equivalent code on the GPU and the CPU (i.e., standard Go), all of the data needs to be represented in large arrays, implemented via [[tensor]]s, and all processing occurs via _parallel for loops_ that effectively process each element of these data arrays in parallel. Enum types are used to define the variables as the last inner-most dimension in the data tensors, e.g., [[doc:physics.BodyVars]], with accessor functions to get the relevant `math32` types (e.g., `math32.Vector3`) across the X,Y,Z components.

## Bodies and Dynamics

The basic element is a _body_, which is a rigid physical entity with a specific shape, mass, position and orientation. Call [[doc:physics.World]] `NewBody` to create a new one. There are (currently) only standard geometric [[#shapes]] available (arbitrary triangular meshes and soft bodies could be supported as needed in the future, based on existing newton-physics code).

By itself, a body is static. To make a body that is subject to forces and can be connected to other bodies via [[#joints]], use `NewDynamic`, which creates and additional set of data to implement the dynamic equations of the physics solver. The initial position and orientation of the body can be restored via the `InitState` method.

To optimize the collision detection computation, it is important to organize bodies into `World` and `Group` elements:
* World: Use different world indexes for separate collections of bodies that only interact amongst themselves, and global bodies that have a -1 index. By default everything goes in world = 0.
* Group: typically just use -1 for all static bodies (non-dynamic), which can interact with any dynamic body, but not with any other static body. And use 1 for all dynamic bodies, which can interact with each-other and static bodies.

There is a special constraint where the parent and child on a same joint do not collide, as this often happens and would lead to weird behavior.

## Shapes

The elemental shapes are a `Box`, `Sphere`, `Cylinder` (Cone if one radius is 0), and `Capsule`: [[doc:physics.Shapes]]. The `Size` property on bodies is always the _half_ size, such as the radius or the half-height of a cylinder or capsule. This is used in `newton-physics` and makes more sense for center-based computations: physics operates on the center-of-mass of a body. Consistent with the overall coordinate system, the `Cylinder` and `Capsule` are oriented with `Y` as the height dimension, which is unfortunately inconsistent with the Z=up convention in `newton-physics`.

## Joints

The supported [[doc:physics.JointTypes]] include the following (DoF = degrees-of-freedom, names are based on standards in mechanical engineering and robotics):

* `Prismatic` Prismatic allows translation along a single axis (i.e., _slider_): 1 DoF.

* `Revolute` allows rotation about a single axis (axel): 1 DoF.

* `Ball` allows rotation about all three axes (3 DoF).

* `Fixed` locks all relative motion: 0 DoF.

* `Free` allows full 6-DoF motion (translation and rotation).

* `Distance` keeps two bodies a distance within joint limits: 6 DoF.

* `D6` is a generic 6-DoF joint that can be configured with up to 3 linear DoF and 3 angular DoF.

Use `NewJoint*` with _dynamic_ body indexes to create joints (e.g., `NewJointPrismatic` etc). Each joint can be positioned with a relative offset and orientation relative to the _parent_ and _child_ elements. The parent index can be set to -1 to anchor a child body in an arbitrary and fixed position within the overall world.

## World viewer

Typically, bodies are created using the enhanced functions in the [[doc:physics/world]] package, which provides a [[doc:physics/world.View]] wrapper for physics bodies. This wrapper has a default `Color` setting to provide simple color coding of bodies, and supports `NewView` and `InitView` functions that allow arbitrary visualization dynamics to be associated with each body (textures, meshes, dynamic updating etc).

## Physics solver algorithms

This section provides a brief overview of different physics solver algorithms, and motivates why we're using XPBD (see [[@MacklinMullerChentanez16]] and [[@MullerMacklinChentanezEtAl20]] for full info). There are two main categories of mathematical problems that these engines solve:

* Impacts from contact / collisions among bodies. When two billiard balls hit each other, they rebound in an _elastic_ collision, for example. There are also forces of friction and graded levels of inelasticity in these dynamics. The primary problem here is that the instantaneous forces involved in these impacts can be huge (this is why objects tend to shatter when you drop them on a hard surface), because the momentum reverses within a very short period of time. Numerical integration techniques tend to perform poorly when dealing with such huge forces and resulting accelerations.

* Integrating the effects of multiple joints connecting rigid bodies. Managing the multiple constraints that arise from a _chain_ of interconnected rigid body objects is particularly challenging, because each element in the chain has impacts on the other elements, as illustrated even with two such objects in the [double pendulum](https://en.wikipedia.org/wiki/Double_pendulum). A _naive_ explicit approach using standard Newtonian physics equations incurs exponentially costly computational cost, so some kind of more sophisticated approach is required for real-time simulation.

For the first problem, the general approach is to summarize the overall effect of the impact at a more abstract level, in terms of net _velocities_, instead of simulating the actual forces and accelerations which get very large and unwieldy. This is what [[@^Mirtich96]] developed in his **impulse-based** approach to impacts.

For the second problem, [[@^Featherstone83]] developed a **reduced coordinates** approach (also known as _generalized_ coordinates) that uses a complex set of mathematical equations to capture exactly the effective degrees of freedom in the whole chain. This requires detailed information about each element in the chain, and also requires sequential evaluation from the end of the chain up to its root.

The widely-used [bullet physics](https://github.com/bulletphysics/bullet3) combines these two approaches as described in [[@^Mirtich96]], and provides a fast and relatively robust solution. However, the C++ based codebase has evolved many times over the years and is very difficult to understand. Furthermore, the complexity of the Featherstone algorithm makes it a formidible challenge to implement.

Another widely-used approach to the joint-chain problem is based on introducing additional soft _constraints_ (technically known as _Lagrange multipliers_) that can iteratively distribute the forces in a way that can be computed in linear time, as developed by [[@^Baraff96]]. This is known as a **maximal coordinates** approach, because each body is represented using their standard position and orientation coordinates as if they were independent freely-moving objects. This approach was used in the widely-used [ODE (Open Dynamics Engine)](https://ode.org/) package. A drawback of this approach is that there can be non-physical gaps that emerge over time between bodies, as these soft constraints work their way through the system.

In this context, the **position based dynamics (PBD)** approach ([[@NealenMullerKeiserEtAl06]]) takes the idea from the impulse-based approach (using velocities instead of accelerations) to the "next level" and goes straight to positions, skipping even velocities! It uses an implicit iterative integration method known as [Gauss-Sidel](https://en.wikipedia.org/wiki/Gauss%E2%80%93Seidel_method) to integrate forces into resulting changes in position, with soft constraint factors that allow this integration method to rapidly converge.

The result is a fully _consistent_ updated set of positions for the objects that avoids the kinds of gaps that emerge in the Lagrange Multiplier approach. The approach is very fast and robust, and also has the distinct advantage of being relatively simple to implement, especially in a GPU-compatible parallel manner. Although the approach was developed for the even more challenging soft-body physics of deformable materials including cloth, it also provides robust solutions to the basic multi-joint rigid-body scenario.

The XPBD solver that we implement ([[@MacklinMullerChentanez16]] and [[@MullerMacklinChentanezEtAl20]]) fixes a few important problems with the PBD approach, so that the same results are obtained regardless of the time step used, and physically accurate forces and velocities can be back-computed from the final integrated position updates, so applications that track these factors can now be used. Overall, it appears to be the most robust solver that can use relatively large step sizes and a fully parallel implementation for high performance. The main downside is a potential loss in precise physical accuracy, but in most situations this is minimal, and the advantages overall should strongly outweigh this disadvantage.

Furthermore, the [newton-physics](https://github.com/newton-physics/newton) code for XPBD was very directly convertible to Go and GoSL (unlike the situation with bullet), so the overall process was relatively straightforward.

