// Code generated by "core generate"; DO NOT EDIT.

package table

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/table.IndexView", IDName: "index-view", Doc: "IndexView is an indexed wrapper around an table.Table that provides a\nspecific view onto the Table defined by the set of indexes.\nThis provides an efficient way of sorting and filtering a table by only\nupdating the indexes while doing nothing to the Table itself.\nTo produce a table that has data actually organized according to the\nindexed order, call the NewTable method.\nIndexView views on a table can also be organized together as Splits\nof the table rows, e.g., by grouping values along a given column.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Methods: []types.Method{{Name: "Sequential", Doc: "Sequential sets indexes to sequential row-wise indexes into table", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "SortColumnName", Doc: "SortColumnName sorts the indexes into our Table according to values in\ngiven column name, using either ascending or descending order.\nOnly valid for 1-dimensional columns.\nReturns error if column name not found.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"column", "ascending"}, Returns: []string{"error"}}, {Name: "FilterColumnName", Doc: "FilterColumnName filters the indexes into our Table according to values in\ngiven column name, using string representation of column values.\nIncludes rows with matching values unless exclude is set.\nIf contains, only checks if row contains string; if ignoreCase, ignores case.\nUse named args for greater clarity.\nOnly valid for 1-dimensional columns.\nReturns error if column name not found.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"column", "str", "exclude", "contains", "ignoreCase"}, Returns: []string{"error"}}, {Name: "AddRows", Doc: "AddRows adds n rows to end of underlying Table, and to the indexes in this view", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"n"}}, {Name: "SaveCSV", Doc: "SaveCSV writes a table index view to a comma-separated-values (CSV) file\n(where comma = any delimiter, specified in the delim arg).\nIf headers = true then generate column headers that capture the type\nand tensor cell geometry of the columns, enabling full reloading\nof exactly the same table format and data (recommended).\nOtherwise, only the data is written.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename", "delim", "headers"}, Returns: []string{"error"}}, {Name: "OpenCSV", Doc: "OpenCSV reads a table idx view from a comma-separated-values (CSV) file\n(where comma = any delimiter, specified in the delim arg),\nusing the Go standard encoding/csv reader conforming to the official CSV standard.\nIf the table does not currently have any columns, the first row of the file\nis assumed to be headers, and columns are constructed therefrom.\nIf the file was saved from table with headers, then these have full configuration\ninformation for tensor type and dimensionality.\nIf the table DOES have existing columns, then those are used robustly\nfor whatever information fits from each row of the file.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename", "delim"}, Returns: []string{"error"}}}, Fields: []types.Field{{Name: "Table", Doc: "Table that we are an indexed view onto"}, {Name: "Indexes", Doc: "current indexes into Table"}, {Name: "lessFunc", Doc: "current Less function used in sorting"}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/table.Table", IDName: "table", Doc: "Table is a table of data, with columns of tensors,\neach with the same number of Rows (outer-most dimension).", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Methods: []types.Method{{Name: "SaveCSV", Doc: "SaveCSV writes a table to a comma-separated-values (CSV) file\n(where comma = any delimiter, specified in the delim arg).\nIf headers = true then generate column headers that capture the type\nand tensor cell geometry of the columns, enabling full reloading\nof exactly the same table format and data (recommended).\nOtherwise, only the data is written.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename", "delim", "headers"}, Returns: []string{"error"}}, {Name: "OpenCSV", Doc: "OpenCSV reads a table from a comma-separated-values (CSV) file\n(where comma = any delimiter, specified in the delim arg),\nusing the Go standard encoding/csv reader conforming to the official CSV standard.\nIf the table does not currently have any columns, the first row of the file\nis assumed to be headers, and columns are constructed therefrom.\nIf the file was saved from table with headers, then these have full configuration\ninformation for tensor type and dimensionality.\nIf the table DOES have existing columns, then those are used robustly\nfor whatever information fits from each row of the file.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename", "delim"}, Returns: []string{"error"}}, {Name: "AddRows", Doc: "AddRows adds n rows to each of the columns", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"n"}}, {Name: "SetNumRows", Doc: "SetNumRows sets the number of rows in the table, across all columns\nif rows = 0 then effective number of rows in tensors is 1, as this dim cannot be 0", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"rows"}, Returns: []string{"Table"}}}, Fields: []types.Field{{Name: "Columns", Doc: "columns of data, as tensor.Tensor tensors"}, {Name: "ColumnNames", Doc: "the names of the columns"}, {Name: "Rows", Doc: "number of rows, which is enforced to be the size of the outer-most dimension of the column tensors"}, {Name: "ColumnNameMap", Doc: "the map of column names to column numbers"}, {Name: "MetaData", Doc: "misc meta data for the table.  We use lower-case key names following the struct tag convention:  name = name of table; desc = description; read-only = gui is read-only; precision = n for precision to write out floats in csv.  For Column-specific data, we look for ColumnName: prefix, specifically ColumnName:desc = description of the column contents, which is shown as tooltip in the tensorcore.Table, and :width for width of a column"}}})
