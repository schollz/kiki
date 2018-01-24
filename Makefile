.PHONY: scratch, install, basicbuild, server, server1, server2, server3, dev1, dev2, dev3


TAG=$(shell git tag)
HASH=$(shell git log --pretty=format:"%h" -n 1)
LDFLAGS=-ldflags "-s -w -X main.Version=${TAG}-${HASH} -X main.RegionPublic=GoAabW4QeCcyeeDWZxu9wFaPAoWhbrwvrFM83JToWk33 -X main.RegionPrivate=6ptaZoSaepphHTqQyCBRBBRF3WyKGoahXUUTVTL5BAQ3"


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
	./kiki -path 1 -no-browser -port-internal 8003 -port-external 8004 -debug

server2:
	go-bindata static/... templates/...
	go build
	./kiki -path 2 -no-browser -port-internal 8005 -port-external 8006 -debug

server3:
	go-bindata static/... templates/...
	go build
	./kiki -path 3 -no-browser -port-internal 8007 -port-external 8008 -debug

dev1:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server1

dev2:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server2

dev3:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server3

	

scratch:
	cd kikiscratch && browser-sync start --server --files . --index index.html

docs:
	cd doc && make serve
