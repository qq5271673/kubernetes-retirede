all: build

build: clean
	godep go build -a heapster.go
	cp ./heapster ./deploy/docker/heapster

container: build
	sudo docker build -t kubernetes/heapster:canary ./deploy/docker/

clean:
	rm -f heapster
	rm -f ./deploy/docker/heapster

