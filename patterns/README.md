# Patterns

This package contains functions that generate n-dimensional patterns (in tensors) based on various algorithms, typically for use as inputs to neural network models or other such learning systems. It also has some routines for helping manage collections of such patterns.

In general the [tensorfs](../tensorfs) system is used to manage a "vocabulary" of such patterns. The `tensor.RowMajor` API is used to organize a list (rows) of patterns.

## Permuted Binary and FlipBits

The `PermutedBinary*` functions create binary patterns with a specific number of "on" vs. "off" bits, which can be useful for enforcing a target level of activity.

The `FlipBits*` functions preserve any existing activity levels while randomly flipping a specific number of bits on or off.

## Mixing patterns

The `Mix` function acts a bit like a multi-track mixer, combining different streams of patterns together in a higher-dimensional composite pattern.

## Managing rows

Some misc functions help managing rows of data:

* `SplitRows`: split out subsets of a larger list.
* `ReplicateRows`: replicate multiple copies of a given row.
* `Shuffle`: permuted order of rows.

## Random seed

A separate random number source can be established, using the [randx](../base/randx) package.

## Usage examples

### Permuted Binary

```Go
	a := dir.Float32("A", 6, 3, 3) // 6 rows of 3x3 patterns
	nOn := patterns.NFromPct(0.3, 9) // 30% activity
	nDiff := patterns.NFromPct(0.4, nOn) // 40% max overlap
	patterns.PermutedBinaryMinDiff(a, nOn, 1, 0, nDiff) // ensures minimum distance
```

### Replicate, assemble, and split rows

```Go
	ctx1 := dir.Float32("ctxt1")
	patterns.ReplicateRows(ctx1, a.SubSpace(0), 6) // 6x first row of 'a' above
	ab := dir.Float32("ab", 0, 3, 3)
	ab.AppendFrom(a) // add a patterns
	ab.AppendFrom(b) // add b patterns
```

```Go
	// split 12 items into 3 sets of 4
	patterns.SplitRows(dir, ab, []string{"as", "bs", "cs"}, 3, 3) 
```

### Mix patterns

```Go
	mix := dir.Float32("mix")
	patterns.Mix(mix, 12, a, b, ctx1, ctx1, empty, b) // make 12 rows from given sources
	mix.SetShapeSizes(12, 3, 2, 3, 3) // reshape to 3x2 = "outer" dims x 3x3 inner
```


