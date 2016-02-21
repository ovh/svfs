# The Swift Virtual File System

*SVFS* is a Virtual File System for Openstack Swift built upon fuse.

### Disclaimer
This is not an official project of the Openstack community.

### Usage
Mount all containers for a given tenant :

`svfs -a auth_url -u user -p password -r region -t tenant /path/to/mountpoint &`

Mount a specific container at this mountpoint rather than all containers :

`svfs -a auth_url -u user -p password -r region -t tenant -c container /path/to/mountpoint &`

Use token and storage URL instead of openstack credentials (this can be useful for [hubiC](https://hubic.com)) :

`svfs -s storage_url -k token /path/to/mountpoint &`


### Caching

Caching can be seens as a 2-layer cache where SVFS is layer 1 and the Kernel layer 2.

SVFS cache configuration is described below.

* Targeted cache size : `--cache-max-entries`. Default is -1 (it grows unlimited).
* Targeted cache entry timeout : `--cache-ttl`. Default is 1 minute.
* Targeted cache entry access count : `--cache-max-access` Default is -1 (unlimited).

### Limitations
For the moment the following limitations will kick-in :
* SVFS container creation and removal is not supported.
* SVFS does not support opening a file in append mode.
* SVFS does not support moving directories.
* SVFS does not support moving/deleting/uploading SLO/DLO objects and may show erroneous file size of 0 (manifest file).

SVFS limitations and particularities of using Openstack Swift as a POSIX filesystem are discussed in the [docs](docs).

### License
This work is under the Apache license, see the [LICENSE](LICENSE) file for details.

### Author
Xavier Lucas
