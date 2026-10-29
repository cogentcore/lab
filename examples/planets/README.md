# Planets

Build and run this `main` package to see a full demo of how to use this system for data analysis, paralleling the example in [Python Data Science](https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html) using pandas, to see directly how that translates into this framework.

The terminal at the bottom runs the interactive `goal` shell, in the app, so you can work with the data directly there: for example, `cd Planets` and `ls` in math mode (`##`), or any Go code using `tensorfs`.

Most of the code is in the `AnalyzePlanets` function, which opens a .csv file, and then performs various analyses, using various `tensor/stats` functions.

The GUI is from the `tensor/databrowser`, showing various views of the data in the left browser. The `Describe` directory has summary stats of various columns of the overall data.by Click into `Stats` and explore from there, doing context menu, `Plot` or double-clicking to get a tabular representation of the data.

