# Simmer: simulation running and monitoring infrastructure

"Watch your sims simmer on the GPU burners.."

# Server types

## Slurm

The slurm backend uses array jobs to distribute computation over many CPU nodes. See `slurm.go` for the relevant code.

## Bare metal

If not using slurm, then the [baremetal](../baremetal) system manages jobs itself, maintaining a record of the jobs launched relative to the resources available.

### Configuring a new "bare metal" linux compute server

```sh
sudo apt install golang gcc libgl1-mesa-dev libegl1-mesa-dev mesa-vulkan-drivers xorg-dev vulkan-tools nvidia-driver-565-server nvidia-utils-565-server
```

