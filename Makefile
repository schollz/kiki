.PHONY: scratch, install

server:
	go-bindata static/... templates/...
	go build -no-browser -no-sync
	./kiki

dev:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make

scratch:
	cd kikiscratch && browser-sync start --server --files . --index index.html

docs:
	cd doc && make serve