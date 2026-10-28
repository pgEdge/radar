# Installing the pgEdge Package

pgEdge publishes radar as the `pgedge-radar` package in the pgEdge Enterprise
Postgres repository for Enterprise Linux, Debian, and Ubuntu. The repository
configuration enables signature checks with the pgEdge signing key. The package
installs the `radar` binary in `/usr/bin`.

Use the following steps to install the package:

1. Configure the pgEdge Enterprise Postgres repository by following the
   instructions for
   [Enterprise Linux](https://docs.pgedge.com/enterprise/el/configure-repo/) or
   for
   [Debian and Ubuntu](https://docs.pgedge.com/enterprise/debian/configure-repo/).

2. Install the package; on Enterprise Linux, run the following command:

    ```bash
    sudo dnf install -y pgedge-radar
    ```

    On Debian and Ubuntu, run the following commands:

    ```bash
    sudo apt-get update
    sudo apt-get install -y pgedge-radar
    ```

3. Confirm the installation by printing the radar version:

    ```bash
    radar --version
    ```
