.PHONY: scratch, install

server:
	go-bindata static/... templates/...
	go build
	./kiki -no-browser

server2:
	go-bindata static/... templates/...
	go build
	./kiki -no-browser -no-sync

server3:
	go-bindata static/... templates/...
	go build
	./kiki -path 1 -no-browser -no-sync -private 8005 -public 8006

dev:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make

dev2:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server2

dev3:
	rerun -p "**/*.{go,tmpl,css,js}" --ignore 'bindata.go' make server3

	

scratch:
	cd kikiscratch && browser-sync start --server --files . --index index.html

docs:
	cd doc && make serve