build:
	go build

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -v ./...
	
smoke-test:
	./testgrid-crawler minikube-presubmits#integration-kvm-containerd-linux-x86
	./testgrid-crawler minikube-periodics#ci-minikube-integration
	./testgrid-crawler minikube-images#post-minikube-gvisor-addon-image
