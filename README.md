# KiKi

*KiKi* is a new social network that's different. 

How is *KiKi* different? The main difference is that *the entire social network exists on your computer, all the time*. In *KiKi*, you are part of the cloud. When you use *KiKi* to post a private message to a friend, everyone in the entire network will store that message for you. Secure end-to-end encryption ensures that only your friend can read it, even though everyone has the message. 

## Quickstart

The easiest way to get started is to [download the latest release](https://github.com/schollz/kiki/releases/latest).

Or, if you have Go installed you can build from the source:

```
go get github.com/schollz/kiki
```

and then run:

```
kiki
```

This will start your local server instance and open up a browser so that you can interact with the network.

## Privacy 

Privacy is built-in with end-to-end encryption using public-key cyrptography. Even though you have all message from all other users, you cannot open any of the messages unless you have the key shared from the user to open it. 

Anything you do in *KiKi* is private automatically, unless it is an action (see below) or if it is for **friends**. A **friend** is anyone is followed by you *and* is following you. If you write a post for friends, then only your friends can see it. Every user has friend keys - a set of keys which are specifically for friends. When you unfriend someone you public a new friends key to your friends but not to the unfriended person.

An **action** helps other users determine how to handle your profile and your data. *Actions are always public*. The following are all of the available actions:

- Liking a post
- Blocking a user
- Erasing your profile
- Following a user
- Assigning a profile name
- Assigning a profile description
- Assigning a profile picture

These actions are public to help other users to determine how to handle your data. Some actions are also public allow users to have *Reputation* (e.g. quantified by number of likes on a post, number of followers, etc.). *Reputation* can be useful for validating one's *Identity* but it is also useful to filter.

You can also other posts public by choosing to share with public.

## Storage

The entire social network is stored on your local computer. This allows you to add entries, share photos, and make comments to others offline. Anything you do offline will be synced up later.

There are limits to the amount of content stored on your computer. Those limits are set in the `feed.json` configuration file, typically 5MB for public and 50MB friends. Your own messages never apply to these limits. 

When specified storage limits *are* reached, then messages begin to be purged. Again, only messages from public/friends are allows to be purged. First, old message edits are purged. Then the oldest posts are purged until the free space is no longer being exceeded.

## Features

- Single binary, single settings file, and a setting database.
- Cross-platform (binaries for every OS/architecture)
- Easily federated (just a command line flag to federate)
- You have total control. You can easily edit/delete posts and profiles (unlike Facebook).
- You can comment on posts so that only friends can see (unlike all other social media)
- Storage goes towards content rather than styling (unlike IPFS).


## *"KiKi"*

Why the name *"KiKi"*? *"KiKi"* is [loosely defined](https://en.wikipedia.org/wiki/Kiki_(gathering)) as a gathering of friends for the purpose of chit-chat. Also, *kiki* is half of the [Bouba/kiki effect](https://en.wikipedia.org/wiki/Bouba/kiki_effect) which demonstrates that some aspects of communication and language are universal. Both are underlying principles behind the design and impelementation.



# Technical overview


High level: KiKi is a distributed social network that you can use off the grid, which provides secure end-to-end communication with people inside your social circles.

Medium level: Kiki is like a poor-man's mailservice. Every user in Kiki is a mail carrier. They carry mail for other people and receive mail from other people. Letters are sealed using special encryption that only allows the recipient (friends, public) to read them.

Low level: All the information in the network is stored in **Letters** that are sealed in special cryptographically secure **Envelopes**. In KiKi, you - the are essentially a pair of keys used for [asymetrical cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography). These keys are used to open Envelopes and seal and sign Letters. You interact with with KiKi through a **Feed** that displays the contents of all Letters that you can open. These letters are downloaded to your computer, so you can access the Feed anytime, even off-line.

Everyone on KiKi also carries Letters addressed to others. This makes KiKi distributed, because everyone is a mailman/mail-woman. Once two people encounter each other on LAN or when connecting to a public server, they will exchange Letters that they carried for each other. This ensures that the network can exist as a mesh, outside of the realm of ISPs. It also ensures that no federated servers are necessary, it will work with a few people that can connect onto a local network.

## ABC's

*Assignment allows assesment*. Reputation can be evaluated through arbitrary quantification of certain public aspects of the social network. I.e. - posts that are more popular have a certain amount of "likes", more reputable people have more followers, etc.

*Believing begets buddies*.

*Contributions create connections*. KiKi should be built around providing a frictionless path towards contributing content to be shared, locally, across the network.

## Access

Each social network has a different way of answering the question: *who can access what?* 

In KiKi the access to content is maintained through *four* levels of privacy. These privacy levels are based on whether you and another person are "friends" where "friends" which are two people that mutually agree to share information. 

These levels of privacy are maintained by using public key cryptography. Whenever you transmit information on KiKi, your information will be encrypted by a one or several of your keys, depending on the level of privacy. Here are the privacy levels, in order from private to public: 

1. **Personal**: only you can read (your Personal public key is used).
2. **Special**: you specify which friends can read (your Personal public key + Personal public keys of specified people is used).
3. **Friends**: all your friends can read (your Personal public keys + your Friends public keys is used).
4. **Region**: everyone can read (your Personal public key is used and the Region public key is used).

At a minimum, each user has three key pairs - a *Personal* key pair, a *Friends* and a *Region* key pair. Your Personal and Friends key pairs are unique to you. The Personal key pair identifies you as you, and is used to verify your Letters. The Friends key pair is transmitted to new friends which allows them to open Letters addressed to friends. The Region key pair is built-in to the application. It basically identifies a specific region (think State, City, Nation) in which communication is passed. A person can belong to multiple regions, and everyone must belong to at least one. 



### Personal key pair

Every time you use KiKi, you will seal and sign your posts with your Personal private key. This prevents others from trying to write posts as you and enables you to send private messages to yourself (as in a diary) or to send secure messages directly to others.

You can have only one Personal key. If you have two computers and need to merge the accounts, you can simply copy the Personal keys from one computer (in a file `feed.json`) to the other. All other information is stored in Letter which can be aquired by syncing with anyone.

### Friends key pair

Your first Friends key pair is generated when you start KiKi for the first time. It is emitted as a Letter to yourself, so it automatically syncs between machines.

When you follow someone else and that person also follows you, then you become "friends", and you send that friend the public and private keys for all your Friends key pairs. These keys allow your new friend to view all the posts you have made available to friends.

Unfriending happens. When you unfriend someone by unfollowing them, then you will generate a new Friends key pair. This new Friends key pair will be sent to the remaining friends and will not be sent to the unfriended person. This means that the unfriended person can not read your future messages, but they can still open the past messages (just like in real-life we still retain the memories with that friend up until they leave).

### Region key pair

Every instance of KiKi is configured with a Region key (or several). When you start KiKi it will be able to open any Envelope that it encounters that is meant for any one of your Regions. This allows sub-networks to be formed easily within KiKi, so that you can specify specific places like "United States" or "University of Alberta" to carry around the Envelopes.

## Reading 

The main interaction with KiKi is through reading personal Letters and Letters from friends in the **Feed**. The feed shows only posts that you are accessible to you (i.e. any Letters you can open with your available keys). The feed is generated from all of the opened Letters (see below), sorted by the time/category.

When you start KiKi, the feed is loaded from a locally stored database. KiKi will then check your LAN and registered servers for other KiKi instances. When another KiKi instance is detected, the two computers will sync up their lists of Letters. If these lists differ, then the computer will download each one it is missing, one-by-one, and try to open the Envelope to retrieve the Letter. The computer will use all available keys to try to open it (i.e. the Personal private keys and also the collected Friends private keys). If successfully opened, the Envelope is saved to the disk. Saving onto disk ensures that the feed will be accessible even if you are off-the-grid.

In order to make the network resilient, your computer will also download some Envelopes that cannot be opened by the keys that you hold. Since these can't be opened they will not be available in your feed. These will be stored on your computer, in the possibility that you might meet someone who needs them who will request them from your computer. In this way, you also act as a mailman/mailwoman who is carrying letters for other people. (You can set the limit the amount of space used for storing other people's envelopes by modifying the configuration file. Your personal data does not count towards this limit.)


# Project

## Status

KiKi has rough edges, and is not yet suitable for non-technical users.

[![Build Status](https://travis-ci.org/schollz/KiKi.svg?branch=master)](https://travis-ci.org/schollz/KiKi)


## Contributing

See the [Contribution Guidelines](CONTRIBUTING.md)
for more information on contributing to the project.


### Reporting issues

Please report issues through
[our issue tracker](https://github.com/KiKi/KiKi/issues).


## Community

All KiKi users should subscribe to the
[KiKi Announcements mailing list](https://groups.google.com/forum/#!forum/KiKi-announce)
to receive critical information about the project.

Use the [KiKi mailing list](https://groups.google.com/forum/#!forum/KiKi)
for discussion about KiKi use and development.


### Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CONDUCT.md).
By participating in this project you agree to abide by its terms.
