# The Swift Virtual File System

*SVFS* is a Virtual File System for Openstack Swift built upon fuse.

### Disclaimer
This is not an official project of the Openstack community.

### Usage
`svfs -a https://auth.cloud.ovh.net/v2.0 -u user -p password -r region -t tenant /path/to/mountpoint`

### Project status
This is the start of this project, and thus it's incomplete. Check the limitations section for these specific details.

### Limitations
As the development goes, features are added one after another.
For the moment the following limitations will occur :
* SVFS is Read-Only.
* SVFS structure and node attributes are cached as you access them.
* SVFS cache is dumb and has no eviction process, thus requiring a remount to trigger a filesystem refresh.

### License
This work is under the Apache license, see the [LICENSE](LICENSE) file for details.
