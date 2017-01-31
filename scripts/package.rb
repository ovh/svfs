#!/usr/bin/env ruby

require 'erb'

# Package dependencies
DEPENDENCIES = {
  'fuse'  => '> 2.8',
  'ruby'  => '> 1.9.1',
}

# Target platforms and architectures
TARGETS = [
  'linux'   => {
    'deb' => ['i386', 'amd64', 'armhf', 'armel'],
    'rpm' => ['i386', 'amd64'],
  },
  'darwin'  => {
    'pkg' => ['i386', 'amd64']
  },
]

# ARM versions mapping for go build
ARM_VERSIONS = {
  'armhf' => 6,
  'armel' => 5,
}

FILES_LINUX = {
  "scripts/hubic-application.rb" => {
    :target => "/usr/local/bin/hubic-application",
    :mode   => 0755,
  },
  "scripts/mount.svfs" => {
    :target => "/sbin/mount.svfs",
    :mode   => 0755,
  }
}

FILES_MACOS = {
  "scripts/hubic-application.rb" => {
    :target => "/usr/local/bin/hubic-application",
    :mode => 0755,
  },
  "scripts/mount.svfs" => {
    :target => "/usr/local/bin/mount_svfs",
    :mode => 0755,
  }
}

class PackageInfo
  def initialize(version, content_path, template)
    @version = version
    @template = template
    @size = directory_size(content_path)
  end

  def render
    ERB.new(@template).result(binding)
  end

  def save(file)
    File.open(file, "w+") do |f|
      f.write(render)
    end
  end

  # Return directory size in KBytes
  def directory_size(path)
    size=0
    if path.nil?
      return size
    end
    Dir.glob(File.join(path, '**', '*')) { |file| size+=File.size(file) }
    return (size/1024)
  end
end

def build(package, type, version, os, arch, deps)
  # Extra package files
  file_mapping = ""

  if os != "darwin"
    FILES_LINUX.each do |file, spec|
      File.chmod(spec[:mode], file)
      file_mapping << "#{file}=#{spec[:target]} "
    end
  end

  # Dependencies
  go_arch  = arch
  pkg_deps = ""

  deps.each do |pkg, constraint|
    pkg_deps << "-d '#{pkg} #{constraint}' "
  end

  # ARM archs
  if arch.start_with?('arm')
    go_arch = 'arm'
    go_extra = "GOARM=#{ARM_VERSIONS[arch]}"
  end

  if arch == 'i386'
    go_arch = '386'
  end

  go_build_target = "#{package[:path]}/#{package[:name]}-#{os}-#{arch}"
  sh %{CGO_ENABLED=0 GOARCH=#{go_arch} GOOS=#{os} #{go_extra} go build -o #{go_build_target}}
  File.chmod(0755, go_build_target)

  if os == "darwin"
    root_dir = "root-pkg"
    pkg_path = "#{package[:path]}/svfs.pkg"
    bin_path = "#{package[:path]}/#{root_dir}/usr/local/bin"

    mkdir_p bin_path
    mkdir_p pkg_path

    # Generate the payload
    FILES_MACOS.each do |file, spec|
      File.chmod(spec[:mode], file)
      cp "#{file}", "#{package[:path]}/#{root_dir}#{spec[:target]}"
    end
    cp go_build_target, "#{bin_path}/svfs"
    system("( cd #{package[:path]}/#{root_dir} && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > #{pkg_path}/Payload")

    # Generate the package description
    template = File.read("scripts/PackageInfo.erb")
    pkg_info = PackageInfo.new(version, "#{package[:path]}/#{root_dir}", template)
    pkg_info.save("#{pkg_path}/PackageInfo")

    # Generate the Bill Of Materials
    system("mkbom -u 0 -g 80 #{package[:path]}/#{root_dir} #{pkg_path}/Bom")

    # Build the resulting pkg
    system("( cd #{pkg_path} && xar --compression none -cf \"../#{package[:name]}-#{version}-#{arch}.pkg\" * )")

    # Clean
    rm_r(pkg_path, :force => true)
    rm_r("#{package[:path]}/#{root_dir}", :force => true)
  else
    sh %W{fpm
      --force
      -s dir
      -t #{type}
      -a #{arch}
      -n #{package[:name]}
      -p #{package[:path]}
      #{pkg_deps}
      --maintainer "#{package[:maintainer]}"
      --description "#{package[:info]}"
      --license "#{package[:licence]}"
      --url "#{package[:url]}"
      --vendor "#{package[:vendor]}"
      --version "#{version}"
      --deb-use-file-permissions
      --rpm-use-file-permissions
      #{file_mapping}
      #{go_build_target}=/usr/local/bin/#{package[:name]}
    }.join(' ')
  end
end

def gen_dockerfile(version)
  template = File.read("scripts/Dockerfile.erb")
  dockerfile = PackageInfo.new(version, nil, template)
  dockerfile.save("docker/Dockerfile")
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
