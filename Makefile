.PHONY: scratch, install

server:
	go-bindata static/... templates/...
	go build
	./kiki -no-sync -no-browser

dev:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make

scratch:
	cd kikiscratch && browser-sync start --server --files . --index index.html

docs:
	cd doc && make serve