gencode:
	(cd embed_assets/;set -e;go build;./embed_assets)

genbuild: gencode
	go build

buildall: genbuild
	echo "Building..."

	mkdir -p release_builds/linux-amd64/
	mkdir -p release_builds/darwin-amd64/
	mkdir -p release_builds/windows-i386/
	mkdir -p release_builds/windows-amd64/

	echo "Linux 64bit"
	GOOS=linux GOARCH=amd64 go build -o release_builds/linux-amd64/

	echo "Darwin (MacOS) 64bit"
	GOOS=darwin GOARCH=amd64 go build -o release_builds/darwin-amd64/

	echo "Windows 32bit"
	GOOS=windows GOARCH=386 go build -o release_builds/windows-i386/

	echo "Windows 64bit"
	GOOS=windows GOARCH=amd64 go build -o release_builds/windows-amd64/

	echo "...Done!"