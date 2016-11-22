# Changelog

### IMPROVEMENTS :

- [#102] Directory times are now set to the filesystem mount time.
- [#101] Application panic events are pushed to syslog.
- Option extra_attr renamed to attr.

### FEATURE :

- [#103] New option : xattr can now be used to handle extended attributes on files.

### BUGFIX :

- [#105] fsync(2) calls are now handled as a no-op.
