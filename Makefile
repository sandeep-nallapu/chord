build:
	go build main.go coordinator.go dictionary.go filereader.go hashing.go DHT.go enums.go fingerTable.go jsonHandler.go node.go

configure:
	go get github.com/alexcesaro/log
