all: build

PREFIX?=gcr.io/google_containers
FLAGS=
ARCH?=amd64
ALL_ARCHITECTURES=amd64 arm arm64 ppc64le s390x
ML_PLATFORMS=linux/amd64,linux/arm,linux/arm64,linux/ppc64le,linux/s390x
GOLANG_VERSION?=1.7
TEMP_DIR:=$(shell mktemp -d)

VERSION?=v1.3.0-beta.1
GIT_COMMIT:=$(shell git rev-parse --short HEAD)

# You can set this variable for testing and the built image will also be tagged with this name
OVERRIDE_IMAGE_NAME?=

# If this session isn't interactive, then we don't want to allocate a
# TTY, which would fail, but if it is interactive, we do want to attach
# so that the user can send e.g. ^C through.
INTERACTIVE := $(shell [ -t 0 ] && echo 1 || echo 0)
TTY=
ifeq ($(INTERACTIVE), 1)
	TTY=-t
endif

SUPPORTED_KUBE_VERSIONS=1.4.6
TEST_NAMESPACE=heapster-e2e-tests

HEAPSTER_LDFLAGS=-w -X k8s.io/heapster/version.HeapsterVersion=$(VERSION) -X k8s.io/heapster/version.GitCommit=$(GIT_COMMIT)

ifeq ($(ARCH),amd64)
	BASEIMAGE?=busybox
endif
ifeq ($(ARCH),arm)
	BASEIMAGE?=armhf/busybox
endif
ifeq ($(ARCH),arm64)
	BASEIMAGE?=aarch64/busybox
endif
ifeq ($(ARCH),ppc64le)
	BASEIMAGE?=ppc64le/busybox
endif
ifeq ($(ARCH),s390x)
	BASEIMAGE?=s390x/busybox
endif

fmt:
	find . -type f -name "*.go" | grep -v "./vendor*" | xargs gofmt -s -w

build: clean fmt
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(HEAPSTER_LDFLAGS)" -o heapster k8s.io/heapster/metrics
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(HEAPSTER_LDFLAGS)" -o eventer k8s.io/heapster/events

sanitize:
	hooks/check_boilerplate.sh
	hooks/check_gofmt.sh
	hooks/run_vet.sh

test-unit: clean sanitize build
	GOARCH=$(ARCH) go test --test.short -race ./... $(FLAGS)

test-unit-cov: clean sanitize build
	hooks/coverage.sh

test-integration: clean build
	go test -v --timeout=60m ./integration/... --vmodule=*=2 $(FLAGS) --namespace=$(TEST_NAMESPACE) --kube_versions=$(SUPPORTED_KUBE_VERSIONS)

container:
	# Run the build in a container in order to have reproducible builds
	# Also, fetch the latest ca certificates
	docker run --rm -i $(TTY) -v $(TEMP_DIR):/build -v $(shell pwd):/go/src/k8s.io/heapster -w /go/src/k8s.io/heapster golang:$(GOLANG_VERSION) /bin/bash -c "\
		cp /etc/ssl/certs/ca-certificates.crt /build \
		&& GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags \"$(HEAPSTER_LDFLAGS)\" -o /build/heapster k8s.io/heapster/metrics \
		&& GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags \"$(HEAPSTER_LDFLAGS)\" -o /build/eventer k8s.io/heapster/events"

	cp deploy/docker/Dockerfile $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i "s|BASEIMAGE|$(BASEIMAGE)|g" Dockerfile

	docker build --pull -t $(PREFIX)/heapster-$(ARCH):$(VERSION) $(TEMP_DIR)
ifneq ($(OVERRIDE_IMAGE_NAME),)
	docker tag -f $(PREFIX)/heapster-$(ARCH):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif
	rm -rf $(TEMP_DIR)

push: ./manifest-tool gcr-login $(addprefix sub-push-,$(ALL_ARCHITECTURES))
	./manifest-tool push from-args --platforms $(ML_PLATFORMS) --template $(PREFIX)/heapster-ARCH:$(VERSION) --target $(PREFIX)/heapster:$(VERSION)

sub-push-%:
	$(MAKE) ARCH=$* PREFIX=$(PREFIX) VERSION=$(VERSION) container
	docker push $(PREFIX)/heapster-$*:$(VERSION)

influxdb:
	ARCH=$(ARCH) PREFIX=$(PREFIX) make -C influxdb build

grafana:
	ARCH=$(ARCH) PREFIX=$(PREFIX) make -C grafana build

push-influxdb:
	PREFIX=$(PREFIX) make -C influxdb push

push-grafana:
	PREFIX=$(PREFIX) make -C grafana push

./manifest-tool:
	curl -sSL https://github.com/luxas/manifest-tool/releases/download/v0.3.0/manifest-tool > manifest-tool
	chmod +x manifest-tool

gcr-login:
ifeq ($(findstring gcr.io,$(PREFIX)),gcr.io)
	@echo "If you are pushing to a gcr.io registry, you have to be logged in via 'docker login'; 'gcloud docker push' can't push manifest lists yet."
	@echo "This script is automatically logging you in now."
	docker login -u oauth2accesstoken -p "$(shell gcloud auth print-access-token)" https://gcr.io
endif

clean:
	rm -f heapster
	rm -f eventer

.PHONY: all build sanitize test-unit test-unit-cov test-integration container grafana influxdb clean
