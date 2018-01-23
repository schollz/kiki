# Documentation

<!--- These tags hold related issue numbers. This page's development
is part of #336. --->

## Introduction

- The [KiKi Overview](/overview) document provides a high-level
  introduction to KiKi.
  It is a good place to start to learn about the motivation for the project
  and overall design.
  It also has introductions to many of the other topics explored in more
  detail in the other documents.

- The [FAQ](/faq) answers common questions about KiKi.

## User guide

- The [Signing up a new user](/signup) document describes the process for
  generating keys and registering a user with the KiKi key server.<!--- #326 #210 --->

- The [KiKi Access Control](/access_control) document describes
  KiKi's access control mechanisms. TODO: Break into user-level pieces
  and implementation details; also linked in Architecture below.

- The [KiKi Configuration](/config) document describes KiKi's
  configuration file format and settings.

## Tools

- The [`kiki`](https://godoc.org/kiki.io/cmd/kiki) command is a
  command-line tool for creating and administering KiKi files, users,
  and servers.

- The [`kiki-ui`](https://godoc.org/augie.kiki.io/cmd/kiki-ui) tool
  presents a web interface to the KiKi name space, and also provides a
  facility to sign up an KiKi user and deploy an kikiserver to Google Cloud
  Platform.

- The [`cacheserver`](https://godoc.org/kiki.io/cmd/cacheserver)
  is a client-side directory and storage cache for KiKi.

- The [`kikifs`](https://godoc.org/kiki.io/cmd/kikifs) command
  is a [FUSE](https://en.wikipedia.org/wiki/Filesystem_in_Userspace)
  interface for KiKi.

## Architecture

- The [KiKi architecture](/arch) page has a number of diagrams
  showing, bottom-up, how the pieces all fit together. TODO: add things like keys,
  sharing etc. as diagrams there.<!---  #217 #209 --->

- The [KiKi Access Control](/access_control) document describes
  KiKi's access control mechanisms. TODO: Break into user-level pieces
  and implementation details. TODO: Server-level access control: Writers file etc.

- The [KiKi Security](/security) document describes KiKi's security
  model.

## System setup and administration

- The [Setting up `kikiserver`](/server_setup) document explains how
  to set up your own KiKi installation on a Linux server.<!--- #406 #326 --->

- TODO: Show how to set up with a reverse proxy. <!--- #233 --->

## Programming

- The [`kiki` package](https://godoc.org/kiki.io/kiki) specifies the core
  interfaces that define the KiKi protocol.

- The [`rpc` package](https://godoc.org/kiki.io/rpc) includes a semiformal
  description of the wire protocol used to communicate between clients and
  servers.

- The [`client` package](https://godoc.org/kiki.io/client) provides a
  simple client interface for communicating with KiKi servers.

- TODO: A worked example (implementer's guide).
