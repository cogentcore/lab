This is a list of all the stuff that doesn't work in goal.

## docs issues

* for loop max in metric.md doesn't work..
* matrix.EigSym return 2 values not working -- only 1

## converting between go / tensor

* `for range` expression that deals with iterating over tensors, in math mode

* CCN temporal-derivative.md: 

```
    for t := range #totalTime[0]# { // or something even simpler?
```

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

```Goal
```


## simrun / baremetal

### baremetal/jobs.goal:
```Goal
    var errs []error // triggers some kind of interp state switch
    // so that subsequent parsing is whack
``` 

### simrun/jobs.goal:

```Goal
    [cd {spath} && /bin/rm -rf {jid}] // && gets captured as &
```

* also possibility of using `&` as a separator like `;` in shell
* also for ssh context you want to pass `;` and `&` etc through to ssh side
but not when in a local connection: some kind of escaping or something?


