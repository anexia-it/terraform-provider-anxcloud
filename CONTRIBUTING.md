# How to Contribute

The project is under [Apache 2.0 licensed](LICENSE) and accept contributions via
GitHub pull requests.  This document outlines some of the conventions on
development workflow, commit message formatting, contact points and other
resources to make it easier to get your contribution accepted.

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. See the [DCO](DCO) file for details.

Any copyright notices in this repo should specify the authors as "the Anexia Terraform provider contributors".

To sign your work, just add a line like this at the end of your commit message:

```
Signed-off-by: Joe Example <joe@example.com>
```

This can easily be done with the `--signoff` option to `git commit`.

By doing this you state that you can certify the following (from https://developercertificate.org/):

## Email and discussions

The terraform-provider-anxcloud project currently uses the general email list and forum topics:

- Email: [opensource@anexia-it.com](mailto:opensource@anexia-it.com)

Please avoid emailing maintainers found in the MAINTAINERS file directly. They
are very busy and read the mailing lists.

## Reporting a security vulnerability

Due to their public nature, GitHub and mailing lists are not appropriate places for reporting vulnerabilities. If you suspect you have found a security vulnerability in rkt, please do not file a GitHub issue, but instead email opensource@anexia-it.com with the full details, including steps to reproduce the issue.

## Getting Started

- Fork the repository on GitHub
- Enable the [`pre-commit` hook](#pre-commit-hook)
- Read the [README](README.md) and [Development](README.md#development) for build and test instructions
- Play with the project, submit bugs, submit patches!

### pre-commit hook

We use [`pre-commit`](https://pre-commit.com/) to run some checks on `git commit`, making sure
the code is clean before added to the local history. It's probably available in your systems package manager.
See the link for install instructions. When installed, run `pre-commit install` in the cloned project root directory
to activate it for your local repository (or `pre-commit install -f` to replace any previous pre-commit configs).

### Contribution Flow

This is a rough outline of what a contributor's workflow looks like:

1. Make your code changes on the branch of your fork of [terraform-provider-anxcloud](https://github.com/anexia-it/terraform-provider-anxcloud)
2. Create a pull request
3. Add your changes to the **Unreleased** section in [CHANGELOG.md](CHANGELOG.md)

Additional steps carried out by maintainers:

4. Review pull request and request changes if necessary
5. Approve workflow run via GitHub
6. Merge the PR into the main branch of [terraform-provider-anxcloud](https://github.com/anexia-it/terraform-provider-anxcloud)

Thanks for your contributions!
