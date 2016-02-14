# The Swift Virtual File System

*SVFS* is a Virtual File System for Openstack Swift built upon fuse.

### Disclaimer
This is not an official project of the Openstack community.

### Usage
`svfs -a https://auth.cloud.ovh.net/v2.0 -u user -p password -r region -t tenant /path/to/mountpoint`

### Project status
This is the start of this project, and thus it's obviously missing pieces. Take a look at the limitations section for details.

### Limitations
As the development goes, features are added one after another. For the moment the following limitations will occur :
* SVFS is using a dumb cache management thus distributed access is not supported.
* SVFS structure and node attributes are cached as you access them and eviction only occurs on write or remove operations.
* SVFS node size is not refreshed after a write operation.
* SVFS container creation and removal is not supported.
* SVFS does not support move/rename/mkdir operations for now (a dirty trick will be required due to the way swift works).

### License
This work is under the Apache license, see the [LICENSE](LICENSE) file for details.

### Author
Xavier Lucas
