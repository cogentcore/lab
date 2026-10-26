This is a list of all the stuff that doesn't work in goal.

## simrun/jobs.goal:

```Goal
    [cd {spath} && /bin/rm -rf {jid}] // && gets captured as &
```

* also possibility of using `&` as a separator like `;` in shell
* also for ssh context you want to pass `;` and `&` etc through to ssh side
but not when in a local connection: some kind of escaping or something?


