.PHONY: scratch, install, basicbuild, server, server1, server2, server3, dev1, dev2, dev3


HASH=$(shell git describe)
LDFLAGS=-ldflags "-s -w -X main.Version=${HASH} -X main.RegionPublic=4NfD9kWESGycUdbhbrFygNDjFun6NPk6utpkviyE1Ai6 -X main.RegionPrivate=btbsjnjTtgi3aL9z2X8bqb1URVnCo3zqg4fC4co2JEu"


basicbuild:
	go-bindata static/... templates/...
	go build ${LDFLAGS}

update:
	dep ensure -v
	dep ensure -v -update
	# remove things that import "testing" so that the flags are not included
	rm -rf vendor/github.com/blevesearch/bleve/index/store/test
	rm -rf vendor/golang.org/x/text/internal/testtext/
	rm -rf vendor/golang.org/x/net/nettest
	rm -rf vendor/github.com/blevesearch/go-porterstemmer/porterstemmer_has_suffix.go
	go build -v -a
	
release:
	docker pull karalabe/xgo-latest
	go get github.com/karalabe/xgo
	mkdir -p bin 
	xgo -go "1.9.2" -dest bin ${LDFLAGS} -targets linux/amd64,linux/arm-6,darwin/amd64,windows/amd64 github.com/schollz/kiki
	# cd bin && upx --brute kiki-linux-amd64

server:
	go-bindata static/... templates/...
	go build
	./kiki -no-browser

server1:
	go-bindata static/... templates/...
	go build
	./kiki -alias 1 -path testing -no-browser -port-internal 8003 -port-external 8004 -debug

server2:
	go-bindata static/... templates/...
	go build
	./kiki -alias 2 -path testing -no-browser -port-internal 8005 -port-external 8006 -debug

server3:
	go-bindata static/... templates/...
	go build
	./kiki -alias 3 -path testing -no-browser -port-internal 8007 -port-external 8008 -debug

server4:
	go-bindata static/... templates/...
	go build
	./kiki -alias 3 -path testing -no-browser -port-internal 8009 -port-external 8010 -debug -region-public 'BdmcuwCLEEhVzfWmzoe6CqRHxTWWfiXHx3bY8mWn2ueH' -region-private '7nQ4t2vqTkLg4uUHarWZjTJYDFqVM9qMq2ie3erGtTQJ'


dev1:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server1

dev2:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server2

dev3:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server3

dev4:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server4

	

scratch:
	cd kikiscratch && browser-sync start --server --files . --index index.html

docs:
	cd doc && make serve
