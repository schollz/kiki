<p align="center">
<img
    src=""
    width="100%" border="0" alt="kiki">
<br>
<a href="https://travis-ci.org/schollz/kiki"><img src="https://travis-ci.org/schollz/kiki.svg?branch=master" alt="Build Status"></a>
<a href="https://github.com/schollz/kiki/releases/latest"><img src="https://img.shields.io/badge/version-0.1.0-brightgreen.svg?style=flat-square" alt="Version"></a>
<a href="https://gitter.im/schollz/kiki?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=body_badge"><img src="https://img.shields.io/badge/chat-on%20gitter-brightgreen.svg?style=flat-square" alt="Gitter"></a>
<a href="https://goreportcard.com/report/github.com/schollz/kiki"><img src="https://goreportcard.com/badge/github.com/schollz/kiki" alt="Go Report Card"></a>
</p>

<p align="center"><em>kiki</em> is a new social network that's different. </p>

How is *kiki* different? The main difference is that *the social network exists on your computer, all the time*. In *kiki*, you are part of the cloud. When you use *kiki* to post a private message to a friend, everyone in the network will store that message for you. Secure end-to-end encryption ensures that only your friend can read it, even though everyone has the message. 

## Features

- Single binary (*kiki*), single settings file (*kiki.json*), and a setting database (*kiki.db*, an `sqlite3` db).
- Cross-platform, with [binaries for every OS/architecture](https://github.com/schollz/kiki/releases/latest).
- Easily federated (just a command line flag to federate).
- You have total control. You can easily edit/delete posts and profiles.
- You can comment on posts so that only friends can see .
- Storage goes towards content rather than styling.

Does this already exist? Yes, of course.

Scuttlebutt is the closest, but currently.

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

This will start your local server instance and open up a browser to `localhost:8003` so that you can interact with the network.

# The 50 precepts 

You will be able to understand the design and usage of *kiki* by reading the following 50 precepts.

*Basics*

1. Almost all information in *kiki* is stored in **letters**.
2. A **letter** is defined to have **recipients**, **content**, and a **purpose**.
3. The **purpose** tells *kiki* how to process the letter (e.g. the sharing of an image).
4. The **content** is the data, which depends on the purpose (e.g. its base64 data when sharing an image).
5. The **recipients** is a list of **persons**.
6. A **person** is just a public-private keypair. Your personal keypair is one of two items not stored as a letter. The second item is the **region** keypair.
7. A **region** is a public-private keypair. 
8. Every instance of *kiki* belongs to a **region**. Everyone that belongs to a region has the *region keypair *and uses it to validate their identity.
10. Information is securely transfered in **envelopes**.
11. An **envelope** contains a letter encrypted using the NaCl secret box symmetric cipher with a random passphrase. The random passphrase is then encrypted using the public key of each recipient. Thus, only recipients can decipher the passphrase and unseal the envelope and obtain the contents of the letter.

*Syncing*

12. Two instances of *kiki* are **synced** by exchanging envelopes that they do not have. 
13. Only instances in the same region can sync. Different regions are autonomous, federated instances of *kiki*.
14. As *kiki* network grows, syncing envelopes will obey restrictions on storage - 5MB/person, unless they are a friend (50MB/person) or yourself (no limit). This setting is configurable.
15. Letters whose **purpose** is an *action* are never privy to storage restrictions.

*Purposes*

16. Currently there are two kinds of **purposes** - a *share* and an *action*.
17. A *share* purpose is to share text/html, images (png/jpg), or keys. 
18. A *action* purpose is to create public information for constructing the social network. 
19. Currently available actions are: following, liking, assigning a profile name, assigning a profile, assigning a profile image, blocking someone, erasing a profile. 
20. Actions are made **public** in order to allow quantifying aspects of the social network to have reliable reputation and identity.

*Access*

32. A envelope is sealed using public-private key encryption so that only intended recipients can open it. You are also a recipient of your own letters.
33. A **public** letter is one which is additionally sealed with the *region keypair*. Everyone on the network has this keypair and will be able to unseal the envelope.
34. A **friends** letter is one that is sealed against the latest personal *friends keypair*.
35. The *friends keypair* is generated for each user on initiation.
36. A *friends keypair* is shared upon making a **friend**.
37. A **friend** is someone that you follow, that also follows you.
38. By unfriending, you generate a new *friends keypair* which is transmitted to your remaining friends. Your ex-friend will still see your old content, but not the new content.
39. You can also send a letter addressed to specific people.
.
*Editing and deletion*

40. Every thing on *kiki* is editable. To edit something you create a new letter that identifies the previous letter using **replaces** tag.
41. Only everything on *kiki* is deletable (e.g. your profile). By sending a letter with an action to erase a profile, it will erase everything but that letter. When synced with others, it will also erase your content on everyone elses computer. (__Note__: since letters are signed, you cannot delete someone else's profile).

*The Feed*

42. Your **feed** is a representation of all the envelopes that are accessible to you. 
43. Letters than contain shared images/text are aggregated in reverse-chronological order in a displayed **feed**.
44. You can also hide things from showing up in the feed by editing a post so that its content is empty (effectively deleting it).
45. When editing content, only the latest edit is shown in the feed.
46. All functions of *kiki* are accessible from the feed (e.g. sending letters of various purposes).
47. Even though you have the majority of the envelopes on the network, you can only open ones you have access to.

*The files*

48. The entire program is a single binary: *kiki*.
49. The entire database of envelopes is a single `sqlite3` database: *kiki.db*.
50. The settings file, containing your personal key, is a single file: *kiki.json*.

# Federation

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

kiki has rough edges, and is not yet suitable for non-technical users.

[![Build Status](https://travis-ci.org/schollz/kiki.svg?branch=master)](https://travis-ci.org/schollz/kiki)


## Contributing

See the [Contribution Guidelines](CONTRIBUTING.md)
for more information on contributing to the project.


### Reporting issues

Please report issues through
[our issue tracker](https://github.com/kiki/kiki/issues).


## Community

All kiki users should subscribe to the
[kiki Announcements mailing list](https://groups.google.com/forum/#!forum/kiki-announce)
to receive critical information about the project.

Use the [kiki mailing list](https://groups.google.com/forum/#!forum/kiki)
for discussion about kiki use and development.


### Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CONDUCT.md).
By participating in this project you agree to abide by its terms.

### License

This project is under the MIT license.