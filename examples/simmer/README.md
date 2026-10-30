# Simmer: simulation running and monitoring infrastructure

TODO: plot mean function on consat is losing the TrainEpoch fields

"Watch your sims simmer on the GPU burners.."

Simmer provides a Cogent Lab based GUI for running simulation jobs, and managing and analyzing the resulting data, using the plotcore Editor. It supports both slurm and [baremetal](../baremetal) job management systems.

The terminal at the bottom runs the interactive `goal` shell, in the app, with the data and the Simmer itself (as `lab.LabBrowser`), so you can work with them directly there.

# Projects

Simmer works on one simulation project at a time, which you open from the `Open recent` menu, or the list of recent projects that it shows when it starts without one, or with `Open`. Each project opens in its own simmer app, so you can run the app from `/Applications` and open the projects of different simulations. Running `simmer` in the directory of a project opens it.

The settings of a project are in the `simmer.json` file in the directory of the simulation, which `Edit project` edits and saves. They include the `Server.SSH` user@host of the server, which is connected as `@1` in goal commands, and connected again after the network is lost.

`New project` makes a new project for the simulation in a given directory: the local data goes into paths like this:

`~/simdata/simname/user` with `jobs` and `labscripts` subdirectories

with a `simdata` symbolic link to it in the directory of the simulation, and a `simmer.json` file.

The server data goes into similar paths but starting with `~/simserver` instead.

A project made before `simmer.json` has its settings moved to it from its `labscripts/defaults.goal` script when it is first opened, including the server from its `gossh` command. The script is then renamed to `defaults.goal.old`.

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


