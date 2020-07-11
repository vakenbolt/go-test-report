genbuild: gencode
	go build

gencode:
	(cd embed_assets/;set -e;go build;./embed_assets)

buildall: genbuild
	echo "Building..."

	mkdir -p release_builds/linux-amd64/
	mkdir -p release_builds/darwin-amd64/
	mkdir -p release_builds/windows-amd64/

	echo "Linux 64bit"
	GOOS=linux GOARCH=amd64 go build -o release_builds/linux-amd64/
	(cd release_builds/linux-amd64/; tar -czf go-test-report-linux-amd64.tgz go-test-report)

	echo "Darwin (MacOS) 64bit"
	GOOS=darwin GOARCH=amd64 go build -o release_builds/darwin-amd64/
	(cd release_builds/darwin-amd64/; tar -czf go-test-report-darwin-amd64.tgz go-test-report)
	(cd release_builds/darwin-amd64/; tar -czf go-test-report-darwin-amd64.tgz go-test-report)

	echo "Windows 64bit"
	GOOS=windows GOARCH=amd64 go build -o release_builds/windows-amd64/
	(cd release_builds/windows-amd64/; zip -r go-test-report-windows-amd64.zip go-test-report.exe)

	echo "...Done!"
