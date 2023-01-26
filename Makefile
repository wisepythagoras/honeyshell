.PHONY: all honeyshell

all: honeyshell

honeyshell:
	$(shell cd cmd/honeyshell; go build .)
	mv cmd/honeyshell/honeyshell .
