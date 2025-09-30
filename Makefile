version := $(shell cat VERSION)

build:
	go build -o oclai

testrelease:
	git tag $(version)
	goreleaser --snapshot --clean --skip=publish
	git tag -d $(version)

release:
	rm -rf dist
	git tag $(version)
	git push origin $(version)
	goreleaser release --clean

checkconfig:
	goreleaser check
