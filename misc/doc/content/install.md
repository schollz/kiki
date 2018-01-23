# Install KiKi

## The easy way

You can download the [latest release on Github](https://github.com/schollz/kiki/releases/latest).

## The hard way

First make sure you installed [Go](). Then you can use `go get` to download the latest:

```bash
$ go get -u -v github.com/schollz/kiki
```

which will download dependencies and generate a binary in `$GOPATH/bin`.

When you run KiKi, it will load the default configuration file stored in `$HOME/.kiki/config.yaml`. Here is a default configuration file:

```
# Specify how much disk spaces is used by Envelopes cannot be opened
PublicEnvelopedDiskSpace: 20MB
# Specify whether to store opened Letters in memory
StoreLettersInMemory: false
```


