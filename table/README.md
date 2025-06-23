# table 

**table** provides a DataTable / DataFrame structure similar to [pandas](https://pandas.pydata.org/) and [xarray](http://xarray.pydata.org/en/stable/) in Python, and [Apache Arrow Table](https://github.com/apache/arrow/tree/master/go/arrow/array/table.go), using [tensor](../tensor) n-dimensional columns aligned by common outermost row dimension.

See the [Cogent Lab Docs](https://cogentcore.org/lab/table) for full documentation.

See [examples/dataproc](examples/dataproc) for a demo of how to use this system for data analysis, paralleling the example in [Python Data Science](https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html) using pandas, to see directly how that translates into this framework.

Whereas an individual `Tensor` can only hold one data type, the `Table` allows coordinated storage and processing of heterogeneous data types, aligned by the outermost row dimension. The main `tensor` data processing functions are defined on the individual tensors (which are the universal computational element in the `tensor` system), but the coordinated row-wise indexing in the table is important for sorting or filtering a collection of data in the same way, and grouping data by a common set of "splits" for data analysis.  Plotting is also driven by the table, with one column providing a shared X axis for the rest of the columns.

The `Table` mainly provides "infrastructure" methods for adding tensor columns and CSV (comma separated values, and related tab separated values, TSV) file reading and writing.  Any function that can be performed on an individual column should be done using the `tensor.Rows` and `Tensor` methods directly.

As a general convention, it is safest, clearest, and quite fast to access columns by name instead of index (there is a `map` from name to index), so the base access method names generally take a column name argument, and those that take a column index have an `Index` suffix.

The table itself stores raw data `tensor.Tensor` values, and the `Column` (by name) and `ColumnByIndex` methods return a `tensor.Rows` with the `Indexes` pointing to the shared table-wide `Indexes` (which can be `nil` if standard sequential order is being used).  

If you call Sort, Filter or other routines on an individual column tensor, then you can grab the updated indexes via the `IndexesFromTensor` method so that they apply to the entire table.  The `SortColumn` and `FilterString` methods do this for you.

There are also multi-column `Sort` and `Filter` methods on the Table itself.

It is very low-cost to create a new View of an existing Table, via `NewView`, as they can share the underlying `Columns` data.

