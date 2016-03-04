# Packaging information
PKG_NAME = svfs
PKG_VEND = OVH
PKG_MAIN = "xavier.lucas@corp.ovh.com"
PKG_INFO = "The Swift Virtual Filesystem"
PKG_VERS := $(shell cat VERSION)
PKG_URL = "https://www.ovh.com"
PKG_LIC = "Apache 2"
PKG_DIR = releases

prepare-release:
	gem install fpm

release:
	scripts/package.rb \
		$(PKG_DIR)     \
		$(PKG_NAME)    \
		$(PKG_MAIN)    \
		$(PKG_VEND)    \
		$(PKG_URL)     \
		$(PKG_INFO)    \
		$(PKG_LIC)     \
		$(PKG_VERS)

.PHONY: prepare-release release
