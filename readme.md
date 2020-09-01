Google Photos API cannot be used to backup due to [numerous bugs](https://github.com/gilesknap/gphotos-sync/issues/119).

[Google Takeout](https://sites.google.com/site/picasaresources/Home/Picasa-FAQ/google-photos-1/how-to/how-to-download-all-autobackupped-pictures#TOC-Download-using-Google-Takeout) is the only way to backup Google Photos, but directory layout of downloaded archive is not convenient. This tool layouts and deduplicate files (as files in albums duplicates files instead of linking).

Status: *alpha*. Only macOS and Linux are supported (because [hard links](https://en.wikipedia.org/wiki/Hard_link) are used).

Not implemented yet:

 * Store albums. Albums info is not copied for now, but should be or symlinks created, or some meta-file. That's why for now hard-links are used — keep original takeout directory.
 * Update EXIF data (geolocation from Google). Sometimes you don't encode geolocation into files, but Google infers it from location history.

## Install

[Download](https://github.com/develar/gphotos-takeout/releases/latest) or install from sources:

```sh
GO111MODULE=on go get github.com/develar/gphotos-takeout
```

## Usage

```sh
gphotos-takeout -i photos-takeout -o photos
```