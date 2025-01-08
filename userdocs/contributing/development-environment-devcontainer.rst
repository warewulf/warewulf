=============================
Development Environment (Dev Container/VSC)
=============================

Using a Dev Container for development
=====================================================

Visual Studio Code (VSC) can utilize a Dev Container for a self-contained environment that has all the necessary tools and dependencies to build and test Warewulf. The Dev Container is based on the Rocky 9 image and is built using the `devcontainer.json` file in the `.devcontainer` directory of the Warewulf repository.  To use this working Docker/Podman and VSC installations are required.  To use the Dev Container, click the "Open a Remote Window" button on the bottom left of the editor (`><` icon) and select "Reopen in Container".  This will build the container and open a new VSC window with the container as the development environment. 
