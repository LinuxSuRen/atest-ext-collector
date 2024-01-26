fmt:
	go mod tidy
	go fmt ./...
build:
	go build -o bin/atest-collector .
build-win:
	GOOS=windows go build -o bin/atest-collector.exe .
test:
	go test ./... -cover -v -coverprofile=coverage.out
	go tool cover -func=coverage.out
build-image:
	docker build .
init-env:
	curl https://linuxsuren.github.io/tools/install.sh|bash
	hd i yaml-readme
	hd i cli/cli
	gh repo fork --remote
	gh repo set-default devops-ws/learn-code
