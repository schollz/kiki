# Dropbox-specific server setup instructions

These instructions are part of the instructions for
[Setting up `kikiserver`](/doc/server_setup.md).
Please make sure you have read that document first.

## Build `kikiserver-dropbox` and `kiki-setupstorage-dropbox`

To use Dropbox Storage fetch the `dropbox.kiki.io` repository and use the
`kikiserver-dropbox` and `kiki-setupstorage-dropbox` variants.

Fetch the repository and its dependencies:

```
local$ go get -d dropbox.kiki.io/cmd/...
```

Install the `kiki-setupstorage-dropbox` command:

```
local$ go install dropbox.kiki.io/cmd/kiki-setupstorage-dropbox
```

Build the `kikiserver-dropbox` binary:

```
local$ GOOS=linux GOARCH=amd64 go build dropbox.kiki.io/cmd/kikiserver-dropbox
```

## Get an Dropbox authorization code

1. Visit the following [site](https://www.dropbox.com/oauth2/authorize?client_id=wt1281n3q768jj3&response_type=code).
2. If necessary, log in with your Dropbox credentials and click on "Allow".
   <img src="/images/dropbox/allow.png" alt="Allow KiKi storage server to access Dropbox"/>
3. Copy the displayed authorization code.
   <img src="/images/dropbox/code.png" alt="Dropbox API code"/>

Now run the `kiki-setupstorage-dropbox` command and pass the previously copied code as
argument:

```
local$ kiki setupstorage-dropbox -domain=example.com <code>
```

It should produce an output like this:

```
You should now deploy the kikiserver binary and run 'kiki setupserver'.
```


## Notes

All KiKi data will be stored under the app folder `kiki` in the user Dropbox (`/App/kiki`).
Users should [disable syncing](https://www.dropbox.com/lp/pro/pro_onboarding_selective_sync) for
this folder on their Dropbox clients.

## Continue

You can now continue following the instructions in
[Setting up `kikiserver`](/doc/server_setup.md).
