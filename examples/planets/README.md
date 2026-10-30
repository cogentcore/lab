# Planets

Build and run this `main` package to see a full demo of how to use this system for data analysis, paralleling the example in [Python Data Science](https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html) using pandas, to see directly how that translates into this framework.

The terminal at the bottom runs the interactive `goal` shell, in the app, so you can work with the data directly there: for example, `cd Planets` and `ls` in math mode (`##`), or any Go code using `tensorfs`.

Most of the code is in the `AnalyzePlanets` function, which opens a .csv file, and then performs various analyses, using various `tensor/stats` functions.

The GUI is from the `lab/Browser`, showing various views of the data in the left browser. The `Describe` directory has summary stats of various columns of the overall data.by Click into `Stats` and explore from there, doing context menu, `Plot` or double-clicking to get a tabular representation of the data.

## Tutorial

The GUI shows a [tensorfs](../tensorfs) **file tree** on the left with **directories** of data, and a set of **tabs** to the right, where you can find this README file and a couple of plots of the data, and the interactive **terminal** shell on the bottom.

### Raw Data

Double-click on the `Data` directory in the file tree. This shows the raw data from the `.csv` file, using a [table](../table) viewer.

The `DataNoNA` shows this data with all the `NaN` missing values (NA = not available) filtered out: any row with a `NaN` is removed, using the efficient index-based filtering in a table.

### Describe

The `Describe` directory has the descriptive stats for each of the different columns of data, from the NA filtered data. This matches the results shown in the tutorial.

### Groups and Stats

The `Stats` directory has the data organized into `Groups` according to the `method` and `decade` columns, with the same descriptive stats as shown in Describe, now applied to these grouped data.

For example, double-click on `Stats/method/orbital_period` and you can see a table of the summary stats by method. You can also see all the stats by double-clicking on `Stats/method` or `Stats/decade`, which are what is shown in the two bar plots that come up at startup.

## Terminial

Here's a transcript of something you can do in the terminal, using the **math mode** activated by typing '##':

```
~/go/src/cogentcore.org/lab/examples/planets/ > ##
data # cd Planets/Data
data/Planets/Data # ls
method number orbital_period mass distance year
data/Planets/Data # stats.Mean(get("orbital_period"))
2002.917596
data/Planets/Data #
```

