.PHONY:all clean

all: bin/windows_amd64/hitminer-file-manager.exe bin/linux_amd64/hitminer-file-manager bin/darwin_amd64/hitminer-file-manager bin/darwin_arm64/hitminer-file-manager

bin/windows_amd64/hitminer-file-manager.exe:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/windows_amd64/hitminer-file-manager.exe

bin/linux_amd64/hitminer-file-manager:
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o bin/linux_amd64/hitminer-file-manager

bin/darwin_amd64/hitminer-file-manager:
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o bin/darwin_amd64/hitminer-file-manager

bin/darwin_arm64/hitminer-file-manager:
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o bin/darwin_arm64/hitminer-file-manager

clean:
	rm -rf ./bin