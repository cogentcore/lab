# Simmer: simulation running and monitoring infrastructure

"Watch your sims simmer on the GPU burners.."

Simmer provides a Cogent Lab based GUI for running simulation jobs, and managing and analyzing the resulting data, using the plotcore Editor. It supports both slurm and [baremetal](../baremetal) job management systems.

# Server types

## Slurm

The slurm backend uses array jobs to distribute computation over many CPU nodes. See `slurm.go` for the relevant code.

## Bare metal

If not using slurm, then the [baremetal](../baremetal) system manages jobs itself, maintaining a record of the jobs launched relative to the resources available.


