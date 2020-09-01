.PHONY: update-deps

update-deps:
	GOPROXY=https://proxy.golang.org go get -d -u ./...
	go mod tidy

lint:
	golangci-lint run

build:
	goreleaser --snapshot --skip-publish --rm-dist

#exiftool -recurse -dateFormat %s -tagsfromfile "%d/%F.json" \
#"-GPSAltitude<geodataaltitude" "-gpslatitude<geodatalatitude" "-gpslatituderef<geodatalatitude" "-gpslongitude<geodatalongitude" "-gpslongituderef<geodatalongitude" \
#-overwrite_original -preserve -progress \
#-ext "*"  --ext json photos