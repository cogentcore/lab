+++
Categories = ["Table"]
+++

**table** provides a DataTable / DataFrame structure similar to [pandas](https://pandas.pydata.org/) and [xarray](http://xarray.pydata.org/en/stable/) in Python, and [Apache Arrow Table](https://github.com/apache/arrow/tree/master/go/arrow/array/table.go), using [[tensor]] n-dimensional columns aligned by common outermost row dimension.

Data in the table is accessed by first getting the `Column` tensor (typically by name), and then using the [[doc:tensor.RowMajor]] methods to access data within that tensor in a row-wise manner:

```Goal
dt := table.New()
dt.AddStringColumn("Name")
dt.AddFloat64Column("Data", 2, 2)
dt.SetNumRows(3)

dt.Column("Name").SetStringRow("item0", 0, 0)
dt.Column("Name").SetStringRow("item1", 1, 0)
dt.Column("Name").SetStringRow("item2", 2, 0)

dt.Column("Data").SetFloatRow(55, 0, 0)
dt.Column("Data").SetFloatRow(102, 1, 1) // note: last arg is 1D "cell" index
dt.Column("Data").SetFloatRow(37, 2, 3)

val := dt.Column("Data").FloatRow(2, 3)

fmt.Println(dt)
fmt.Printf("val: %v\n", val)
```

## Sorting and filtering

The `Column` method creates a [[doc:tensor.Rows]] for the underlying column values, with a list of indexes used for the row-level access, which enables efficient sorting and filtering by row, as only these indexes need to be updated, not the underlying data values. The indexes are maintained on the table, which provides an indexed view onto the underlying data values that are stored in a separate [[doc:table.Columns]] structure. Thus, there can be multiple different such table views onto the same underlying columns data.

```Goal
dt := table.New()
dt.AddStringColumn("Name")
dt.AddFloat64Column("Data")
dt.SetNumRows(3)

fruits := []string{"peach", "apple", "orange"}

for i := range 3 {
	dt.Column("Name").SetStringRow(fruits[i], i, 0)
	dt.Column("Data").SetFloatRow(float64(i+1), i, 0)
}

dt.Sequential()
dt.SortColumn("Data", tensor.Descending)
fmt.Println(dt)

dt.Sequential()
dt.Filter(func(dt *table.Table, row int) bool {
	return dt.Column("Data").FloatRow(row, 0) > 1
})
fmt.Println(dt)
```

## CSV / TSV file format

Tables can be saved and loaded from CSV (comma separated values) or TSV (tab separated values) files.  See the next section for special formatting of header strings in these files to record the type and tensor cell shapes.

### Type and Tensor Headers

To capture the type and shape of the columns, we support the following header formatting.  We weren't able to find any other widely supported standard (please let us know if there is one that we've missed!)

Here is the mapping of special header prefix characters to standard types:
```go
'$': etensor.STRING,
'%': etensor.FLOAT32,
'#': etensor.FLOAT64,
'|': etensor.INT64,
'@': etensor.UINT8,
'^': etensor.BOOL,
```

Columns that have tensor cell shapes (not just scalars) are marked as such with the *first* such column having a `<ndim:dim,dim..>` suffix indicating the shape of the *cells* in this column, e.g., `<2:5,4>` indicates a 2D cell Y=5,X=4.  Each individual column is then indexed as `[ndims:x,y..]` e.g., the first would be `[2:0,0]`, then `[2:0,1]` etc.

### Example

Here's a TSV file for a scalar String column (`Name`), a 2D 1x4 tensor float32 column (`Input`), and a 2D 1x2 float32 `Output` column.

```
_H:	$Name	%Input[2:0,0]<2:1,4>	%Input[2:0,1]	%Input[2:0,2]	%Input[2:0,3]	%Output[2:0,0]<2:1,2>	%Output[2:0,1]
_D:	Event_0	1	0	0	0	1	0
_D:	Event_1	0	1	0	0	1	0
_D:	Event_2	0	0	1	0	0	1
_D:	Event_3	0	0	0	1	0	1
```


