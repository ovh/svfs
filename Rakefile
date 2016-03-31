#!/usr/bin/env ruby

require 'rake/testtask'

package = {
  :name       => 'svfs',
  :vendor     => 'OVH SAS',
  :maintainer => 'xavier.lucas@corp.ovh.com',
  :info       => 'The Swift Virtual Filesystem',
  :url        => 'https://www.ovh.com',
  :licence    => 'BSD',
  :path       => 'releases',
}

# -----------------------------------------
#  PREPARE-RELEASE
# -----------------------------------------
desc 'Prepare the releasing processs'
task :prepare_release do |t, args|
  system("gem install fpm")
end

# -----------------------------------------
#  RELEASE
# -----------------------------------------
desc 'Release a new version'
task :release, [:version] => [:prepare_release] do |t, args|
  system(%Q(scripts/package.rb \
         "#{package[:path]}" \
         "#{package[:name]}" \
         "#{package[:maintainer]}" \
         "#{package[:vendor]}" \
         "#{package[:url]}" \
         "#{package[:info]}" \
         "#{package[:licence]}" \
         "#{args.version}"
  ))
end

# -----------------------------------------
#  TESTS
# -----------------------------------------
# In order to run, the following env vars
# must be set.
#
# TEST_MOUNTPOINT : an svfs mountpoint
# TEST_SEG_SIZE   : segmented file size
# TEST_NSEG_SIZE  : standard file size
#
desc 'Run tests'
Rake::TestTask.new do |t|
  t.libs << "test"
  t.test_files = FileList['test/*.rb']
  t.verbose = true
end
