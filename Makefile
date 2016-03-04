# Packaging information
PKG_NAME = svfs
PKG_VEND = OVH
PKG_MAIN = "xavier.lucas@corp.ovh.com"
PKG_INFO = "The Swift Virtual Filesystem"
PKG_VERS := $(shell cat VERSION)
PKG_URL = "https://ovh.com"
PKG_DIR = releases

prepare-release:
	npm install publish-release
	gem install fpm

release:
	$(eval FILES := $(shell ./package.rb $(PKG_DIR) $(PKG_NAME) $(PKG_MAIN) $(PKG_VEND) $(PKG_URL) $(PKG_INFO) $(PKG_VERS)))
	publish-release \
		--token $(TOKEN) \
		--reuseRelease \
		--tag v$(PKG_VERS) \
		--owner xlucas \
		--name "Version $(PKG_VERS)" \
		--repo $(PKG_NAME) \
		--assets $(FILES)

.PHONY: prepare-release release
