VERSION_REGEX  := 's/(v[0-9\.]+)/$(version)/g'

build:
	go build -o smug *.go

test:
	go test ./pkg/commander/ ./pkg/config/ ./pkg/context/  ./pkg/smug/ ./pkg/tmux/

coverage:
	go test -coverprofile=coverage.out ./pkg/commander/ ./pkg/config/ ./pkg/context/  ./pkg/smug/ ./pkg/tmux/
	go tool cover -html=coverage.out

release:
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is not defined)
endif
	sed -E -i.bak $(VERSION_REGEX) 'main.go' && rm main.go.bak
	git commit -am 'Update version to $(version)'
	git tag -a $(version) -m '$(version)'
	git push origin $(version)
	goreleaser --rm-dist
