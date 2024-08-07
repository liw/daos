# MPI-IO Support

The Message Passing Interface (MPI) Standard,
maintained by the [MPI Forum](https://www.mpi-forum.org/docs/),
includes a chapter on MPI-IO.

[ROMIO](https://www.mcs.anl.gov/projects/romio/) is a well-known
implementation of MPI-IO and is included in many MPI implementations.
DAOS provides its own MPI-IO ROMIO ADIO driver.
This driver has been merged in the upstream MPICH repository, see the
[adio/ad\_daos](https://github.com/pmodels/mpich/tree/main/src/mpi/romio/adio/ad_daos)
section in the MPICH git repository for details.


## Supported MPI Version

### MPICH

The DAOS ROMIO ADIO driver has been accepted into [MPICH](https://www.mpich.org/).
It is included in
mpich-3.4.1 (released Jan 2021),
mpich-3.4.2 (released May 2021),
[mpich-3.4.3 (released Dec 2021)](https://www.mpich.org/downloads/)
as well as the MPICH 4.x releases.

!!! note
    Support for container labels has been added with MPICH 3.4.3.
    Older MPICH releases only support container UUIDs.

!!! note
    Starting with DAOS 1.2, the `--svc` parameter (number of service replicas)
    is no longer needed, and the DAOS API has been changed accordingly.
    Patches have been contributed to MPICH that detect the DAOS API version
    to gracefully handle this change. MPICH 3.4.2 and later includes those changes.

To build MPICH, including ROMIO with the DAOS ADIO driver:

```bash
export MPI_LIB=""

# to clone the latest development snapshot:
git clone https://github.com/pmodels/mpich
cd mpich

# to clone a specific tagged version:
git clone -b v3.4.3 https://github.com/pmodels/mpich mpich-3.4.3
cd mpich-3.4.3

git submodule update --init

./autogen.sh

PREFIX=$HOME/software/mpich
# or PREFIX=$HOME/software/mpich-3.4.3 for a specific tagged version
mkdir -p $PREFIX

./configure --prefix=$PREFIX --enable-fortran=all --enable-romio \
 --enable-cxx --enable-g=all --enable-debuginfo --with-device=ch3:nemesis \
 --with-file-system=ufs+daos --with-daos=/usr

# compiling 3.4.3 may need FFLAGS=-fallow-argument-mismatch on the configure command line

make -j8; make install
```

This assumes that DAOS is installed into the `/usr` tree, which is the case for
the DAOS RPM installation. Other configure options can be added, modified, or
removed as needed, like the network communicatio device, fortran support,
etc. For those, please consule the mpich user guide.

Set the `PATH` and `LD_LIBRARY_PATH` and `INCLUDE` to the path of the installed MPICH
before building your applications that use MPI:

```bash
export PATH="$PREFIX/bin:$PATH"
export LD_LIBRARY_PATH="$PREFIX/lib:$LD_LIBRARY_PATH"
export INCLUDE="$PREFIX/include:$INCLUDE"
```

To use MPI-IO in MPICH with DAOS, the following
environment variable can be used to instruct the runtime
to interpret all file paths in MPI-IO calls as DAOS paths:

```bash
export ROMIO_FSTYPE_FORCE="daos:"
```

Alternatively, the individual file paths can be prefixed with `daos:`.

Using either of these two methods will enable the MPI-IO DFS backend.
More information can be found in the
[ANL ROMIO pages](https://wordpress.cels.anl.gov/romio/2019/02/20/useful-environment-variables/).

### Intel MPI

The [Intel MPI Library](https://software.intel.com/content/www/us/en/develop/tools/mpi-library.html)
includes DAOS support since the
[2019.8 release](https://software.intel.com/content/www/us/en/develop/articles/intel-mpi-library-release-notes-linux.html).

Note that Intel MPI uses `libfabric` and includes it as part of the Intel MPI installation:
* 2019.8 and 2019.9 includes `libfabric-1.10.1-impi`
* 2021.1, 2021.2 and 2021.3 includes `libfabric-1.12.1-impi`

Care must be taken to ensure that the version of libfabric that is used
is at a level that includes the patches that are critical for DAOS.
DAOS 1.0.1 includes `libfabric-1.9.0`, DAOS 1.2 includes `libfabric-1.12`,
DAOS 2.0 includes `libfabric-1.14`, and DAOS 2.2 includes `libfabric-1.15`.

To use DAOS with Intel MPI, the `libfabric` that is supplied by DAOS
(and that is installed into `/usr/lib64` by default) must be used.
Intel MPI provides a mechanism to indicate that the Intel MPI version of
`libfabric` should **not** be used, by setting this variable **before**
loading the Intel MPI environment:

```bash
export I_MPI_OFI_LIBRARY_INTERNAL=0
```

This is normally sufficient to ensure that the `libfabric` provided by DAOS is used.
Depending on how the environment is set up, it may be necessary to add the
system library search path back as the first path in the library search path:

```bash
export LD_LIBRARY_PATH="/usr/lib64/:$LD_LIBRARY_PATH"
```

There are other environment variables that need to be set on the client side to
ensure proper functionality with the DAOS MPIIO driver, including:

```bash
export FI_UNIVERSE_SIZE=16383
export FI_OFI_RXM_USE_SRX=1
```

To use MPI-IO in Intel MPI with DAOS, the following two
environment variables can be used to instruct the runtime
to interpret all file paths in MPI-IO calls as DAOS paths:

```bash
export I_MPI_FILESYSTEM=on
export I_MPI_FILESYSTEM_FORCE=daos
```
(Note that in older versions of Intel MPI, the environment variables
`I_MPI_EXTRA_FILESYSTEM=on` and `I_MPI_EXTRA_FILESYSTEM_FORCE=daos`
were used.)

Alternatively, the individual file paths can be prefixed with `daos:`.

Using either of these two methods will enable the MPI-IO DFS backend.
More information on the `I_MPI` variables can be found in the
[Intel MPI documentation](https://www.intel.com/content/www/us/en/docs/mpi-library/developer-reference-linux/2021-13/other-environment-variables.html).

### Open MPI

[Open MPI](https://www.open-mpi.org/) 4.0.5 does not yet provide DAOS support.
Since one of its MPI-IO implementations is based on ROMIO,
it will likely pick up DAOS support in an upcoming release.

### MVAPICH2

[MVAPICH2](https://mvapich.cse.ohio-state.edu/) 2.3.4 does not yet provide DAOS support.
Since its MPI-IO implementation is based on ROMIO,
it will likely pick up DAOS support in an upcoming release.


## Testing MPI-IO with DAOS

Build any client (HDF5, ior, mpi test suites) normally with the mpicc command
and mpich library installed above.

### Using the UNS

DAOS UNS allows encoding pool and container information into a path on the filesystem, so one can
easily access that container using that path instead of using explicit addressing using the pool and
container uuids/labels.

Create a container with a path on dfuse or lustre, or any file system that supports extended
attributes:
```bash
daos cont create mypool mycont --path=/mnt/dfuse/ --type POSIX
```

Then using that path, one can start creating files using the DAOS MPIIO driver by just appending
`daos:` to the filename/path. For example:
`daos:/mnt/dfuse/file`
`daos:/mnt/dfuse/dir1/file`

### Using a Prefix Environment Variable

Another way to use the DAOS MPIIO driver is using an environment variable to set the prefix itself
for the file:
```bash
export DAOS_UNS_PREFIX="path"
```
That prefix path can be:

1. The UNS prefix if that exists (similar to the UNS mode above): /mnt/dfuse

2. A direct path using the pool and container label (or uuid): daos://pool/container/

Then one can specify the path to the file relative to the root of the container being set in the
prefix. So in the example above, if the file to be accessed is under /dir1 in the container, one
would pass `daos:/dir1/file' to MPI_File_open().

### Using Pool and Container Environment Variables

This mode is meant just for quick testing to use the MPIIO DAOS driver bypassing the UNS and setting
direct access with pool and container environment variables. At the client side, the following
environment variables need to be set:
`export DAOS_POOL={uuid/label}; export DAOS_CONT={uuid/label}; export DAOS_BYPASS_DUNS=1`.
The user still need to append the `daos:` prefix to the file being passed to MPI_File_open().

## Known limitations

Limitations of the current implementation include:

-   No support for MPI file atomicity, preallocate, or shared file pointers.
