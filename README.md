<h1 align="center">
    <a href="https://cogentcore.org/lab">
        <img alt="Cogent Core Logo" src="icon.svg"><br>
    </a>
</h1>

<p align="center">
    <a href="https://goreportcard.com/report/cogentcore.org/lab"><img src="https://goreportcard.com/badge/cogentcore.org/lab" alt="Go Report Card"></a>
    <a href="https://pkg.go.dev/cogentcore.org/lab"><img src="https://img.shields.io/badge/dev-reference-007d9c?logo=go&logoColor=white&style=flat" alt="pkg.go.dev docs"></a>
    <a href="https://github.com/cogentcore/lab/actions"><img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/cogentcore/lab/go.yml"></a>
    <a href="https://raw.githack.com/wiki/cogentcore/lab/coverage.html"><img alt="Test Coverage" src="https://github.com/cogentcore/lab/wiki/coverage.svg"></a>
    <a href="https://github.com/cogentcore/lab/tags"><img alt="Version" src="https://img.shields.io/github/v/tag/cogentcore/lab?label=version"></a>
</p>

Cogent Lab is a free and open source data science and visualization framework for the Go language, built on top of the [Cogent Core](https://cogentcore.org/core) GUI framework, so that the same code runs on macOS, Windows, Linux, iOS, Android, and the web. Features include:

* The `goal` language transpiler that supports more concise math and matrix expressions that are largely compatible with the widely-used NumPy framework, in addition to shell command syntax, so it can be used as a replacement for a command-line shell. 

* A `tensor` representation for n-dimensional data, which serves as the universal data type within the Lab framework. The `table` uses tensors as columns for tabular, heterogenous data (similar to the widely-used pandas data table), and the `tensorfs` is a hierarchical filesystem for tensor data that serves as the shared data workspace.

* Interactive, full-featured plots and other GUI visualization tools.

* The overarching `lab` API for flexibly connecting data and visualization components, providing the foundation for interactive data analysis applications integrating different Cogent Lab elements.

* The `gosl` _Go shader language_ system that allows you to write Go (and `goal`) functions that run on either the CPU or the GPU, using the WebGPU framework that supports full GPU compute functionality through the web browser.

See the [Cogent Lab Website](https://cogentcore.org/lab) for more information, including extensive documentation and editable interactive running examples. 

