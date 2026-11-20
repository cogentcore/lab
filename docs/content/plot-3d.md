+++
Categories = ["Plot"]
+++

3D [[plot]]s are planned: these are notes for them.

TODO: use math32 3D projection math and you can just take each 3d point and reduce to 2D. For stuff you want to actually be able to use in SVG, it needs to ultimately be 2D, so it makes sense to support basic versions here, including XYZ (points, lines), Bar3D, wireframe.

Could also have a separate plot3d package based on `xyz` that is true 3D for interactive 3D plots of surfaces or things that don't make sense in this more limited 2D world.
