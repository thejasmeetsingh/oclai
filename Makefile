version := $(shell cat VERSION)

build:
	go build -o oclai

testrelease:
	git tag v$(version)
	goreleaser --snapshot --clean --skip=publish
	git tag -d v$(version)

release:
	rm -rf dist
	git tag v$(version)
	git push origin v$(version)
	goreleaser release --clean

checkconfig:
	goreleaser check
