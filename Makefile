.PHONY: all honeyshell vfsutil tests

all: honeyshell vfsutil

honeyshell:
	$(shell cd cmd/honeyshell; go build .)
	mv cmd/honeyshell/honeyshell .

vfsutil:
	$(shell cd cmd/vfsutil; go build .)
	mv cmd/vfsutil/vfsutil .

# This is meant for experiments which are not committed to the repo
tests:
	$(shell cd cmd/tests; go build .)
	mv cmd/tests/tests .
