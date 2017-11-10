# The Swift Virtual File System

[![Release](https://badge.fury.io/gh/ovh%2Fsvfs.svg)](https://github.com/vpalmisano/svfs/releases)
[![Github All Releases](https://img.shields.io/github/downloads/ovh/svfs/total.svg)](https://github.com/vpalmisano/svfs/releases)
[![Build Status](https://travis-ci.org/ovh/svfs.svg?branch=master)](https://travis-ci.org/ovh/svfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/vpalmisano/svfs)](https://goreportcard.com/report/github.com/vpalmisano/svfs)
[![Coverage Status](https://coveralls.io/repos/github/ovh/svfs/badge.svg?branch=master)](https://coveralls.io/github/ovh/svfs?branch=master)
[![GoDoc](https://godoc.org/github.com/vpalmisano/svfs/svfs?status.svg)](https://godoc.org/github.com/vpalmisano/svfs/svfs)

**SVFS** is a Virtual File System over Openstack Swift built upon fuse. It is compatible with [hubiC](https://hubic.com),
[OVH Public Cloud Storage](https://www.ovh.com/fr/cloud/storage/object-storage) and basically every endpoint using a standard
Openstack Swift setup.  It brings a layer of abstraction over object storage, making it as accessible and convenient as a
filesystem, without being intrusive on the way your data is stored.

## Disclaimer
This is not an official project of the Openstack community.

## Installation

Download and install the latest [release](https://github.com/vpalmisano/svfs/releases) packaged for your distribution.

## Usage

#### Mount command

On Linux (requires fuse and ruby) :

```
mount -t svfs -o <options> <device> /mountpoint
```

On OSX (requires osxfuse and ruby) :

```
mount_svfs <device> /mountpoint -o <options>
```

Notes :
- You can pick any name you want for the `device` parameter.
- All available mount options are described later in this document.

Credentials can be specified in mount options, however this may be desirable to read them from an external source. The following sections desribe alternative approaches.

#### Reading credentials from the environment

SVFS supports reading the following set of environment variables :

* If you are using HubiC :
```
 HUBIC_AUTH
 HUBIC_TOKEN
```
* If you are using a vanilla Swift endpoint (like OVH PCS), after sourcing your [OpenRC](http://docs.openstack.org/user-guide/common/cli-set-environment-variables-using-openstack-rc.html) file :
```
 OS_AUTH_URL
 OS_USERNAME
 OS_PASSWORD
 OS_REGION_NAME
 OS_TENANT_NAME
```
* If you already authenticated to an identity endpoint :
```
 OS_AUTH_TOKEN
 OS_STORAGE_URL
```

#### Reading credentials from a configuration file

All environment variables can also be set in a YAML configuration file placed at `/etc/svfs.yaml`.

For instance :
```yaml
hubic_auth: XXXXXXXXXX..
hubic_token: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX...
```

## Usage with OVH products

- Usage with OVH Public Cloud Storage is explained [here](docs/PCS.md).
- Usage with hubiC is explained [here](docs/HubiC.md).

## Mount options

#### Keystone options

* `auth_url`: keystone URL (default is https://auth.cloud.ovh.net/v2.0).
* `username`: your keystone user name.
* `password`: your keystone password.
* `tenant`: your project name.
* `region`: the region where your tenant is.
* `version`: authentication version (`0` means auto-discovery which is the default).
* `storage_url`: the storage endpoint holding your data.
* `internal_endpoint`: the storage endpoint type (default is `false`).
* `token`: a valid token.

Options `region`, `version`, `storage_url` and `token` are guessed during authentication if
not provided.

#### Hubic options

* `hubic_auth`: hubic authorization token as returned by the `hubic-application` command.
* `hubic_times` : use file times set by hubic synchronization clients. Option `attr`
should also be set for this to work.
* `hubic_token` : hubic refresh token as returned by the `hubic-application` command.

#### Swift options

* `container`: which container should be selected while mounting the filesystem. If not set,
all containers within the tenant will be available under the chosen mountpoint.
* `storage_policy`: expected containers storage policy. This is used to ignore containers
not matching a particular storage policy name. If empty, this setting is ignored (default).
* `segment_size`: large object segments size in MB. When an object has a content larger than
this setting, it will be uploaded in multiple parts of the specified size. Default is 256 MB.
Segment size should not exceed 5 GB.
* `connect_timeout`: connection timeout to the swift storage endpoint. Default is 15 seconds.
* `request_timeout`: timeout of requests sent to the swift storage endpoint. Default is 5 minutes.

#### Prefetch options

* `block_size`: Filesystem block size in bytes. This is only used to report correct `stat()` results.
* `readahead_size`: Readahead size in KB. Default is 128 KB.
* `readdir`: Overall concurrency factor when listing segmented objects in directories (default is 20).
* `attr`: Handle base attributes.
* `xattr`: Handle extended attributes.
* `transfer_mode`: Enforce network transfer optimizations. The following flags / features can be combined :
 - `1` : disable explicit empty file creation.
 - `2` : disable explicit directory creation.
 - `4` : disable directory content check on removal.
 - `8` : disable file check in read only opening.

#### Cache options

* `cache_access`: cache entry access count before refresh. Default is -1 (unlimited access).
* `cache_entries`: maximum entry count in cache. Default is -1 (unlimited).
* `cache_ttl`: cache entry timeout before refresh. Default is 1 minute.

#### Access restriction options

* `allow_other`: Bypass `allow_root`.
* `allow_root`: Restrict access to root and the user mounting the filesystem.
* `default_perm`: Restrict access based on file mode (useful with `allow_other`).
* `uid`: default files uid (defaults to current user uid).
* `gid`: default files gid (defaults to current user gid).
* `mode`: default files permissions (default is 0700).
* `ro`: enable read-only access.

#### Debug options

* `debug`: enable debug log.
* `stdout` : stdout redirection expression (e.g. `>/dev/null`).
* `stderr` : stderr redirection expression (e.g. `>>/var/log/svfs.log`).
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

* Opening files in other modes than `O_CREAT`, `O_RDONLY` and `O_WRONLY`.
* Moving directories.
* Renaming containers.
* SLO (but supports DLO).
* Per-file uid/gid/permissions (but per-mountpoint).
* Symlink targets across containers (but within the same container).

Take a look at the [docs](docs) for further discussions about SVFS approach.

## FAQ

Got errors using `rsync` with svfs ? Can't change creation time ? Why svfs after all ?

Take a look at the [FAQ](docs/FAQ.md).

## Hacking

Make sure to use the latest version of go and follow [contribution guidelines](CONTRIBUTING.md) of SVFS.

## License
This work is under the BSD license, see the [LICENSE](LICENSE) file for details.
