.PHONY: scratch, install

server:
	go-bindata static/... templates/...
	go build
	./kiki -no-browser

server1:
	go-bindata static/... templates/...
	go build
	./kiki -path 1 -no-browser -internal-port 8003 -external-port 8004

server2:
	go-bindata static/... templates/...
	go build
	./kiki -path 2 -no-browser -internal-port 8005 -external-port 8006

server3:
	go-bindata static/... templates/...
	go build
	./kiki -path 3 -no-browser -no-sync -internal-port 8007 -external-port 8008

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