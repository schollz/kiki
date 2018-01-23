# GCP-specific server setup instructions

These instructions are part of the instructions for
[Setting up `kikiserver`](/doc/server_setup.md).
Please make sure you have read that document first.

## Build `kikiserver-gcp` and `kiki-setupstorage-gcp`

To use Google Cloud Storage fetch the `gcp.kiki.io` repository and use the
`kikiserver-gcp` and `kiki-setupstorage-gcp` variants.

Fetch the repository and its dependencies:

```
local$ go get -d gcp.kiki.io/cmd/...
```

Install the `kiki-setupstorage-gcp` command:

```
local$ go install gcp.kiki.io/cmd/kiki-setupstorage-gcp
```

Build the `kikiserver-gcp` binary:

```
local$ GOOS=linux GOARCH=amd64 go build gcp.kiki.io/cmd/kikiserver-gcp
```

## Create a Google Cloud Project

First create a Google Cloud Project and associated Billing Account by visiting the
[Cloud Console](https://cloud.google.com/console).
See the corresponding documentation
[here](https://support.google.com/cloud/answer/6251787?hl=en) and
[here](https://support.google.com/cloud/answer/6288653?hl=en)
for help.
For the project name, we suggest you use a string similar to your domain.
(We will use `example-com`.)

Then, install the Google Cloud SDK by following
[the official instructions](https://cloud.google.com/sdk/downloads).

Finally, use the `gcloud` tool to enable the required APIs:

```
local$ gcloud components install beta
local$ gcloud config set project example-com
local$ gcloud auth login
local$ gcloud beta services enable iam.googleapis.com
local$ gcloud beta services enable storage_api
```

## Create a Google Cloud Storage bucket

Use the `gcloud` tool to obtain "application default credentials" so that the
`kiki setupstorage-gcp` command can make changes to your Google Cloud Project:

```
local$ gcloud auth application-default login
```

Now use `kiki setupstorage-gcp` to create a storage bucket and an associated
service account for accessing the bucket.
Note that the bucket name must be globally unique among all Google Cloud
Storage users, so it is prudent to include your domain name in the bucket name.
(We will use `example-com-kiki`.)

```
local$ kiki setupstorage-gcp -domain=example.com -project=<project> example-com-kiki
```

It should produce output like this:

```
Service account "kikistorage@example-com.iam.gserviceaccount.com" created.
Bucket "example-com-kiki" created.
You should now deploy the kikiserver binary and run 'kiki setupserver'.
```

## Continue

You can now continue following the instructions in
[Setting up `kikiserver`](/doc/server_setup.md).
