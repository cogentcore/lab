+++
bibfile = "ccnlab.json"
+++

**Physics** is a 3D physics simulator for creating virtual environments, including simulated robots and animals, which can run on the GPU or CPU using [[GoSL]]. See [[doc:physics]] for the API docs. The [xyz](https://cogentcore.org/core/xyz) 3D visualization framework can be used to view the physics using the [[doc:physics/phyxyz]] package, including grabbing first-person views from the perspective of a body element within the virtual world. It is actively used for simulating motor learning in [axon](https://github.com/emergent/axon).

Physics is based on the design and algorithms from the [newton-physics](https://newton-physics.github.io/newton/guide/overview.html) framework developed by Disney Research, Google DeepMind, and NVIDIA, and implemented in the [NVIDIA Warp](https://nvidia.github.io/warp/basics.html) framework, which is conceptually similar to GoSL.

The XPBD (eXtended Position-Based Dynamics) solver is exclusively implemented ([[@MacklinMullerChentanez16]] and [[@MullerMacklinChentanezEtAl20]]), which has impressive capabilities as shown in this [YouTube video](https://www.youtube.com/watch?v=CPq87E1vD8k). The key idea is to avoid working in the world of higher-order derivatives (accelerations) and use robust position-based updates, as discussed in [[#Physics solver algorithms]].

Consistent with [xyz](https://cogentcore.org/core/xyz) and [gpu](https://cogentcore.org/core/gpu), the default coordinate system has `Y` as the up axis, and `Z` is the depth axis, consistent with the [USD](https://openusd.org/release/index.html) standard. Newton-physics uses `Z` up by default, which is the robotics standard.

## Examples

### Multi-link pendulum

The following example shows a complete simulation of a multi-link pendulum: 

{id="sim_pendula" title="N Pendula" collapsed="false"}
```Goal
ed := phyxyz.NewEditor(b)
ed.CameraPos = math32.Vec3(0, 3, 3)
ed.Styler(func(s *styles.Style) {
    s.Min.Y.Em(40)
})

params := struct{NPendula int}{}

params.NPendula = 2
ed.SetUserParams(&params)

ed.SetConfigFunc(func() {
	ml := ed.Model
	sc := ed.Scene
    hsz := math32.Vec3(0.05, .2, 0.05)
    mass := float32(0.1)
    stY := 4*hsz.Y
	x := -hsz.Y

    rleft := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), -math32.Pi/2)
    pb := sc.NewDynamic(ml, "top", physics.Capsule, "blue", mass, hsz, math32.Vec3(x, stY, 0), rleft)
	pb.SetBodyGroup(1) // no collide across groups
	ji := sc.NewJointRevolute(ml, nil, pb, math32.Vec3(0, stY, 0), math32.Vec3(0, hsz.Y, 0), math32.Vec3(0, 0, 1))
	physics.SetJointTargetPos(ji, 0, 0, 0)
	physics.SetJointTargetVel(ji, 0, 0, 0)

	for i := 1; i < params.NPendula; i++ {
		clr := colors.Names[i%len(colors.Names)]
		x = -float32(i)*hsz.Y*2 - hsz.Y
		cb := sc.NewDynamic(ml, "child", physics.Capsule, clr, mass, hsz, math32.Vec3(x, stY, 0), rleft)
		cb.SetBodyGroup(1+i)
		ji = sc.NewJointRevolute(ml, pb, cb, math32.Vec3(0, -hsz.Y, 0), math32.Vec3(0, hsz.Y, 0), math32.Vec3(0, 0, 1))
		physics.SetJointTargetPos(ji, 0, 0, 0)
		physics.SetJointTargetVel(ji, 0, 0, 0)
		pb = cb
    }
})
```

The [[doc:physics/phyxyz/Editor]] widget provides the [[doc:physics/Model]] and [[doc:physics/phyxyz/Scene]] elements, and the `ConfigFunc` function that configures the physics elements. Stepping through these elements in order:

```go
    rleft := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), -math32.Pi/2)
    pb := sc.NewDynamic(ml, "top", physics.Box, "blue", mass, hsz, math32.Vec3(x, stY, 0), rleft)
```

The `math32.Quat` quaternion provides all the rotational math used in `xyz` and `physics`, and the `rleft` instance represents a -90 degree rotation about the Z (depth) axis, which is what causes the pendulum to start in a horizontal orientation.

The `NewDynamic` method adds a new dynamic body element with a default visualization (this is a `phyxyz` wrapper around the same method in the `physics` package). Dynamic elements are updated by the physics engine, while `NewBody` would create a static rigid body element that doesn't move (unless you specifically change its position). The return value is a [[doc:physics/phyxyz/Skin]] which provides the visualization of a physics body. It uses the `BodyIndex` of the body to get updated values.

```go
	pb.SetBodyGroup(1) // no collide across groups
```

The `Group` property of a body can be set to fine-tune collision logic. Positive-numbered groups only collide with each other and any negative-numbered groups, while negative-numbered groups only collide with positive numbered and not within their own group. 0 means it doesn't collide with anything. With the crazy dynamics that emerge with multiple arms, it is good to let them all pass through each other.

```go
	ji := sc.NewJointRevolute(ml, nil, pb, math32.Vec3(0, stY, 0), math32.Vec3(0, hsz.Y, 0), math32.Vec3(0, 0, 1))
	physics.SetJointTargetPos(ji, 0, 0, 0)
	physics.SetJointTargetVel(ji, 0, 0, 0)
```

This creates a new joint where the `pb` body is the child and `nil` means the "world" is the parent (i.e., it is just anchored in a fixed world location). We specify the parent and child relative positions for this joint, which is relative to each such body (in the case of a world joint, it is in world coordinates). Note that these are the _unrotated_ positions, so we are specifying the _vertical_ (Y) axis offset here for the child (`pb`) body. This means that the joint is positioned at the top of the body, because all sizes are specified as _half_ sizes (like a radius instead of a diameter).

The next two lines specify the target position and velocity of this joint, along with a _stiffness_ (position) and _damping_ (velocity) parameter, which indicate how _strongly_ to enforce these constraints. The values of 0 here indicate that they are not enforced at all, because we want the links to swing freely. You can try using positive values there and see what happens!

The remaining code just does this same kind of thing for the further links, and should be relatively clear.

### Joint control

{id="sim_prismatic" title="Prismatic Joint" collapsed="true"}
```Goal
ed := phyxyz.NewEditor(b)
ed.CameraPos = math32.Vec3(0, 10, 10)
ed.Styler(func(s *styles.Style) {
    s.Min.Y.Em(40)
})

ed.SetConfigFunc(func() {
	ml := ed.Model
	sc := ed.Scene
    hsz := math32.Vec3(1, 2, 0.5)
    mass := float32(0.1)
    
    obj := sc.NewDynamic(ml, "body", physics.Box, "blue", mass, hsz, math32.Vec3(0, hsz.Y, 0), math32.NewQuatIdentity())
	ji := sc.NewJointPrismatic(ml, nil, obj, math32.Vec3(-5, 0, 0), math32.Vec3(0, hsz.Y, 0), math32.Vec3(1, 0, 0))
})

// variables to control
pos := float32(1)
stiff := float32(10)
vel := float32(0)
damp := float32(10)

var posStr, stiffStr, velStr, dampStr string

ed.SetControlFunc(func(timeStep int) {
	physics.SetJointTargetPos(0, 0, pos, stiff)
	physics.SetJointTargetVel(0, 0, vel, damp)
})

func update() {
    posStr = fmt.Sprintf("Pos: %g", pos)
    stiffStr = fmt.Sprintf("Stiff: %g", stiff)
    velStr = fmt.Sprintf("Vel: %g", vel)
    dampStr = fmt.Sprintf("Damp: %g", damp)
}

update()

func addSlider(label *string, val *float32, maxVal float32) {
    tx := core.NewText(b)
    tx.Styler(func(s *styles.Style) {
        s.Min.X.Ch(40)  // clean rendering with variable width content
    })
	core.Bind(label, tx)
	sld := core.NewSlider(b).SetMin(0).SetMax(maxVal).SetStep(1).SetEnforceStep(true)
	sld.SendChangeOnInput()
	sld.OnChange(func(e events.Event) {
		update()
		tx.UpdateRender()
	})
	core.Bind(val, sld)
}

addSlider(&posStr, &pos, 10)
addSlider(&stiffStr, &stiff, 1000)
addSlider(&velStr, &vel, 2)
addSlider(&dampStr, &damp, 1000)
```

This simulation allows interactive control over the parameters of a `Prismatic` joint, which sets the linear position of a body along a given axis, in this case along the horizontal (`X`) axis.

Click the `Step 10000` button and then start moving the sliders to see the effects interactively. Here's what you should observe:

* `Stiff` (stiffness) determines how quickly the joint responds to the position changes. You can make this variable even stronger in practice (e.g., 10,000).

* `Damp` (damping) opposes `Stiff` in resisting changes, but some amount of damping is essential to prevent oscillations (definitely try Damp = 0). In general a value above 20 or so seems to be necessary for preventing significant oscillations.

{id="sim_ball" title="Ball Joint" collapsed="true"}
```Goal
ed := phyxyz.NewEditor(b)
ed.CameraPos = math32.Vec3(0, 10, 10)
ed.Styler(func(s *styles.Style) {
    s.Min.Y.Em(40)
})

ed.SetConfigFunc(func() {
	ml := ed.Model
	sc := ed.Scene
    hsz := math32.Vec3(0.5, 1.5, 0.2)
    mass := float32(0.1)
    
    obj := sc.NewDynamic(ml, "body", physics.Box, "blue", mass, hsz, math32.Vec3(0, hsz.Y, 0), math32.NewQuatIdentity())
	ji := sc.NewJointBall(ml, nil, obj, math32.Vec3(0, 0, 0), math32.Vec3(0, -hsz.Y, 0))
})

// variables to control
posX := float32(0)
posY := float32(0)
posZ := float32(0)
stiff := float32(500) // note: higher values can get unstable for large angles
damp := float32(20)

var posXstr, posYstr, posZstr, stiffStr, dampStr string

ed.SetControlFunc(func(timeStep int) {
	physics.SetJointTargetAngle(0, 0, posX, stiff)
	physics.SetJointTargetAngle(0, 1, posY, stiff)
	physics.SetJointTargetAngle(0, 2, posZ, stiff)
	physics.SetJointTargetVel(0, 0, 0, damp)
	physics.SetJointTargetVel(0, 1, 0, damp)
	physics.SetJointTargetVel(0, 2, 0, damp)
})

func update() {
    posXstr = fmt.Sprintf("Pos X: %g", posX)
    posYstr = fmt.Sprintf("Pos Y: %g", posY)
    posZstr = fmt.Sprintf("Pos Z: %g", posZ)
    stiffStr = fmt.Sprintf("Stiff: %g", stiff)
    dampStr = fmt.Sprintf("Damp: %g", damp)
}

update()

func addSlider(label *string, val *float32, minVal, maxVal float32) {
    tx := core.NewText(b)
    tx.Styler(func(s *styles.Style) {
        s.Min.X.Ch(40)  // clean rendering with variable width content
    })
	core.Bind(label, tx)
	sld := core.NewSlider(b).SetMin(minVal).SetMax(maxVal).SetStep(1).SetEnforceStep(true)
	sld.SendChangeOnInput()
	sld.OnChange(func(e events.Event) {
		update()
		tx.UpdateRender()
	})
	core.Bind(val, sld)
}

addSlider(&posXstr, &posX, -179, 179)
addSlider(&posYstr, &posY, -179, 179)
addSlider(&posZstr, &posZ, -179, 179)
addSlider(&stiffStr, &stiff, 0, 1000)
addSlider(&dampStr, &damp, 0, 1000)
```

The above `Ball` joint example demonstrates a 3 angular degrees-of-freedom joint, using the `SetJointTargetAngle` function that takes degrees as input, and automatically wraps the values in the -180..180 degree (-PI..PI) range, which is the natural range of position values for angular joints.

You can see that the control can become a bit unstable at extreme angles and angle combinations. Increasing damping and reducing stiffness can help in these situations.

## GoSL infrastructure

As discussed in [[GoSL]], to run equivalent code on the GPU and the CPU (i.e., standard Go), all of the data needs to be represented in large arrays, implemented via [[tensor]]s, and all processing occurs via _parallel for loops_ that effectively process each element of these data arrays in parallel. Enum types are used to define the variables as the last inner-most dimension in the data tensors, e.g., [[doc:physics.BodyVars]], with accessor functions to get the relevant `math32` types (e.g., `math32.Vector3`) across the X,Y,Z components.

## Bodies and Dynamics

The basic element is a _body_, which is a rigid physical entity with a specific shape, mass, position and orientation. Call [[doc:physics.Model]] `NewBody` to create a new one. There are (currently) only standard geometric [[#shapes]] available (arbitrary triangular meshes and soft bodies could be supported as needed in the future, based on existing newton-physics code).

By itself, a body is static. To make a body that is subject to forces and can be connected to other bodies via [[#joints]], use `NewDynamic`, which creates an additional set of data to implement the dynamic equations of the physics solver. The initial position and orientation of a dynamic body can be restored via the `InitState` method.

To optimize the collision detection computation, it is important to organize bodies into `World` and `Group` elements:

* World: Use different world indexes for separate collections of bodies that only interact amongst themselves, and global bodies that have a -1 index. By default everything goes in world = 0. See [[#parallel worlds]] for more info.

* Group: by default this is set to -1 for all static bodies (non-dynamic), which can interact with any dynamic body, but not with any other static body, and to 1 for all dynamic bodies, which can interact with each other and static bodies. To make dynamic bodies that don't interact, assign them increasing group numbers.

There is also a special constraint where the parent and child on a same joint do not collide, as this often happens and would lead to weird behavior.

There is also an `Object` index for each body, that is used for external manipulation and control purposes, but does not affect collision or physics.

## Shapes

The elemental shapes are a `Plane`, `Sphere`, `Capsule`, `Cylinder`, `Cone`, and `Box`: [[doc:physics.Shapes]]. The `Size` property on bodies is always the _half_ size, such as the radius or the half-height of a cylinder or capsule. This is used in `newton-physics` and makes more sense for center-based computations: physics operates on the center-of-mass of a body. Consistent with the overall coordinate system, the `Cylinder` and `Capsule` are oriented with `Y` as the height dimension, which is unfortunately inconsistent with the Z=up convention in `newton-physics`.

### Multi-shape bodies

The newton-physics framework, and MuJoCo upon which it is based, support multiple shapes per body, which can then be integrated to produce an aggregate inertia. This adds an additional level of complexity and management overhead, which we are currently avoiding in favor of putting the shapes directly on the body, so each body has 1 and only 1 shape. This simplifies collision considerably as well. It would not be difficult to add a shape layer at some point in the future. The same goes for Mesh, SDF, and HeightField types.

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

## Phyxyz viewer

Typically, bodies are created using the enhanced functions in the [[doc:physics/phyxyz]] package, which provides a [[doc:physics/phyxyz.View]] wrapper for physics bodies. This wrapper has a default `Color` setting to provide simple color coding of bodies, and supports `NewView` and `InitView` functions that allow arbitrary visualization dynamics to be associated with each body (textures, meshes, dynamic updating etc).

## Parallel worlds

TODO: switch over to builder here.

The compute efficiency of the GPU goes up with the more elements that are processed in parallel, amortizing the memory transfer overhead and leveraging the parallel cores. Furthermore, in AI applications for example, models can be trained in parallel on different instances of the same environment, with each instance having its own random initial starting point and trajectory over time. All of these instances can be simulated in one `physics.Model` by using the `World` index on the bodies, with the shared static environment living in World -1, and the elements of each instance (e.g., a simulated robot) living in its own separate world.

The `NewBody` and `NewDynamic` methods automatically use the `Model.CurrentWorld` index by default, or you can directly use `SetBodyWorld` to assign a specific world index.

The `ReplicateWorld` method creates N replicas of an existing world, including all associated joints. This can only be called once, as it records the start and N-per-world of each such replicated world, which allows the `phyxyz` viewer to efficiently view a specific world. Thus, under this scenario, you create world 0 and then replicate it, then modify the initial positions and orientations accordingly, using `PositionObject`, as described next. The object numbers are also replicated so uniquely indexing a specific object instance requires specifying the world and object indexes.

The phyxyz viewer can display a specific world, or all worlds.

## Manipulating objects

TODO.

## Sensors

TODO.

## Physics solver algorithms

This section provides a brief overview of different physics solver algorithms, and motivates why we're using XPBD (see [[@MacklinMullerChentanez16]] and [[@MullerMacklinChentanezEtAl20]] for full info). See [[@CollinsChandVanderkopEtAl21]] for a recent review of relevant software and approaches. There are two main categories of mathematical problems that these engines solve:

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

