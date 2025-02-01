<!--
SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>

SPDX-License-Identifier: CC0-1.0
-->

# AWS Neuron Driver - Publish Source

Disclaimer: This is a personal project and not related to,
or endorsed by, Amazon Web Services.

This repository contains a command line tool, written in Go,
to publish the source code of new AWS Neuron Driver releases.

Why? The driver is licensed under GPL-2.0 but currently
only distributed as rpm package with a DKMS wrapper.
The code is not available as archive or in a public git repository.

The tool is used to add releases and code updates to the unofficial
[aws-neuron-driver](https://git.sr.ht/~wombelix/aws-neuron-driver)
repository. Checksum and GPG verifications are performed and metadata added.
This creates an audit trail and allows to validate that the code is coming
from the official repository
[yum.repos.neuron.amazonaws.com](https://yum.repos.neuron.amazonaws.com/)
and wasn't altered.

[![REUSE status](https://api.reuse.software/badge/git.sr.ht/~wombelix/aws-neuron-driver-publish-source)](https://api.reuse.software/info/git.sr.ht/~wombelix/aws-neuron-driver-publish-source)

## Table of Contents

* [Usage](#usage)
* [Source](#source)
* [Contribute](#contribute)
* [License](#license)

## Usage

tbd

## Source

The primary location is:
[git.sr.ht/~wombelix/aws-neuron-driver-publish-source](https://git.sr.ht/~wombelix/aws-neuron-driver-publish-source)

Mirrors are available on
[Codeberg](https://codeberg.org/wombelix/aws-neuron-driver-publish-source),
[Gitlab](https://gitlab.com/wombelix/aws-neuron-driver-publish-source)
and
[GitHub](https://github.com/wombelix/aws-neuron-driver-publish-source).

## Contribute

Please don't hesitate to provide Feedback,
open an Issue or create a Pull / Merge Request.

Just pick the workflow or platform you prefer and are most comfortable with.

Feedback, bug reports or patches to my sr.ht list
[~wombelix/inbox@lists.sr.ht](https://lists.sr.ht/~wombelix/inbox) or via
[Email and Instant Messaging](https://dominik.wombacher.cc/pages/contact.html)
are also always welcome.

## License

Unless otherwise stated: `MIT`

All files contain license information either as
`header comment` or `corresponding .license` file.

[REUSE](https://reuse.software) from the [FSFE](https://fsfe.org/)
implemented to verify license and copyright compliance.
