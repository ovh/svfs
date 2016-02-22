# The Swift Virtual File System

*SVFS* is a Virtual File System for Openstack Swift built upon fuse.

### Disclaimer
This is not an official project of the Openstack community.

### Usage
Mount all containers for a given tenant :

```
svfs \
--os-auth-url auth_url \
--os-username user \
--os-password password \
--os-region-name region \
--os-tenant-name tenant \
/path/to/mountpoint &
```

Mount a specific container at this mountpoint rather than all containers :

```
svfs \
--os-auth-url auth_url \
--os-username user \
--os-password password \
--os-region-name region \
--os-tenant-name tenant \
--os-container-name container \
/path/to/mountpoint &
```

Use token and storage URL instead of openstack credentials (this can be useful for [hubiC](https://hubic.com)) :

```
svfs \
--os-storage-url storage_url \
--os-auth-token token \
/path/to/mountpoint &
```


### Caching

Caching can be seen as a 2-layer cache where SVFS is layer 1 and the Kernel layer 2.

You should adapt SVFS cache configuration to your network connection and system resources.
For instance a high latency system with low bandwith should maximize the cache entry size and use high ttls while a system with a reliable connection should lower these parameters or set access rates. Listing large directories or segmented objects too frequently from your storage provider may significantly slow down SVFS performance.

Available options are :
* `--cache-max-entries` : targeted cache size. Default value is -1 (it grows unlimited).
* `--cache-max-access` : targeted cache entry access count. Default value is -1 (unlimited).
* `--cache-ttl`: targeted cache entry timeout. Default value is 1 minute.

### Limitations
For the moment the following limitations will kick-in :
* SVFS container creation and removal is not supported.
* SVFS does not support opening a file in append mode.
* SVFS does not support moving directories.
* SVFS does not support moving/deleting/uploading SLO/DLO objects.

SVFS limitations and particularities of using Openstack Swift as a POSIX filesystem are discussed in the [docs](docs).

### License
This work is under the Apache license, see the [LICENSE](LICENSE) file for details.

### Author
Xavier Lucas
