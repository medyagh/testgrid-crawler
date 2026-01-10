build:
	go build

smoke-test:
	./test-grid-crawler minikube-presubmits#integration-kvm-containerd-linux-x86
	./test-grid-crawler minikube-periodics#ci-minikube-integration
	./test-grid-crawler minikube-images#post-minikube-gvisor-addon-image
