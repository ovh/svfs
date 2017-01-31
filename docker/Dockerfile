FROM ubuntu:latest
MAINTAINER Valentin Pichard <valentin.pichard@corp.ovh.com>

# Get the last package of svfs and set workdir
ADD https://github.com/ovh/svfs/releases/download/v0.8.1/svfs_0.8.1_amd64.deb /root/build/svfs_0.8.1_amd64.deb
WORKDIR /root/build

# Update system, get dependencies and install svfs package
RUN apt-get update && apt-get -y install ruby fuse && apt-get clean && dpkg -i svfs_0.8.1_amd64.deb

ENTRYPOINT ["mount -t svfs -o"]

CMD ["--help"]