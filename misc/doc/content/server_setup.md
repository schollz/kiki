# Setting up kikiserver

## The easy way

The [KiKi tools](/dl/) include a program called `kiki-ui` that automates
the deployment of an `kikiserver` to Google Cloud Platform.
If you wish to deploy to GCP, try using `kiki-ui` instead of following this
guide.
See the [signup document](signup.md) for more information.

## Conventions
Throughout this document, we will mark commands to be run on your
local machine with the shell prompt `local$` and commands to be
run on your server with `server%`.

For example:

```
local$ kiki signup -server=kiki.example.com you@gmail.com
```
and
```
server% sudo systemctl stop kikiserver.service
```

## Introduction
This document describes the process for creating an KiKi installation by deploying
an `kikiserver`, a combined KiKi Store and Directory server, to
a Linux-based machine.

The installation will use the central KiKi key server (`key.kiki.io`) for
authentication, which permits inter-operation with other KiKi servers.

There are multiple versions of `kikiserver`, each depending on where the
associated storage is kept, either on the server's local disk or with a cloud
storage provider.
The binaries that use cloud storage providers each have a suffix that
identifies the provider, such as `kikiserver-gcp` for the Google Cloud
Platform.
These binaries are also kept in distinct repositories, such as `gcp.kiki.io`
for the Google Cloud Platform.

The process follows these steps:

- [sign up](#signup) for an KiKi user account
- [configure](#domain) a domain name and create an KiKi user for the server,
- if necessary, [set up the cloud](#cloud
) storage service,
- [deploy](#deploy) the `kikiserver` to a Linux-based server,
- [configure](#configure) the `kikiserver`.

Each of these steps (besides deployment) has a corresponding `kiki`
subcommand to assist you with the process.

## Prerequisites

To deploy an `kikiserver` you need to decide on values for:

- An Internet domain to which you can add DNS records.
  (We will use `example.com` in this document.)
  Note that the domain need not be dedicated to your KiKi installation; it
  just acts as a name space inside which you can create KiKi users for
  administrative purposes.

- Your KiKi user name (an email address).
  (We will use `you@gmail.com` in this document.)
  This user will be the administrator of your KiKi installation.
  The address may be under any domain,
  as long you can receive mail at that address.

- The host name of the server on which `kikiserver` will run.
  (We will use `kiki.example.com` in this document.)

## Sign up for an KiKi account {#signup}

To register your public key with the central key server run `kiki signup`,
passing your chosen host name as its `-server` argument
and your chosen KiKi user name as its final argument.
Then follow the onscreen instructions.

The [Signing up a new user](/doc/signup.md) document describes this process in
detail.
If you change your mind about the host name, you can update with `kiki user -put`.

## Set up your domain {#domain}

KiKi servers also run as KiKi users, with all the rights and requirements
that demands, and so they need usernames and key pairs registered with the
KiKi key server.
The KiKi user for your server is typically under the domain you are setting up.

You need not use the signup process to create users for your servers.
Instead, the `kiki setupdomain` command will do the work for you.
The `kiki setupdomain` command assumes you want to use `kiki@` followed by
your domain name as your server user name.
(For our example, that's `kiki@example.com`.)

This command sets up users for our example domain:

```
local$ kiki setupdomain -domain=example.com
```

It should produce output like this:

```
Domain configuration and keys for the user
	kiki@example.com
were generated and placed under the directory:
	/home/you/kiki/deploy/example.com
If you lose the keys you can re-create them by running this command
	kiki keygen -secretseed zapal-zuhiv-visop-gagil.dadij-lnjul-takiv-fomin /home/you/kiki/deploy/example.com
Write this command down and store it in a secure, private place.
Do not share your private key or this command with anyone.

To prove that you@gmail.com is the owner of example.com,
add the following record to example.com's DNS zone:

	NAME	TYPE	TTL	DATA
	@	TXT	15m	kiki:aff6a1083da7f1cdb182d43aa3

(Note that '@' here means root, not a literal '@' subdomain).

Once the DNS change propagates the key server will use the TXT record to verify
that you@gmail.com is authorized to register users under example.com.
At a later step, the 'kiki setupserver' command will register your server
user for you automatically.

After that, the next step is to run 'kiki setupstorage' (to configure a cloud
storage provider) or 'kiki setupserver' (if you want to store KiKi data on
your server's local disk).
```

Follow the instructions: place a new TXT field in the `example.com`'s DNS entry
to prove to the key server that you control the DNS records for the domain
`example.com`.
Once the DNS records have propagated, `you@gmail.com` will in effect be
administrator of KiKi's use of `example.com`.

As a guide, here's what the DNS record looks like in Google Domains:

![DNS Entries](https://kiki.io/images/txt_dns.png)

Consult your registrar's documentation if it is not clear how to add a TXT
record to your domain.

Note that some registrars will display the root subdomain name as `@`; you
should not type in the `@` character.

On a Unix machine you can verify that your record is in place (it may take a
few minutes to propagate) by running:

```
local$ host -t TXT example.com
```

Once the TXT record is in place, the key server will permit you to register the
newly-created users that will identify the servers you will deploy (as well as
any other users you may choose to give KiKi user names within `example.com`).
At a later step, the `kiki setupserver` command will register your server
user for you automatically.


## Set up storage and build the `kikiserver` binary

The following sub-sections each describe how to obtain and build a
`kikiserver` binary and set up the storage for a particular location,
such as the server's local disk or a cloud storage provider.

Follow the instructions appropriate for your chosen storage location.

You will need to build an `kikiserver` binary for the server's operating
system and processor architecture.
We will assume 64-bit Linux in this document.


### Local disk

To run off local disk you need to build the `kiki.io/cmd/kikiserver` binary:

```
local$ GOOS=linux GOARCH=amd64 go build kiki.io/cmd/kikiserver
```

The default is to store data in $HOME/kiki/storage.
TODO kiki-setupstorage stuff

**If you choose to store your KiKi data on your server's local disk then
in the event of a disk failure all your KiKi data will be lost.**

### Specific instructions for cloud services {#cloud}

+ [Google Cloud Services](/doc/server_setup_gcp.md)
+ [Google Drive](/doc/server_setup_drive.md)
+ [Amazon Web Services](/doc/server_setup_aws.md)
+ [Dropbox](/doc/server_setup_dropbox.md)

## Set up a server and deploy the `kikiserver` binary {#deploy}

Now provision a server and deploy the `kikiserver` binary to it.

### Provision a server

You can run an `kikiserver` on any server, including Linux, macOS, Windows,
and [more](https://golang.org/doc/install#requirements), as long as it has a
publicly-accessible IP address and can run Go programs.

> Note that KiKi has been mostly developed under Linux and macOS.
> You may encounter issues running it on other platforms.

For a personal KiKi installation, a server with 1 CPU core, 2GB of memory,
and 20GB of available disk space should be sufficient.

If you're using the Google Cloud Platform, you can provision a suitable Linux
VM by visiting the Compute section of the
[Cloud Console](https://cloud.google.com/console) and clicking "Create VM".

> If you're unfamiliar with Google Cloud's virtual machines, here are some sane
> defaults: choose the `n1-standard-1` machine type, select the Ubuntu 16.04
> boot disk image, check "Allow HTTPS traffic", and under "Networking" make
> sure the "External IP" is a reserved static address (rather than
> ephemeral).

Once provisioned, make a note of the server's IP address.

### Create a DNS record

With a server provisioned, you must create a DNS record for its host name.
As you did earlier with the `TXT` record, visit your registrar to create an `A`
record that points your chosen host name (`kiki.example.com`) to the server's
IP address.

### Deploy `kikiserver`

Now deploy your `kikiserver` binary to your server and configure it to run on
startup and serve on port `443`.

You may do this however you like, but you may wish to follow one of these
guides:

- [Running `kikiserver` on Ubuntu 16.04](/doc/server_setup_ubuntu.md)
- (More coming soon...)


## Test connectivity

At this point, you should have an `kikiserver` running on your server in
"setup mode", which means that it is ready to be configured by the `kiki
setupserver` command.
This state is indicated by a log message printed on startup:

```
Configuration file not found. Running in setup mode.
```

Test that the `kikiserver` is accessible from the outside by making an HTTP
request to it. Using your web browser, navigate to the URL of your
`kikiserver` (`https://kiki.example.com/`). You should see the text:

```
Unconfigured KiKi Server
```

If the page fails to load, check the `kikiserver` logs for clues.


## Configure `kikiserver` {#configure}

On your workstation, run `kiki setupserver` to send your server keys and
configuration to the `kikiserver` instance:

```
local$ kiki setupserver -domain=example.com -host=kiki.example.com
```

This registers the server user with the public key server, copies the
configuration files from your workstation to the server, restarts the server
and makes the KiKi user roots for `kiki@example.com` (the server user)
and `you@gmail.com`.

It also creates a special `Group` file for the store server,
`kiki@example.com/Group/Writers`,
whose contents are the names of KiKi users allowed to store data in
the server.
If later you decide to allow more people to use your system, you must update
this file.
See the documentation for `kiki setupwriters` for more information about
this.

It should produce output like this:

```
Successfully put "kiki@example.com" to the key server.
Configured kikiserver at "kiki.example.com:443".
Created root "you@gmail.com".
```

If you make a mistake configuring your server, you can start over by
removing `$HOME/kiki/server` and re-running `kiki setupserver`.
Note that the `$HOME/kiki/server` directory contains your directory server
data, and—if you are using the local disk for storage—any store server objects.
Deleting these files effectively deletes all the data you have put into KiKi.
If you are using a cloud service you may want to delete the contents of your
storage bucket before running `kiki setupserver` again to avoid paying to
store orphaned objects.


## Use your server

You should now be able to communicate with your KiKi installation using the
`kiki` command and any other KiKi-related tools.

To test that you can write and read to your KiKi tree, first create a file:

```
local$ echo Hello, KiKi | kiki put you@gmail.com/hello
```

The `kiki put` command reads data from standard input and writes it to a file
in the root of your KiKi tree named "hello".

Then read the file back, and you should see the greeting echoed back to you.

```
local$ kiki get you@gmail.com/hello
Hello, KiKi
```

If you see the message, then congratulations!
You have successfully set up an `kikiserver`.


## Purging your storage

> TODO: move this to an administrative document.

For a number of reasons, you may wish to discard all your stored data:

1. KiKi is in its early days. As a result we may make incompatible
   changes in the storage or directory formats. This should be rare but
   it may happen.
2. When experimenting with the system, you may create a lot of garbage.
   We hope to have a garbage collector for storage soon, but do not
   have one yet. The only way to clean up is to purge everything and
   start again.
3. Even with a garbage collector, you may find that it is easier to purge
   and restart from scratch than selectively delete files, especially
   when experimenting.

We detail here how to perform the purge if you are running an `kikiserver` on
machine running Ubuntu 16.04 or later.
You will have to tailor these instructions to your own environment
if you are doing something different.

On your server machine, as root, stop the `kikiserver`,
and remove the local server configuration.
This will remove all information about user trees.

```
local$ ssh kiki@kiki.example.com
server% sudo systemctl stop kikiserver.service
server% sudo rm -r ~kiki/kiki/server
```

If you configured your server to use Google Cloud Storage with `kiki
setupstorage-gcp` then you should also purge all references from your storage
bucket.
Run the following command, substituting your own bucket name for
`example-com-kiki`.
(If you have forgotten its name, use `gsutil ls` to list all your bucket names.)
You can do this anywhere you have authenticated as the account used
to set up your Google Cloud instance.

```
local$ gsutil -m rm 'gs://example-com-kiki/**'
```

The `-m` speeds things up by working in parallel.

Now that all your KiKi data has been purged, restart the server.

```
local$ ssh kiki@kiki.example.com
server% sudo systemctl start kikiserver.service
```

Since you have removed its configuration information, the `kikiserver` won't
serve regular KiKi requests until you run `kiki setupserver`.

Reconfigure the server from a host that has your original `$HOME/kiki/deploy`
directory tree.
This gives the server its KiKi keys, the initial contents of its `Writers`
file, and authentication information for accessing cloud storage (if any).

```
local$ kiki setupserver -domain=example.com -host=kiki.example.com
```

Now the server should be ready to use once more.
If you want snapshots, configure them with `kiki snapshot`.
