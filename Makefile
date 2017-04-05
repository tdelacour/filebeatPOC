all: bin server filebeatTest

bin:
	mkdir -p bin

server: bin
	go build -o bin/httpServer httpServer/main.go

filebeatTest: bin
	go build -o bin/filebeatTest libFilebeatTest/main.go

runServer:
	./bin/httpServer

runFilebeat:
	./bin/filebeatTest

clean:
	rm -r bin
	rm registry
