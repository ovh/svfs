# Unit tests guide


## Prerequisites

You must have [Rake](http://rake.rubyforge.org/) installed on your testing environment before running unit tests, you can install it using `gem install rake`.

You can see all possible commands using `Rake -T`.

Before running unit tests, you have to set 3 env vars :

* `TEST_MOUNTPOINT`: the svfs mountpoint.
* `TEST_SEG_SIZE`: segmented file size value (in megabytes).
* `TEST_NSEG_SIZE`: standard file size value (in megabytes).

Now the test container will be created automatically.


## Launch unit tests

In order to run unit tests, you just have to use `rake test` command once you have followed previous instructions.


## Write unit tests

At first, you must follow [contribution guidelines](CONTRIBUTING.md) of SVFS and use the latest version of go.

All the tests are written in the [test directory](test), if you want to add some tests you need consider the [Rakefile](Rakefile) and comment your tests like previously written tests.

