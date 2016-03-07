# FAQ

### Why can't I set uid/guid and permissions ?

Openstack Swift object storage does not handle file ownership or permissions
in a way which is compatible with POSIX filesystems. Indeed, Swift supports
ACLs however it can not be converted reliably as file permissions or ownership.
A basic implementation could use Swift's metadata to make this possible but the
performance impact would be huge since a request would be necessary for every
single file within a container.

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
