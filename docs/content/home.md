+++
URL = ""
Title = "Cogent Lab"
+++

Cogent Lab is a free and open source data science and visualization framework for the Go language, built on top of the [Cogent Core](https://cogentcore.org/core) GUI framework, so that the same code runs on macOS, Windows, Linux, iOS, Android, and web.

Features include:

* The `goal` language transpiler (generates standard Go code) that supports more concise math and matrix expressions that are largely compatible with the widely used numpy framework, in addition to shell command syntax, so it can be used as a replacement for a command-line shell.

* A `tensor` representation for n-dimensional data, which serves as the universal data type within the Lab framework. The `table` uses tensors as columns for tabular, heterogenous data (similar to the widely-used pandas data table), and the `tensorfs` is a hierarchical filesystem for tensor data that serves as the shared data workspace.

* Interactive, full-featured plots and other GUI visualization tools.

* The overarching `lab` API for flexibly connecting data and visualization components, providing the foundation for interactive data analysis applications integrating different Cogent Lab elements.

* The `gosl` _Go shader language_ system that allows you to write Go (and `goal`) functions that run on either the CPU or the GPU, using the WebGPU framework that supports full GPU compute functionality through the web browser.
