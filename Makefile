BINARY_DIR=bin
SRC_NAME=.\main.go

build:
	@go build -o $(BINARY_DIR)/golanglox.exe $(SRC_NAME)

run:
	@.\$(BINARY_DIR)\golanglox.exe $(testfile)

quickrun:
	@go run ${SRC_NAME} ${testfile}

install: go-get

compile:
	@echo "Compiling for every OS and Platform"
	GOOS=freebsd GOARCH=386 go build -o $(BINARY_DIR)/golanglox-freebsd-386 $(SRC_NAME)
	GOOS=linux GOARCH=386 go build -o $(BINARY_DIR)/golanglox-linux-386 $(SRC_NAME)
	GOOS=windows GOARCH=386 go build -o $(BINARY_DIR)/golanglox-windows-386 $(SRC_NAME)