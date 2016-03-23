#!/usr/bin/env ruby

pkg_path = ARGV[0]
pkg_name = ARGV[1]
pkg_main = ARGV[2]
pkg_vend = ARGV[3]
pkg_url  = ARGV[4]
pkg_info = ARGV[5]
pkg_lic  = ARGV[6]
pkg_vers = ARGV[7]

# Package dependencies
DEPENDENCIES = {
  'fuse'  => '> 2.8',
  'ruby'  => '> 1.9.1',
}

# Target platforms and architectures
TARGETS = [
  'linux'   => {
    'deb' => ['386', 'amd64', 'armhf', 'armel'],
    'rpm' => ['386', 'amd64'],
  },
]

# ARM versions mapping for go build
ARM_VERSIONS = {
  'armhf' => 6,
  'armel' => 5,
}

pkg_deps = ""

# Make release directory
unless File.directory?("#{pkg_path}")
  Dir.mkdir("#{pkg_path}")
end

# Build dependencies chain
DEPENDENCIES.each do |pkg, constraint|
  pkg_deps << "-d '#{pkg} #{constraint}' "
end

# Build go binary and package it
TARGETS.each do |target|
  target.each do |os, pkgmap|
    pkgmap.each do |pkg, archs|
      archs.each do |arch|

        go_arch = arch

        if arch.start_with?('arm')
          go_arch = 'arm'
          go_extra = "GOARM=#{ARM_VERSIONS[arch]}"
        end

        pkg_fullname = "#{pkg_path}/go-#{pkg_name}-#{os}-#{arch}"
        system("GOARCH=#{go_arch} GOOS=#{os} #{go_extra} go build -o #{pkg_fullname}")
        system("chmod 755 #{pkg_fullname}")
        system(%{fpm --force \
               -s dir \
               -t #{pkg} \
               -a #{arch} \
               -n #{pkg_name} \
               -p #{pkg_path} \
               #{pkg_deps} \
               --maintainer "#{pkg_main}" \
               --description "#{pkg_info}" \
               --license "#{pkg_lic}" \
               --url "#{pkg_url}" \
               --vendor "#{pkg_vend}" \
               --version "#{pkg_vers}" \
               #{pkg_fullname}=/usr/local/bin/#{pkg_name} \
               scripts/mount.#{pkg_name}=/sbin/mount.#{pkg_name} \
               scripts/hubic-application.rb=/usr/local/bin/hubic-application
         })

      end
    end
  end
end
