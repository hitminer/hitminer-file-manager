.PHONY:windows_amd64 linux_amd64 darwin_amd64 darwin_arm64

windows_amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o hitminer-file-manager.exe

linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

darwin_amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build

darwin_arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build