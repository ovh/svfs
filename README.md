# The Swift Virtual File System

*SVFS* is a Virtual File System for Openstack Swift built upon fuse.

### Disclaimer
This is not an official project of the Openstack community.


### Requirements

You will need :

- fuse
- ruby

### Installation

Download the latest [release](https://github.com/xlucas/svfs/releases) and unzip it.

Then :
```
   $ mv svfs /usr/local/bin/svfs
   $ chmod +x !$

   $ mv mount.svfs /sbin/mount.svfs
   $ chmod +x !$
```

### Usage

You can either use standard mount conventions :

```
mount -t svfs -o username=..,password=..,tenant=..,region=..,container=.. myName /mountpoint
```

Change your `/etc/ftab` :
```
myName /path/to/mountpoint svfs user=..,password=..,tenant=..,region=..,container=.. 0 0
```

Or use the svfs command directly :

```
svfs --os-username=.. --os-password=.. ... myName /mountpoint &
```

### Options

#### Keystone options

* `identity_url`: keystone endpoint URL (default is https://auth.cloud.ovh.net/v2.0).
* `username`: your keystone user name.
* `password`: your keystone password.
* `tenant`: your project name.
* `region`: the region where your tenant is.
* `version`: authentication version (0 means auto-discovery which is the default).

In case you already have a token and storage URL (for instance with [hubiC](https://hubic.com)) :
* `storage_url`: the URL to your data
* `token`: your token

#### Swift options

* `container`: which container should be selected while mounting the filesystem. If not set,
all containers within the tenant will be available under the chosen mountpoint.
* `segment_size`: large object segments size in MB. When an object has a content larger than
this setting, it will be uploaded in multiple parts, each of this size. Default is 256 MB.
* `timeout`: connection timeout to the swift storage endpoint. If an operation takes longer
than this timeout and no data has been seen on open sockets, it will stop and return as an
error. This can happen when copying very large files server-side. Default is 5 minutes.

#### Prefetch options

* `readahead_size`: Readahead size in bytes. Default is 128 KB.
* `readdir`: Overall concurrency factor when listing segmented objects in directories (default is 20).

#### Cache options

* `cache_access`: targeted cache entry access count. Default is -1 (unlimited).
* `cache_entries`: targeted cache size. Default is -1 (it grows unlimited).
* `cache_ttl`: targeted cache entry timeout. Default is 1 minute.

#### Debug options

* `debug`: set to true to enable debug log.
* `profile_cpu`: path where golang CPU profiling will be stored.
* `profile_ram`: path where golang RAM profiling will be stored.

### Limitations
* SVFS does not support creating, moving or deleting containers.
* SVFS does not support opening a file in append mode.
* SVFS does not support moving directories.
* SVFS does not support SLO (but supports DLO).
* SVFS does not support uid/gid/permissions.

SVFS limitations and particularities of using Openstack Swift as a POSIX filesystem are discussed in the [docs](docs).

### License
This work is under the Apache license, see the [LICENSE](LICENSE) file for details.

### Author
Xavier Lucas
