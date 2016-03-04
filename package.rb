#!/usr/bin/env ruby

pkg_path = ARGV[0]
pkg_name = ARGV[1]
pkg_main = ARGV[2]
pkg_vend = ARGV[3]
pkg_url  = ARGV[4]
pkg_info = ARGV[5]
pkg_vers = ARGV[6]

TARGETS = [
  'linux'   => {
    'deb' => ['386', 'amd64', 'arm'],
    'rpm' => ['386', 'amd64'],
  },
]

unless File.directory?("#{pkg_path}")
  Dir.mkdir("#{pkg_path}")
end

TARGETS.each do |target|
  target.each do |os, pkgmap|
    pkgmap.each do |pkg, archs|
      archs.each do |arch|
        pkg_fullname = "#{pkg_path}/go-#{pkg_name}-#{os}-#{arch}"
        system("GOARCH=#{arch} GOOS=#{os} go build -o #{pkg_fullname}")
        system("chmod 755 #{pkg_fullname}")
        system(%{fpm --force \
               -s dir \
               -t #{pkg} \
               -a #{arch} \
               -n #{pkg_name} \
               -p #{pkg_path} \
               --maintainer "#{pkg_main}" \
               --description "#{pkg_info}" \
               --url "#{pkg_url}" \
               --vendor "#{pkg_vend}" \
               --version "#{pkg_vers}" \
               #{pkg_fullname}=/usr/local/bin/#{pkg_name} \
               mount.#{pkg_name}=/sbin/mount.#{pkg_name} 1>&2 2>/dev/null
         })
      end
    end
  end
end

puts Dir["#{pkg_path}/#{pkg_name}*"].join(',')

