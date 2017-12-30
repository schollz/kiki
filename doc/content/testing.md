# Notes on testing `kiki`

These instructions are to guide you for testing the internals of kiki.

## Introduction

Go into the kiki directory

```
go get github.com/pilu/fresh
cd $GOPATH/src/github.com/schollz/kiki
fresh
```

Install `httpie` if you don't have it already

```
sudo -H python3 -m pip install httpie
```

Capture a new identity

```
http localhost:8003/identity > identity.json
```

Put in a new envelope:

```
http -f POST localhost:8003/letter identity@./identity.json message@./message.txt public=yes recipients='["1CppUCZT1_RBGnl9WrFWHxmV9scuHQykoiNuZayEfAU"]'
```

Open envelopes (and regenerate feed)

```
http -f POST localhost:8003/open identity@./identity.json
```

Assign name

```
http -f POST localhost:8003/assign identity@./identity.json assign=name data=zack
```