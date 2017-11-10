#!/usr/bin/env ruby

require 'git'
require 'logger'
require 'rake/testtask'
require 'rubygems'
require_relative 'scripts/package'

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
  system("gem install bundler")
  system("bundle install")
  system("go get github.com/aktau/github-release")
end

# -----------------------------------------
#  RELEASE
# -----------------------------------------
desc 'Release a new version'
task :release, [:version] => [:prepare_release] do |t, args|
  path = package[:path]
  tag = "v#{args.version}"

  # Update source
  source = File.read("svfs/version.go")
  new_source = source.gsub(/Version = "[^\"]+"/,"Version = \"#{args.version}\"")
  File.open("svfs/version.go", "w") { |file| file << new_source }

  # Push on master and wait for build to complete
  g = Git.open("#{ENV['GOPATH']}/src/github.com/vpalmisano/svfs", :log => Logger.new(STDOUT))
  g.checkout(:master)
  g.add(['docs/RELEASE.md', 'svfs/version.go'])
  g.commit(["Release #{args.version}",
    "",
    "Signed-off-by: #{g.config('user.name')} <#{g.config('user.email')}>",
    "",
  ].join("\n"))
  g.push(:origin, :master)

  # Should use the travis lib once access token issue is fixed
  # See https://github.com/travis-ci/travis.rb/issues/315
  print "Has travis build passed ? (y/N) "
  unless STDIN.gets.chomp == 'y'
    abort("User asked to abort releasing process")
  end

  # Merge to release branch and push
  g.checkout(:release)
  g.merge(:master)
  g.add_tag(tag)
  g.push(:origin, :release)
  g.push(:origin, :release, {:tags => true})

  # Release
  mkdir_p path
  release(package, args.version)

  description = File.read("docs/RELEASE.md")
  system("github-release release --user ovh --repo svfs --tag #{tag} --name \"Version #{args.version}\" --description \"#{description}\" --pre-release")

  Dir["#{path}/*"].each do |file|
    system("github-release upload --user ovh --repo svfs --tag #{tag} --name #{File.basename(file)} --file #{file}")
    File.delete(file)
  end

  rm_rf path
  g.checkout(:master)
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
