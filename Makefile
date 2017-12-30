.PHONY: scratch, install

server:
	fresh
	
scratch:
	cd kikiscratch && browser-sync start --server --files . --index index.html

docs:
	cd doc && make serve