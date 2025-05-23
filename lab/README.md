# lab

The lab package provides GUI elements for data exploration and visualization, and a simple `Browser` implementation that combines these elements.

* `FileTree` (with `FileNode` elements), implementing a [filetree](https://github.com/cogentcore/tree/main/filetree) that has support for a [tensorfs](../tensorfs) filesystem, and data files in an actual filesystem. It has a `Tabber` pointer that handles the viewing actions on `tensorfs` elements (showing a Plot, etc).

* `Tabber` interface and `Tabs` base implementation provides methods for showing data plots and editors in tabs.

* `Terminal` running a `goal` shell that supports interactive commands operating on the `tensorfs` data etc. TODO!

* `Browser` provides a hub structure connecting the above elements, which can be included in an actual GUI widget, that also provides additional functionality / GUI elements.

The basic `Browser` puts the `FileTree` in a left `Splits` and the `Tabs` in the right, and supports interactive exploration and visualization of data. See the [basic](examples/basic) example for a simple instance.

In the [emergent](https://github.com/emer) framework, these elements are combined with other GUI elements to provide a full neural network simulation environment on top of the databrowser foundation.

