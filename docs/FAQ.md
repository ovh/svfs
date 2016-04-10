# FAQ

### How does it compare to hubicfuse, cloudfuse, swiftFS ... ?

SVFS supports authentication versions 1/2/3, is stable and fully connected.
This means svfs won't play with unsafe pointers all the time,
avoiding segfaults. Also, svfs will make read/write operations atomic :
it won't return to the caller until data has been read/written from/to the
network. In other words no local buffering is taking place so i.e. if you
want to rsync to/from a svfs mountpoint you will get the real progress
of this operation and not the progress of local buffering. Another consequence
is that you'll never require space on your local disk to read/write data from
a file. Standard buffering is done in memory and your data is always a stream
on the network. This is not the role of a network filesystem to chose how your
data should be accessed : it should be consistent across operations. If you are
looking for local speed rates, then this means you are looking for a local
filesystem and svfs is no more than an easier way to achieve synchronization
between both since it brings support for usual tools when used with appropriate options
(for instance `rsync -avW --inplace --no-g --no-o --progress`). In this case you should
rely on an appropriate, journalized, battle-hardened local filesystem. This is also
where you should manage ownership, permissions and other ACL/extended attributes
information, relatively to your local users and groups.

### Why can't I set uid/guid and permissions ?

Openstack Swift does not handle file ownership or permissions in a way which is
familiar to filesystems. It is actually relying on some form of ACLs which can't
be a good match for filesystem permissions. Also, it has little sense to set
file ownership or permissions over object storage : there's no such thing as
uid/gid when you store an object : it's *your* object. These informations come
from your local filesystem while you are storing data on a remote location.
Given that, svfs doesn't support setting this information per file but provides
per mountpoint options.

### Why are access/creation/modification times erroneous ?

Openstack Swift generates and stores modification time so that users can't change
it. In svfs we use metadata to store this information if you supply a specific
mount option (`extended_attr`). This has a performance impact since fetching
metadata is only possible by requesting extra details on each node.
So if you want the best performance, you shouldn't use it. Note that mtime
can't be set on a directory/container/mountpoint because every change occuring
within one of this node would trigger too many requests. Usually that's not an
issue for backup tools as they don't rely on directory metas.

### Why does an entire tree disappear when I remove the sole object in it ?

Openstack Swift can support directories as standard objects when they are
uploaded without content. However, most of the time swift clients will not
proceed this way. In this case, deleting an object will mean deleting all
empty intermediate directories within the object path as well.

### Does it run on Mac OS X ?

SVFS is tested on Linux, however this should run out of the box under Mac OS X.
Feedback are welcome on this particular aspect.
