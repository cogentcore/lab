// Code generated by "core generate"; DO NOT EDIT.

package table

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/table.Table", IDName: "table", Doc: "Table is a table of Tensor columns aligned by a common outermost row dimension.\nUse the [Table.Column] (by name) and [Table.ColumnIndex] methods to obtain a\n[tensor.Rows] view of the column, using the shared [Table.Indexes] of the Table.\nThus, a coordinated sorting and filtered view of the column data is automatically\navailable for any of the tensor package functions that use [tensor.Tensor] as the one\ncommon data representation for all operations.\nTensor Columns are always raw value types and support SubSpace operations on cells.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Methods: []types.Method{{Name: "Sequential", Doc: "Sequential sets Indexes to nil, resulting in sequential row-wise access into tensor.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "SortColumn", Doc: "SortColumn sorts the indexes into our Table according to values in\ngiven column, using either ascending or descending order,\n(use [tensor.Ascending] or [tensor.Descending] for self-documentation).\nUses first cell of higher dimensional data.\nReturns error if column name not found.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"columnName", "ascending"}, Returns: []string{"error"}}, {Name: "SortColumns", Doc: "SortColumns sorts the indexes into our Table according to values in\ngiven column names, using either ascending or descending order,\n(use [tensor.Ascending] or [tensor.Descending] for self-documentation,\nand optionally using a stable sort.\nUses first cell of higher dimensional data.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"ascending", "stable", "columns"}}, {Name: "FilterString", Doc: "FilterString filters the indexes using string values in column compared to given\nstring. Includes rows with matching values unless the Exclude option is set.\nIf Contains option is set, it only checks if row contains string;\nif IgnoreCase, ignores case, otherwise filtering is case sensitive.\nUses first cell from higher dimensions.\nReturns error if column name not found.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"columnName", "str", "opts"}, Returns: []string{"error"}}, {Name: "SaveCSV", Doc: "SaveCSV writes a table to a comma-separated-values (CSV) file\n(where comma = any delimiter, specified in the delim arg).\nIf headers = true then generate column headers that capture the type\nand tensor cell geometry of the columns, enabling full reloading\nof exactly the same table format and data (recommended).\nOtherwise, only the data is written.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename", "delim", "headers"}, Returns: []string{"error"}}, {Name: "OpenCSV", Doc: "OpenCSV reads a table from a comma-separated-values (CSV) file\n(where comma = any delimiter, specified in the delim arg),\nusing the Go standard encoding/csv reader conforming to the official CSV standard.\nIf the table does not currently have any columns, the first row of the file\nis assumed to be headers, and columns are constructed therefrom.\nIf the file was saved from table with headers, then these have full configuration\ninformation for tensor type and dimensionality.\nIf the table DOES have existing columns, then those are used robustly\nfor whatever information fits from each row of the file.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename", "delim"}, Returns: []string{"error"}}, {Name: "AddRows", Doc: "AddRows adds n rows to end of underlying Table, and to the indexes in this view.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"n"}, Returns: []string{"Table"}}, {Name: "SetNumRows", Doc: "SetNumRows sets the number of rows in the table, across all columns.\nIf rows = 0 then effective number of rows in tensors is 1, as this dim cannot be 0.\nIf indexes are in place and rows are added, indexes for the new rows are added.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"rows"}, Returns: []string{"Table"}}}, Fields: []types.Field{{Name: "Columns", Doc: "Columns has the list of column tensor data for this table.\nDifferent tables can provide different indexed views onto the same Columns."}, {Name: "Indexes", Doc: "Indexes are the indexes into Tensor rows, with nil = sequential.\nOnly set if order is different from default sequential order.\nThese indexes are shared into the `tensor.Rows` Column values\nto provide a coordinated indexed view into the underlying data."}, {Name: "Meta", Doc: "Meta is misc metadata for the table. Use lower-case key names\nfollowing the struct tag convention:\n\t- name string = name of table\n\t- doc string = documentation, description\n\t- read-only bool = gui is read-only\n\t- precision int = n for precision to write out floats in csv."}}})