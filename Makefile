.PHONY: clean build

clean:
	rm -rf obj-* debian/.debhelper/ debian/fgofile/ debian/tmp fgofile debian/debhelper-build-stamp debian/fgofile.substvars debian/files

build: clean
	DEB_BUILD_OPTIONS=nocheck dpkg-buildpackage -us -uc
