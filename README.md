# Semantic ID

## Overview

This library provides support for creating semantically meaningful identifiers for use in SOA environment.

Semantically meaningful identifiers (further referred simply as 'semantic IDs) allow encoded representation of internal IDs used within the service to be translated to and from opaque string IDs that conceal the nature of internal identifiers and thus allow flexibility of internal implementation (e.g. change from internal ID to external and vice versa).

Also semantic IDs allow easier debugging, i.e. when referred from system logs, developers can easily understand what service/entity particular ID refers to.

## Representation

Semantic ID includes 3 parts:

* Encoded service name and version, e.g. 'foo1'
* Optional, encoded entity name, e.g. 'user'
* Encoded ID, e.g. 'q6b60'

Examples of such IDs: ``foo1-user-q6b60``, ``req-9j35zf3``.

Notes:

* ID are case insensitive, i.e. semantic ID ``a-b-1cd`` should always be equivalent to ``A-B-1CD`` and vice versa.
* Encoded ID prefixes that are used to construct an instance of codec are always required when decoding an ID. E.g. ``foo1-qw1`` is a valid semantic ID (with service name ``foo`` and version ``1``), ``qw1`` is not.
  * Note, that ID may be created without any prefixes, in this case only ID body is required for decoding


