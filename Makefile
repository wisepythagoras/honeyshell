.PHONY: all honeyshell vfsutil

all: honeyshell vfsutil

honeyshell:
	$(shell cd cmd/honeyshell; go build .)
	mv cmd/honeyshell/honeyshell .

vfsutil:
	$(shell cd cmd/vfsutil; go build .)
	mv cmd/vfsutil/vfsutil .
