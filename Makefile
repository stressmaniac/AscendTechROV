all: buildstatic buildgo

buildstatic:
	$(MAKE) -C static all

buildgo:
	$(MAKE) -C go all

clean:
	$(MAKE) -C static clean
	$(MAKE) -C go clean

test:
	$(MAKE) -C static test
	$(MAKE) -C go test
