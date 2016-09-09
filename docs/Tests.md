# Running svfs tests

When contributing to svfs, you'll have to make sure you contribution doesn't break current features. Always make sure unit and integration tests pass.


## Unit tests

#### Authentication

In order to test features against a remote Swift storage, tests use your environment for authentication :

- Set `SVFS_TEST_AUTH` to either `HUBIC` or `OPENRC`.
- Set relevant variables as described below.

#### HubiC

```
SVFS_TEST_HUBIC_AUTH
SVFS_TEST_HUBIC_TOKEN
```

#### OpenRC

```
SVFS_TEST_AUTH_URL
SVFS_TEST_USERNAME
SVFS_TEST_PASSWORD
SVFS_TEST_TENANT_NAME
SVFS_TEST_REGION_NAME
```

#### Execution

Run `go test -v github.com/ovh/svfs/svfs`.


## Integration tests

#### Prerequisites

You must have [Rake](http://rake.rubyforge.org/) installed before running tests, you can install it using `gem install rake`.

Before running integration tests, you have to set 3 environment variables :

* `TEST_MOUNTPOINT` : the svfs mountpoint.
* `TEST_SEG_SIZE` : segmented file size value (in megabytes).
* `TEST_NSEG_SIZE` : standard file size value (in megabytes).

A test container will be created automatically.


#### Execution

Use `rake test` command once you have followed previous instructions.

#### Writing more tests

At first, you must follow [contribution guidelines](CONTRIBUTING.md) of SVFS and use the latest version of go.
All integration tests are located in the [test directory](test). If you want to add some, you need consider the [Rakefile](Rakefile) and comment your tests like existing ones.
