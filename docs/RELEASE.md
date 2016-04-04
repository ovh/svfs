# Changelog

### IMPROVEMENT :
- [#33] When uploading objects, auto-detect `content-type` header from file suffix.
- [#35] Support for opening a file with an offset (make svfs usable with file browsers,
allow media streaming, etc).

### BUGFIXES :
- [#37] Segments not removed in certain cases due to bulk-delete.
- [#38] Don't allow segment size greater than Swift's maximum object size.
