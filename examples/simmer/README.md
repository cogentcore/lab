# Simmer: simulation running and monitoring infrastructure

"Watch your sims simmer on the GPU burners.."

Simmer provides a Cogent Lab based GUI for running simulation jobs, and managing and analyzing the resulting data, using the plotcore Editor. It supports both slurm and [baremetal](../baremetal) job management systems.

The local data goes into paths like this:

`~/simdata/simname/user` with `jobs` and `labscripts` subdirectories

The server data goes into similar paths but starting with `~/simserver` instead.

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

