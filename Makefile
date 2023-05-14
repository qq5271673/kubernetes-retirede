all: build

build: clean
	godep go build -a ./...
	godep go build -a
	cp ./heapster ./deploy/docker/heapster

test-unit: clean
	godep go build -a ./...
	godep go test -test.short ./...

container: build
	sudo docker build -t heapster:canary ./deploy/docker/

clean:
	rm -f heapster
	rm -f ./deploy/docker/heapster

