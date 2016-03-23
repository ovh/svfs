# FAQ

### How does it compare to hubicfuse, cloudfuse, swiftFS ... ?

SVFS supports authentication versions 1/2/3, is stable and fully connected.
This means svfs won't play with unsafe pointers all the time,
avoiding segfaults. Also, svfs will make read/write operations atomic :
it won't return to the caller until data has been read/written from/to the
network. In other words no local buffering is taking place so i.e. if you
want to rsync to/from an svfs mountpoint you will get the real progress
of this operation and not the progress of local buffering. An other consequence
is that you'll never require space on your local disk to read/write data from
a file. Standard buffering is done in memory and your data is always a stream
on the network. This is not the role of a network filesystem to chose how your
data should be accessed : it should be consistent across operations. If you are
looking for local speed rates, then this means you are looking for a local
filesystem and svfs is no more than a easier way to achieve synchronization
between both since it brings support for usual tools when used with appropriate options
(for instance `rsync -av -W --inplace --update --progress`). In this case you should
rely on an appropriate, journalized, battle-hardened local filesystem. This is also
where you should manage ownership, permissions and other ACL/extended attributes
information, relatively to your local users and groups.

### Why can't I set uid/guid and permissions ?

Openstack Swift object storage does not handle file ownership or permissions
in a way which is compatible with POSIX filesystems. Indeed, Swift supports
ACLs however it can not be converted reliably as file permissions or ownership.
A basic implementation could use Swift's metadata to make this possible but the
performance impact would be huge since a request would be necessary for every
single file within a container and it makes little sense as uid/gid mapping
can differ between two mountpoints.

### Why are creation/access times erroneous ?

Openstack Swift only stores the modification time of an object so this
information won't be available when used as a POSIX filesystem. Again, using
metadata could solve this issue but the performance impact wouldn't make this
a worthy tradeoff.

### Why does an entire tree disappears when I remove the sole object in it ?

Openstack Swift can support directories as standard objects when they are
uploaded without content. However, most of the time swift clients will not
proceed this way. In this case, deleting an object will mean deleting all
empty intermediate directories within the object path as well.

### Does it run on Mac OS X ?

SVFS is tested on Linux, however this should run out of the box under Mac OS X.
Feedback are welcome on this particular aspect.
