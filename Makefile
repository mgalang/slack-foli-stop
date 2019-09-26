.PHONY: build clean deploy

build:
	go build -o bin/busstop .

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
