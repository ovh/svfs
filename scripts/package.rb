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
  'darwin'  => {
    'pkg' => ['386', 'amd64']
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
    :target => "root-pkg/usr/local/bin/hubic-application",
    :mode => 0755,
  },
  "scripts/mount.svfs" => {
    :target => "root-pkg/usr/local/bin/mount_svfs",
    :mode => 0755,
  },
  "scripts/PackageInfo.erb" => {
    :target => "svfs.pkg/PackageInfo.erb",
    :mode => 0644,
  }
}

class PackageInfo
  attr_accessor :version, :root_path, :path

  def initialize(version, root_path, path)
    @version = version
    @path = path
    @template = File.read("#{@path}/PackageInfo.erb")
    @size = self.class.directory_size(root_path)
  end

  def render
    ERB.new(@template).result(binding)
  end

  def save(file)
    File.open(file, "w+") do |f|
      f.write(render)
    end
    FileUtils.rm_f("#{path}/PackageInfo.erb")
  end

  # Return directory size in KBytes
  def self.directory_size(path)
    size=0
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

  go_build_target = "#{package[:path]}/go-#{package[:name]}-#{os}-#{arch}"
  sh %{GOARCH=#{go_arch} GOOS=#{os} #{go_extra} go build -o #{go_build_target}}
  File.chmod(0755, go_build_target)

  if os == "darwin"
    root_path = "#{package[:path]}/root-pkg/usr/local/bin"
    mkdir_p root_path
    pkg_path = "#{package[:path]}/svfs.pkg"
    mkdir_p pkg_path
    FILES_MACOS.each do |file, spec|
      File.chmod(spec[:mode], file)
      cp "#{file}", "#{package[:path]}/#{spec[:target]}"
    end
    cp go_build_target, "#{root_path}/svfs"
    pkg_info = PackageInfo.new(version, "#{package[:path]}/root-pkg", "#{pkg_path}")
    pkg_info.save("#{pkg_path}/PackageInfo")
    system("( cd #{package[:path]}/root-pkg && find . | cpio -o --format odc --owner 0:80 | gzip -c ) > #{pkg_path}/Payload")
    system("mkbom -u 0 -g 80 #{package[:path]}/root-pkg #{pkg_path}/Bom")
    system("( cd #{pkg_path} && xar --compression none -cf \"../#{package[:name]}-#{version}-#{arch}.pkg\" * )")
    rm_r(pkg_path, :force => true)
    rm_r("#{package[:path]}/root-pkg", :force => true)
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

  File.delete(go_build_target)
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
