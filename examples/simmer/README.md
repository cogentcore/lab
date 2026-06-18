# Simmer: simulation running and monitoring infrastructure

"Watch your sims simmer on the GPU burners.."

Simmer provides a Cogent Lab based GUI for running simulation jobs, and managing and analyzing the resulting data, using the plotcore Editor. It supports both slurm and [baremetal](../baremetal) job management systems.

The local data goes into paths like this:

`~/simdata/simname/user` with `jobs` and `labscripts` subdirectories

The server data goes into similar paths but starting with `~/simserver` instead.

The `install.goal` script makes the paths and a shortcut in given sim:
```sh
> ./install.goal ~/full/path/to/sim
```

# Server types

## Slurm

The slurm backend uses array jobs to distribute computation over many CPU nodes. See `slurm.go` for the relevant code.

## Bare metal

If not using slurm, then the [baremetal](../baremetal) system manages jobs itself, maintaining a record of the jobs launched relative to the resources available.

baremetal can be run on localhost to manage jobs on your own local machine, using this config:

```toml
[[Servers.Values]]
	Name = "lh"
	SSH = "localhost"
	NGPUs = 4
```

# Workflow

* `Jobs` shows current jobs in a Jobs tab.

* `Submit` runs a new sim job. Use `Message` for general, more stable info, and `Label` for a short unique identifier for this specific job.

* `Status` gets status of any running jobs.  Anything done running gets status of `Finalized` and is no longer updated by Status.  All job metadata is downloaded from host, _but not the Results_ output data, which is `Fetch`ed separately because it may be large and often needs to be consolidated in a particular way, because multiple runs of a job are executed in parallel.

* `Fetch` gets result `.tsv` files from server. It can be run on running or Finalized jobs.  When run on Finalized, then the status is set to `Fetched` and it is automatically skipped in any future Fetch actions.

* `Results` grabs specific result data files into a `Results` tab, from which further examination and plotting occurs. This step is necessary because there are typically multiple different types of results files, so you need to select which type you want view.

* `Plot` plots combined data across any selected files in `Results` tab, allowing you to compare them, using the `JobLabel` as a legend so each job has its own line color.

* `Diff` shows a diff browser for any two selected Jobs, or one selected job vs. the current sim working directory -- key for tracking changes made across jobs.


