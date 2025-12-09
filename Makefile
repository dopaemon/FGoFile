.PHONY: clean build

run:
	go run ./main.go

clean:
	rm -rf obj-* debian/.debhelper/ debian/fgofile/ debian/tmp out/ debian/debhelper-build-stamp debian/fgofile.substvars debian/files fgofile

build-deb: clean
	mkdir -p out
	DEB_BUILD_OPTIONS=nocheck dpkg-buildpackage -us -uc
	mv ../fgofile* out/ 2>/dev/null || true

build-binary: clean
	go build -v

up:
	docker compose up --build -d

down:
	docker compose down
