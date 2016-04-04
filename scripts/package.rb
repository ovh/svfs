#!/usr/bin/env ruby

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


def build(package, type, version, os, arch, deps)
  go_arch  = arch
  pkg_deps = ""

  deps.each do |pkg, constraint|
    pkg_deps << "-d '#{pkg} #{constraint}' "
  end

  if arch.start_with?('arm')
    go_arch = 'arm'
    go_extra = "GOARM=#{ARM_VERSIONS[arch]}"
  end

  go_build_target = "#{package[:path]}/go-#{package[:name]}-#{os}-#{arch}"
  sh %{GOARCH=#{go_arch} GOOS=#{os} #{go_extra} go build -o #{go_build_target}}
  sh %{chmod 755 #{go_build_target}}
  sh %{fpm --force \\
    -s dir \\
    -t #{type} \\
    -a #{arch} \\
    -n #{package[:name]} \\
    -p #{package[:path]} \\
    #{pkg_deps} \\
    --maintainer "#{package[:maintainer]}" \\
    --description "#{package[:info]}" \\
    --license "#{package[:licence]}" \\
    --url "#{package[:url]}" \\
    --vendor "#{package[:vendor]}" \\
    --version "#{version}" \\
    #{go_build_target}=/usr/local/bin/#{package[:name]} \\
    scripts/mount.#{package[:name]}=/sbin/mount.#{package[:name]} \\
    scripts/hubic-application.rb=/usr/local/bin/hubic-application
  }

end

def release(package, version)
  TARGETS.each do |target|
    target.each do |os, pkgmap|
      pkgmap.each do |pkg, archs|
        archs.each do |arch|
          build(package, pkg, version, os, arch, DEPENDENCIES)
        end
      end
    end
  end
end


