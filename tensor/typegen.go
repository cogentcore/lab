// Code generated by "core generate"; DO NOT EDIT.

package tensor

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/tensor.Indexed", IDName: "indexed", Doc: "Indexed provides an arbitrarily indexed view onto another \"source\" [Tensor]\nwith each index value providing a full n-dimensional index into the source.\nThe shape of this view is determined by the shape of the [Indexed.Indexes]\ntensor up to the final innermost dimension, which holds the index values.\nThus the innermost dimension size of the indexes is equal to the number\nof dimensions in the source tensor. Given the essential role of the\nindexes in this view, it is not usable without the indexes.\nThis view is not memory-contiguous and does not support the [RowMajor]\ninterface or efficient access to inner-dimensional subspaces.\nTo produce a new concrete [Values] that has raw data actually\norganized according to the indexed order (i.e., the copy function\nof numpy), call [Indexed.AsValues].", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Tensor", Doc: "Tensor source that we are an indexed view onto."}, {Name: "Indexes", Doc: "Indexes is the list of indexes into the source tensor,\nwith the innermost dimension providing the index values\n(size = number of dimensions in the source tensor), and\nthe remaining outer dimensions determine the shape\nof this [Indexed] tensor view."}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/tensor.Masked", IDName: "masked", Doc: "Masked is a filtering wrapper around another \"source\" [Tensor],\nthat provides a bit-masked view onto the Tensor defined by a [Bool] [Values]\ntensor with a matching shape. If the bool mask has a 'false'\nthen the corresponding value cannot be Set, and Float access returns\nNaN indicating missing data (other type access returns the zero value).\nA new Masked view defaults to a full transparent view of the source tensor.\nTo produce a new [Values] tensor with only the 'true' cases,\n(i.e., the copy function of numpy), call [Masked.AsValues].", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Tensor", Doc: "Tensor source that we are a masked view onto."}, {Name: "Mask", Doc: "Bool tensor with same shape as source tensor, providing mask."}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/tensor.Reshaped", IDName: "reshaped", Doc: "Reshaped is a reshaping wrapper around another \"source\" [Tensor],\nthat provides a length-preserving reshaped view onto the source Tensor.\nReshaping by adding new size=1 dimensions (via [NewAxis] value) is\noften important for properly aligning two tensors in a computationally\ncompatible manner; see the [AlignShapes] function.\n[Reshaped.AsValues] on this view returns a new [Values] with the view\nshape, calling [Clone] on the source tensor to get the values.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Tensor", Doc: "Tensor source that we are a masked view onto."}, {Name: "Reshape", Doc: "Reshape is the effective shape we use for access.\nThis must have the same Len() as the source Tensor."}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/tensor.Rows", IDName: "rows", Doc: "Rows is a row-indexed wrapper view around a [Values] [Tensor] that allows\narbitrary row-wise ordering and filtering according to the [Rows.Indexes].\nSorting and filtering a tensor along this outermost row dimension only\nrequires updating the indexes while leaving the underlying Tensor alone.\nUnlike the more general [Sliced] view, Rows maintains memory contiguity\nfor the inner dimensions (\"cells\") within each row, and supports the [RowMajor]\ninterface, with the [Set]FloatRow[Cell] methods providing efficient access.\nUse [Rows.AsValues] to obtain a concrete [Values] representation with the\ncurrent row sorting.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Methods: []types.Method{{Name: "Sequential", Doc: "Sequential sets Indexes to nil, resulting in sequential row-wise access into tensor.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "ExcludeMissing", Doc: "ExcludeMissing deletes indexes where the values are missing, as indicated by NaN.\nUses first cell of higher dimensional data.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "FilterString", Doc: "FilterString filters the indexes using string values compared to given\nstring. Includes rows with matching values unless the Exclude option is set.\nIf Contains option is set, it only checks if row contains string;\nif IgnoreCase, ignores case, otherwise filtering is case sensitive.\nUses first cell of higher dimensional data.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"str", "opts"}}, {Name: "addRowsIndexes", Doc: "addRowsIndexes adds n rows to indexes starting at end of current tensor size", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"n"}}, {Name: "AddRows", Doc: "AddRows adds n rows to end of underlying Tensor, and to the indexes in this view", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"n"}}}, Fields: []types.Field{{Name: "Tensor", Doc: "Tensor source that we are an indexed view onto.\nNote that this must be a concrete [Values] tensor, to enable efficient\n[RowMajor] access and subspace functions."}, {Name: "Indexes", Doc: "Indexes are the indexes into Tensor rows, with nil = sequential.\nOnly set if order is different from default sequential order.\nUse the [Rows.RowIndex] method for nil-aware logic."}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/tensor.FilterOptions", IDName: "filter-options", Doc: "FilterOptions are options to a Filter function\ndetermining how the string filter value is used for matching.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Exclude", Doc: "Exclude means to exclude matches,\nwith the default (false) being to include"}, {Name: "Contains", Doc: "Contains means the string only needs to contain the target string,\nwith the default (false) requiring a complete match to entire string."}, {Name: "IgnoreCase", Doc: "IgnoreCase means that differences in case are ignored in comparing strings,\nwith the default (false) using case."}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/tensor.Sliced", IDName: "sliced", Doc: "Sliced provides a re-sliced view onto another \"source\" [Tensor],\ndefined by a set of [Sliced.Indexes] for each dimension (must have\nat least 1 index per dimension to avoid a null view).\nThus, each dimension can be transformed in arbitrary ways relative\nto the original tensor (filtered subsets, reversals, sorting, etc).\nThis view is not memory-contiguous and does not support the [RowMajor]\ninterface or efficient access to inner-dimensional subspaces.\nA new Sliced view defaults to a full transparent view of the source tensor.\nThere is additional cost for every access operation associated with the\nindexed indirection, and access is always via the full n-dimensional indexes.\nSee also [Rows] for a version that only indexes the outermost row dimension,\nwhich is much more efficient for this common use-case, and does support [RowMajor].\nTo produce a new concrete [Values] that has raw data actually organized according\nto the indexed order (i.e., the copy function of numpy), call [Sliced.AsValues].", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Methods: []types.Method{{Name: "Sequential", Doc: "Sequential sets all Indexes to nil, resulting in full sequential access into tensor.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}}, Fields: []types.Field{{Name: "Tensor", Doc: "Tensor source that we are an indexed view onto."}, {Name: "Indexes", Doc: "Indexes are the indexes for each dimension, with dimensions as the outer\nslice (enforced to be the same length as the NumDims of the source Tensor),\nand a list of dimension index values (within range of DimSize(d)).\nA nil list of indexes for a dimension automatically provides a full,\nsequential view of that dimension."}}})
