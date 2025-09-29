version := $(shell cat VERSION)

build:
	go build -o oclai

testrelease:
	git tag $(version)
	goreleaser --snapshot --clean --skip=publish
	git tag -d $(version)

release:
	rm -rf dist/*
	git tag $(version)
	goreleaser --snapshot --clean --skip=publish
