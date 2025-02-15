# Bare Metal

Bare metal is a simple compute job manager for bare metal machines, with the manager running on a local "client" (laptop), using the goal language ssh facilities to connect to the servers and execute all the management through that connection, so that nothing needs to be installed on the server.

Jobs are submitted through an RPC connection (e.g., from `simrun`) also running on the local client typically.  The job itself consists of a gzipped tar file ("job blob") containing an executable script (chmod +x) that is run on a server, along with relevant metadata.

There is no attempt to prioritize jobs: it is just FIFO. The main function is just to manage a queue of jobs so that the compute resources are not overloaded, along with basic job monitoring for completion, canceling, etc.

Each job consumes one GPU, as key a simplification to minimize resource management complexity.

## Environment variables

* `BARE_GPU` = the allocated GPU number (0..N]

## job.* files

* `job.out` contains all the output from running the job script.
* `job.pid` has the pid process id of the job.
* `job.files.tar.gz` has the files submitted for the job.
* `job.results.tar.gz` has the results from the job

## Configuring a new "bare metal" linux compute server

```sh
sudo apt install golang gcc libgl1-mesa-dev libegl1-mesa-dev mesa-vulkan-drivers xorg-dev vulkan-tools nvidia-driver-565-server nvidia-utils-565-server
```

