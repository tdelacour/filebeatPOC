all:
	go build -o bin/httpServer httpServer/main.go
	go build -o bin/filebeatTest libFilebeatTest/main.go

clean:
	rm -r bin/*
