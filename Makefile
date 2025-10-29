.PHONY: clean build

clean:
	rm -rf obj-* debian/.debhelper/ debian/fgofile/ debian/tmp out/ debian/debhelper-build-stamp debian/fgofile.substvars debian/files

build: clean
	mkdir -p out
	DEB_BUILD_OPTIONS=nocheck dpkg-buildpackage -us -uc
	mv ../fgofile* out/ 2>/dev/null || true
