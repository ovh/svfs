# The Swift Virtual File System

[![Build Status](https://travis-ci.org/ovh/svfs.svg?branch=master)](https://travis-ci.org/ovh/svfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/ovh/svfs)](https://goreportcard.com/report/github.com/ovh/svfs)
[![GoDoc](https://godoc.org/github.com/ovh/svfs/svfs?status.svg)](https://godoc.org/github.com/ovh/svfs/svfs)

**SVFS** is a Virtual File System over Openstack Swift built upon fuse. It is compatible with [hubiC](https://hubic.com),
[OVH Public Cloud Storage](https://www.ovh.com/fr/cloud/storage/object-storage) and basically every endpoint using a standard
Openstack Swift setup.  It brings a layer of abstraction over object storage, making it as accessible and convenient as a
filesystem, without being intrusive on the way your data is stored.

## Disclaimer
This is not an official project of the Openstack community.

## Installation

Download and install the latest [release](https://github.com/ovh/svfs/releases) packaged for your distribution.

## Usage

You can either use standard mount conventions or use the svfs binary directly.

Using the mount command :
```
mount -t svfs -o username=..,password=..,tenant=..,region=..,container=.. myName /mountpoint
```

Using `/etc/fstab` :
```
myName   /mountpoint   svfs   username=..,password=..,tenant=..,region=..,container=..  0 0
```

Using svfs directly :

```
svfs --os-username=.. --os-password=.. ... myName /mountpoint &
```

With OSX after [osxfuse](https://github.com/osxfuse/osxfuse), ruby and last pkg installation :

```
mount_svfs myName /mountpoint -o username=..,password=..,tenant=..,region=..,container=..
```


## Usage with OVH products

- Usage with OVH Public Cloud Storage is explained [here](docs/PCS.md).
- Usage with hubiC is explained [here](docs/HubiC.md).

## FAQ

Got errors using `rsync` with svfs ? Can't change creation time ? Why svfs after all ?

Take a look at the [FAQ](docs/FAQ.md).


## Options

#### Keystone options

* `identity_url`: keystone URL (default is https://auth.cloud.ovh.net/v2.0).
* `username`: your keystone user name.
* `password`: your keystone password.
* `tenant`: your project name.
* `region`: the region where your tenant is.
* `version`: authentication version (`0` means auto-discovery which is the default).
* `storage_url`: the storage endpoint holding your data.
* `token`: a valid token.

Options `region`, `version`, `storage_url` and `token` are guessed during authentication if
not provided.

#### Hubic options

* `hubic_auth`: hubic authorization token as returned by the `hubic-application` command.
* `hubic_times` : use file times set by hubic synchronization clients. Option `extra_attr`
should also be set for this to work.
* `hubic_token` : hubic refresh token as returned by the `hubic-application` command.

#### Swift options

* `container`: which container should be selected while mounting the filesystem. If not set,
all containers within the tenant will be available under the chosen mountpoint.
* `segment_size`: large object segments size in MB. When an object has a content larger than
this setting, it will be uploaded in multiple parts of the specified size. Default is 256 MB.
Segment size should not exceed 5 GB.
* `timeout`: connection timeout to the swift storage endpoint. If an operation takes longer
than this timeout and no data has been seen on open sockets, an error is returned. This can
happen when copying non-segmented large files server-side. Default is 5 minutes.

#### Prefetch options

* `block_size`: Filesystem block size in bytes. This is only used to report correct `stat()` results.
* `readahead_size`: Readahead size in KB. Default is 128 KB.
* `readdir`: Overall concurrency factor when listing segmented objects in directories (default is 20).
* `extra_attr`: Fetch extended attributes.

#### Cache options

* `cache_access`: cache entry access count before refresh. Default is -1 (unlimited access).
* `cache_entries`: maximum entry count in cache. Default is -1 (unlimited).
* `cache_ttl`: cache entry timeout before refresh. Default is 1 minute.

#### Access restriction options

* `allow_other`: Bypass `allow_root`.
* `allow_root`: Restrict access to root and the user mounting the filesystem.
* `default_perm`: Restrict access based on file mode (useful with `allow_other`).
* `uid`: default files uid (default is 0 i.e. root).
* `gid`: default files gid (default is 0 i.e. root).
* `mode`: default files permissions (default is 0700).
* `ro`: enable read-only access.

#### Debug options

* `debug`: enable debug log.
* `profile_addr`: Golang profiling information will be served at this address (`ip:port`) if set.
* `profile_cpu`: Golang CPU profiling information will be stored to this file if set.
* `profile_ram`: Golang RAM profiling information will be stored to this file if set.

#### Performance options
* `go_gc`: set garbage collection target percentage. A garbage collection is triggered when the
heap size exceeds, by this rate, the remaining heap size after the previous collection. A lower
value triggers frequent GC, which means memory usage will be lower at the cost of higher CPU
usage. Setting a higher value will let the heap size grow by this percent without collection,
reducing GC frequency. A Garbage collection is forced if none happened for 2 minutes. Note that
unused heap memory is not reclaimed after collection, it is returned to the operating system
only if it appears unused for 5 minutes.


## Limitations

**Be aware that SVFS doesn't transform object storage to block storage.**

SVFS doesn't support :

* Opening files in append mode.
* Moving directories.
* Renaming containers.
* SLO (but supports DLO).
* Per-file uid/gid/permissions (but per-mountpoint).
* Symlink targets across containers (but within the same container).

Take a look at the [docs](docs) for further discussions about SVFS approach.

## Hacking

Make sure to use the latest version of go and follow [contribution guidelines](CONTRIBUTING.md) of SVFS.

## License
This work is under the BSD license, see the [LICENSE](LICENSE) file for details.
