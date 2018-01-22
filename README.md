<p align="center">
<img
    src="https://user-images.githubusercontent.com/6550035/35201555-7014440a-feea-11e7-9f21-7fe831a35768.png"
    width="80%" border="0" alt="kiki">
<br>
<a href="https://travis-ci.org/schollz/kiki"><img src="https://travis-ci.org/schollz/kiki.svg?branch=master" alt="Build Status"></a>
<a href="https://github.com/schollz/kiki/releases/latest"><img src="https://img.shields.io/badge/version-0.1.0-brightgreen.svg?style=flat-square" alt="Version"></a>
<a href="https://kiki.network/?hashtag=kikihelp"><img src="https://img.shields.io/badge/chat-on%20kiki-brightgreen.svg?style=flat-square" alt="Kiki"></a>
<a href="https://goreportcard.com/report/github.com/schollz/kiki"><img src="https://goreportcard.com/badge/github.com/schollz/kiki" alt="Go Report Card"></a>
</p>

<p align="center"><em>kiki</em> is an experimental social network. </p>

How is *kiki* different from other social networks? The main difference is that *the social network exists on your computer, all the time*. This means *kiki* will work offline, and it means there's nobody tracking your browsing. 

In *kiki*, you are part of the cloud. When you use *kiki* to post a private message to a friend, everyone in the network will store that message for you. Secure end-to-end encryption ensures that only your friend can read it, even though everyone has the message. 

_Note:_ This software is experimental at the moment. It uses end-to-end encryption so it *should be secure*, but the codebase has not been audited so do not post a bank statement.


## Features

*kiki* is heavily inspired by [Patchwork](https://github.com/ssbc/patchwork), but aims to improve some facets such as: simple federation, multi-computer logins, add in post editing/deletion and even profile deletion. Here are the main features of *kiki*:

- You can use *kiki* offline.
- You have all the data, all the time. 
- You have total control of your posts - you can easily edit/delete posts and profiles.
- You can comment on posts so that only friends can see.
- Storage goes towards content rather than styling.
- Single binary (*kiki*), single settings file (*kiki.json*), and a setting database (*kiki.db*, an `sqlite3` db).
- Cross-platform, with [binaries for every OS/architecture](https://github.com/schollz/kiki/releases/latest).
- Easily federated (just a command line flag to federate).
- Multi-machine use - just transfer your *kiki.json* settings file to each computer you want to use on.


## Quickstart

The easiest way to get started is to [download the latest release](https://github.com/schollz/kiki/releases/latest).

Or, if you have Go installed you can build from the source:

```
go get -u github.com/jteeuwen/go-bindata/...
go get github.com/schollz/kiki
cd $GOPATH/src/github.com/schollz/kiki
go-bindata static/... templates/...
go build -v
```

and then run:

```
kiki
```

This will start your local server instance and open up a browser to `localhost:8003` so that you can interact with the network. Right now, to sync you can add another open server (currently the only available one is https://kiki.network, but you can make your own).

# The 50 precepts 

You will be able to understand the design and usage of *kiki* by reading the following 50 precepts.

*Fundamentals*

1. Information in *kiki* is stored in **letters**.
2. A **letter** is defined to have **recipients**, **content**, and a **purpose**.
3. The **purpose** specifies how a letter is processed (e.g. whether the letter is an image to be shared, or the liking of a post, etc.).
4. The **content** is the data, which depends on the purpose (e.g. its base64 data when sharing an image, or the ID of the post if liking, etc.).
5. The **recipients** is a list of **persons**.
6. A **person** is just a public-private keypair. Your personal keypair is one of two items not stored as a letter. The second item is the **region** keypair.
8. Every instance of *kiki* belongs to a **region**. Everyone that belongs to a region has the **region keypair** that is used to validate identities.
7. A **region keypair** is a public-private keypair that is shared by everyone.
10. Information is securely transfered in **envelopes**.
11. An **envelope** contains a letter encrypted using the NaCl secret box symmetric cipher with a random passphrase. The random passphrase is then encrypted using the public key of each recipient. Thus, only recipients can decipher the passphrase and unseal the envelope and obtain the contents of the letter. The sender signs the envelope using the region keypair and their personal keypair to verify authenticity.

*Syncing*

12. Two instances of *kiki* are **synced** by exchanging envelopes that they do not have. 
13. Only instances in the same region can sync. Different regions are autonomous, federated instances of *kiki*.
14. As *kiki* network grows, syncing envelopes will obey restrictions on storage - 5MB/person, unless they are a friend (50MB/person) or yourself (no limit). This setting is configurable.
15. Letters whose **purpose** is an *action* are never privy to storage restrictions.

*Purposes*

16. Currently there are two kinds of **purposes** - a *share* and an *action*.
17. A **share** purpose is to share text/html, images (png/jpg), or keys. 
18. A **action** purpose is to create public information for constructing the social network. 
19. Currently available actions are: following, liking, assigning a profile name, assigning a profile, assigning a profile image, blocking someone, erasing a profile. 
20. Actions are made **public** in order to allow quantifying aspects of the social network to have reliable reputation and identity.

*Access*

32. A envelope is sealed using public-private key encryption so that only intended recipients can open it. You are also a recipient of your own letters.
33. A public letter is one which is additionally sealed with the *region keypair*. Everyone on the network has this keypair and will be able to unseal the envelope.
34. A letter for a **friend** is one that is sealed against the latest personal *friends keypair.
35. A **friend** is someone that you follow, that also follows you.
36. The *friend keypair* from each friend are shared upon making a **friend**.
36. The *friends keypair* is just a keypair that is generated for each user on initiation, that allows friends to decrypt your messages.
38. By unfriending, you generate a new *friends keypair* which is transmitted to your remaining friends. Your ex-friend will still see your old content, but not the new content.
39. You can also send a letter addressed to specific people by specifying their public keys.

*Editing and deletion*

40. Every thing on *kiki* is editable. To edit something you create a new letter that identifies the original letter using a *first_id* tag.
41. You can delete anything you made on *kiki*. By sending a letter with an action to erase a profile, it will erase everything but that letter. When synced with others, it will also erase your content on everyone elses computer. (_Note_: since letters are signed, you cannot delete someone else's profile).

*The Feed*

42. Your **feed** is a representation of all the envelopes that are accessible to you (i.e. addressed to you, addressed to friends, or addressed to public).
43. The representation of letters is most generally a website where shared images/text are aggregated in reverse-chronological order in a displayed **feed**. (_Note_: *kiki* is not a website - it is an infrastructure. Feel free to build your own display).
44. You can also hide things from showing up in the feed by editing a post so that its content is empty (effectively deleting it).
45. When editing content, only the latest edit is shown in the feed.
46. All functions of *kiki* are accessible from the feed (e.g. sending letters of various purposes).
47. Even though you have the majority of the envelopes on the network, you can only open ones you have access to.

*The files*

48. The settings file, containing your personal key, is a single file: *kiki.json*. *You can transfer this file to any computer to reconstitute your entire social network!*
49. The entire program is a single binary: *kiki*.
50. The entire database of envelopes is a single `sqlite3` database: *kiki.db*.

# Neat features

## Make your own sync hub

Currently the only public syncing up is https://kiki.network. To make your own, just start up a new instance of *kiki* and reverse proxy to the external port (port `8004` by default). Other instances will be able to exchange with this server but will not be able to modify the user data (which is only accessible via the private port, `8003` by default).


## Multi-computer user

Since everything on *kiki* is stored in a cloud, you can use *kiki* on multiple computers by just transfering your key file - `kiki.json` to another computer (by default at `.kiki/kiki.json`). Once you re-connect to a hub, it will download and parse your entire feed and recapitulate everything you had before! 

Since letters are bagged (and not appended to a log) you can have multiple instances out-of-sync without causing any problems.

## Federation

By federating, you will have your network of *kiki* instances which can only communicate among themselves. Only people that have been given the region keys will be able to join this federated system.

To federate your own system simple run:

```
kiki -generate-region
```

This will generate a unique public-private keypair for you to use and a way to start *kiki* to utilize this region instead of the default, e.g.:

```
kiki -region-public 'X' -region-private 'Y'
```

# Project

## Status

*kiki* is in alpha status. You can use it, but breaking changes might still occur. *kiki* has rough edges, and is not yet suitable for non-technical users.

[![Build Status](https://travis-ci.org/schollz/kiki.svg?branch=master)](https://travis-ci.org/schollz/kiki)


## Contributing

Please contribute! Try *kiki* out, ask questions, submit PRs. Anything is welcome.

### Reporting issues

Please report issues through
[our issue tracker](https://github.com/kiki/kiki/issues).


## Community

We use *kiki* for development and questions. For development, check out [#kikidev](http://localhost:8003/?hashtag=kikidev) and for general help checkout [#kikihelp](http://localhost:8003/?hashtag=kikihelp).


### Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CONDUCT.md).
By participating in this project you agree to abide by its terms.

### License

This project is under the MIT license.

The Kiki mascot is Copyright 2018 Jessie Doyle and Cloud Supernova.
