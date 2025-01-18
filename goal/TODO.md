This is a list of all the stuff that doesn't work in goal.


## `compression-tutorial`

* len in `compress` function:

```Goal
    # sz := a.len // or len(a) or a.size or size(a)
    // note that above is a tensor.Int b/c everything is a tensor!
    # top := sorted[:n]
    # res := zeros(sz)  // sz here should absolutely work; doesn't
    for i := range sz { // this is very tricky and probably can't work.
    // but also a.Len() directly doesn't work either! seems like an issue
    // with the yaegi range expression?  dunno.
```


