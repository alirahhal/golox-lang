BINARY_DIR=bin
SRC_NAME=.\main.go

all: run

build:
	@go build -o $(BINARY_DIR)/golox-lang.exe $(SRC_NAME)

run:
	@go run ${SRC_NAME} ${file}

install: go-get

compile:
	@echo "Compiling for every OS and Platform"
	GOOS=freebsd GOARCH=386 go build -o $(BINARY_DIR)/golox-lang-freebsd-386 $(SRC_NAME)
	GOOS=linux GOARCH=386 go build -o $(BINARY_DIR)/golox-lang-linux-386 $(SRC_NAME)
	GOOS=windows GOARCH=386 go build -o $(BINARY_DIR)/golox-lang-windows-386 $(SRC_NAME)