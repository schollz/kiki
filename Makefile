.PHONY: scratch, install

server:
	go-bindata static/... templates/...
	go build
	./kiki -no-browser

server2:
	go-bindata static/... templates/...
	go build
	./kiki -no-browser -no-sync

dev:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make

dev2:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server2

scratch:
	cd kikiscratch && browser-sync start --server --files . --index index.html

docs:
	cd doc && make serve