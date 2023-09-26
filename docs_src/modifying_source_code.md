# Modifying source code

When modifying source code in the Automated Vending reference implementation, Docker images need to be rebuilt and services need to be updated to run newly built images. This document contains the steps for accomplishing this.

## Assumptions

This document assumes the Automated Vending services are already running. Additionally, it assumes you've already made a code change and saved the changes.

## Building the service's new image

Once the code change is saved, proceed to build the service's image. In this example, assume that the `ds-card-reader` service's source code has been altered.

Start by navigating to the root of this repository:

```bash
cd <repository_root>
```

Next, build the specific service's image:

```bash
make ds-card-reader
```

After Docker builds the image (by executing the steps in [`ds-card-reader/Dockerfile`](https://github.com/intel-retail/automated-vending/blob/Edgex-3.0/ds-card-reader/Dockerfile)), proceed to the next section.

## Remove and update the running service

One of the most effective methods of updating a Docker compose service is to remove the running container, and then re-run the `make` commands to bring up the entire Automated Vending reference implementation stack.

First, identify the running container for the service (again, `ds-card-reader` in this example):

```bash
docker ps | grep -i ds-card-reader
```

Using the output from the previous command, remove the container by referring to either its name or ID:

```bash
docker rm -f <name_or_ID_from_previous_command_output>
```

Once the container has been removed, bring up the entire stack using the `Makefile` command corresponding to your environment, which is **one** of the following commands.

For a standard simulated environment:

```bash
make run
```

For a physical card reader only, and simulating all other services:

```bash
make run-physical-card-reader
```

For a physical controller board only, and simulating all other services:

```bash
make run
```

For all physical device services:

```bash
make run-physical
```

Once **one** of the above commands has been run, the modified `ds-card-reader` service will automatically start up with the newly built image.
