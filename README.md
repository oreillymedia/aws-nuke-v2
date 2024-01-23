# aws-nuke

[![license](https://img.shields.io/github/license/ekristen/aws-nuke.svg)](https://github.com/ekristen/aws-nuke/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/ekristen/aws-nuke.svg)](https://github.com/ekristen/aws-nuke/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/ekristen/aws-nuke)](https://goreportcard.com/report/github.com/ekristen/aws-nuke)
[![Maintainability](https://api.codeclimate.com/v1/badges/bf05fb12c69f1ea7f257/maintainability)](https://codeclimate.com/github/ekristen/aws-nuke/maintainability)

## Overview

Remove all resources from an AWS account.

*aws-nuke* is stable, but it is likely that not all AWS resources are covered by it. Be encouraged to add missing
resources and create a Pull Request or to create an [Issue](https://github.com/ekristen/aws-nuke/issues/new).

## Documentation

All documentation is in the [docs/](docs) directory and is built using [Material for Mkdocs](https://squidfunk.github.io/mkdocs-material/). 

It is hosted at [https://ekristen.github.io/aws-nuke/](https://ekristen.github.io/aws-nuke/).

## History of this Fork

**Important:** this is a full fork of the original tool written by the folks over at [rebuy-de](https://github.com/rebuy-de).
This fork became necessary after attempting to make contributions and respond to issues to learn that the current
maintainers only have time to work on the project about once a month and while receptive to bringing in other 
people to help maintain, made it clear it would take time. Considering the feedback cycle was already weeks on 
initial communications, I had to make the hard decision to fork and maintain it.

### libnuke

I also needed a version of this tool for Azure and GCP, and initially I just copied and altered the code I needed for
Azure, but I didn't want to have to maintain multiple copies of the same code, so I decided to create 
[libnuke](https://github.com/ekristen/libnuke) to abstract all the code that was common between the two tools and write
proper unit tests for it. 

## Attribution, License, and Copyright

The rewrite of this tool to use [libnuke](https://github.com/ekristen/libnuke) would not have been posssible without the
hard work that came before me on the original tool by the team and contributors over at [rebuy-de](https://github.com/rebuy-de)
and their original work on [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).

This tool is licensed under the MIT license. See the [LICENSE](LICENSE) file for more information. The bulk of this
tool was rewritten to use [libnuke](https://github.com/ekristen/libnuke) which was in part originally sourced from
[rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).

## Contribute

You can contribute to *aws-nuke* by forking this repository, making your changes and creating a Pull Request against
this repository. If you are unsure how to solve a problem or have other questions about a contributions, please create
a GitHub issue.

## Version 3

Version 3 is a rewrite of this tool using [libnuke](https://github.com/ekristen/libnuke) with a focus on improving a number of the outstanding things
that I couldn't get done with the original project without separating out the core code into a library. See Goals 
below for more.

### Changes

- The root command will result in help now on v3, the primary nuke command moved to `nuke`. **Breaking**
- CloudFormation Stacks now support a hold and wait for parent deletion process. **Quasi-Breaking**
- Nested CloudFormation Stacks are now eligible for deletion and no longer omitted. **Quasi-Breaking**
- The entire resource lister format has changed and requires a struct.
- Context is passed throughout the entire library now, including the listing function and the removal function.
  - This is in preparation for supporting AWS SDK Go v2

### Goals

- Adding additional tests
- Adding additional resources
- Adding documentation for adding resources and using the tool
- Consider adding DAG for dependencies between resource types and individual resources
  - This will improve the process of deleting resources that have dependencies on other resources and reduce 
    errors and unnecessary API calls.

## Documentation

The project is built to have the documentation right alongside the code in the `docs/` directory leveraging 
[Material for Mkdocs](https://squidfunk.github.io/mkdocs-material/)

In the root of the project exists mkdocs.yml which drives the configuration for the documentation.

This README.md is currently copied to `docs/index.md` and the documentation is automatically published to the GitHub
pages location for this repository using a GitHub Action workflow. It does not use the `gh-pages` branch.


## Use Cases

- We are testing our [Terraform](https://www.terraform.io/) code with Jenkins. Sometimes a Terraform run fails during development and
  messes up the account. With *aws-nuke* we can simply clean up the failed account, so it can be reused for the next
  build.
- Our platform developers have their own AWS Accounts where they can create their own Kubernetes clusters for testing
  purposes. With *aws-nuke* it is very easy to clean up these account at the end of the day and keep the costs low.


### Feature Flags

There are some features, which are quite opinionated. To make those work for
everyone, *aws-nuke* has flags to manually enable those features. These can be
configured on the root-level of the config, like this:

```yaml
---
feature-flags:
  disable-deletion-protection:
    RDSInstance: true
    EC2Instance: true
    CloudformationStack: true
  force-delete-lightsail-addons: true
```

### Filtering Resources

It is possible to filter this is important for not deleting the current user
for example or for resources like S3 Buckets which have a globally shared
namespace and might be hard to recreate. Currently the filtering is based on
the resource identifier. The identifier will be printed as the first step of
*aws-nuke* (eg `i-01b489457a60298dd` for an EC2 instance).

**Note: Even with filters you should not run aws-nuke on any AWS account, where
you cannot afford to lose all resources. It is easy to make mistakes in the
filter configuration. Also, since aws-nuke is in continous development, there
is always a possibility to introduce new bugs, no matter how careful we review
new code.**

The filters are part of the account-specific configuration and are grouped by
resource types. This is an example of a config that deletes all resources but
the `admin` user with its access permissions and two access keys:

```yaml
---
regions:
- global
- eu-west-1

account-blocklist:
- 1234567890

accounts:
  0987654321:
    filters:
      IAMUser:
      - "admin"
      IAMUserPolicyAttachment:
      - "admin -> AdministratorAccess"
      IAMUserAccessKey:
      - "admin -> AKSDAFRETERSDF"
      - "admin -> AFGDSGRTEWSFEY"
```

Any resource whose resource identifier exactly matches any of the filters in
the list will be skipped. These will be marked as "filtered by config" on the
*aws-nuke* run.

#### Filter Properties

Some resources support filtering via properties. When a resource support these
properties, they will be listed in the output like in this example:

```log
global - IAMUserPolicyAttachment - 'admin -> AdministratorAccess' - [RoleName: "admin", PolicyArn: "arn:aws:iam::aws:policy/AdministratorAccess", PolicyName: "AdministratorAccess"] - would remove
```

To use properties, it is required to specify a object with `properties` and
`value` instead of the plain string.

These types can be used to simplify the configuration. For example, it is
possible to protect all access keys of a single user:

```yaml
IAMUserAccessKey:
- property: UserName
  value: "admin"
```

#### Filter Types

There are also additional comparision types than an exact match:

- `exact` – The identifier must exactly match the given string. This is the default.
- `contains` – The identifier must contain the given string.
- `glob` – The identifier must match against the given [glob
  pattern](https://en.wikipedia.org/wiki/Glob_(programming)). This means the
  string might contains wildcards like `*` and `?`. Note that globbing is
  designed for file paths, so the wildcards do not match the directory
  separator (`/`). Details about the glob pattern can be found in the [library
  documentation](https://godoc.org/github.com/mb0/glob).
- `regex` – The identifier must match against the given regular expression.
  Details about the syntax can be found in the [library
  documentation](https://golang.org/pkg/regexp/syntax/).
- `dateOlderThan` - The identifier is parsed as a timestamp. After the offset is added
  to it (specified in the `value` field), the resulting timestamp must be AFTER the
  current time. Details on offset syntax can be found in the [library documentation](https://golang.org/pkg/time/#ParseDuration).
  Supported date formats are epoch time, `2006-01-02`, `2006/01/02`, `2006-01-02T15:04:05Z`,
  `2006-01-02T15:04:05.999999999Z07:00`, and `2006-01-02T15:04:05Z07:00`.

To use a non-default comparision type, it is required to specify an object with
`type` and `value` instead of the plain string.

These types can be used to simplify the configuration. For example, it is
possible to protect all access keys of a single user by using `glob`:

```yaml
IAMUserAccessKey:
- type: glob
  value: "admin -> *"
```

#### Using Them Together

It is also possible to use Filter Properties and Filter Types together. For
example to protect all Hosted Zone of a specific TLD:

```yaml
Route53HostedZone:
- property: Name
  type: glob
  value: "*.rebuy.cloud."
```

#### Inverting Filter Results

Any filter result can be inverted by using `invert: true`, for example:

```yaml
CloudFormationStack:
- property: Name
  value: "foo"
  invert: true
```

In this case *any* CloudFormationStack ***but*** the ones called "foo" will be
filtered. Be aware that *aws-nuke* internally takes every resource and applies
every filter on it. If a filter matches, it marks the node as filtered.

#### Filter Presets

It might be the case that some filters are the same across multiple accounts.
This especially could happen, if provisioning tools like Terraform are used or
if IAM resources follow the same pattern.

For this case *aws-nuke* supports presets of filters, that can applied on
multiple accounts. A configuration could look like this:

```yaml
---
regions:
- "global"
- "eu-west-1"

account-blocklist:
- 1234567890

accounts:
  555421337:
    presets:
    - "common"
  555133742:
    presets:
    - "common"
    - "terraform"
  555134237:
    presets:
    - "common"
    - "terraform"
    filters:
      EC2KeyPair:
      - "notebook"

presets:
  terraform:
    filters:
      S3Bucket:
      - type: glob
        value: "my-statebucket-*"
      DynamoDBTable:
      - "terraform-lock"
  common:
    filters:
      IAMRole:
      - "OrganizationAccountAccessRole"
```

## Install

### For macOS
`brew install aws-nuke`

### Use Released Binaries

The easiest way of installing it, is to download the latest
[release](https://github.com/ekristen/aws-nuke/releases) from GitHub.

#### Example for Linux Intel/AMD

Download and extract
`$ wget -c https://github.com/rebuy-de/aws-nuke/releases/download/v2.25.0/aws-nuke-v2.25.0-linux-amd64.tar.gz -O - | tar -xz -C $HOME/bin`

Run
`$ aws-nuke-v2.25.0-linux-amd64`

### Compile from Source

To compile *aws-nuke* from source you need a working
[Golang](https://golang.org/doc/install) development environment.

*aws-nuke* uses go modules and so the clone path should no matter.

The easiest way to compile is by using [goreleaser](https://goreleaser.io)

```bash
goreleaser --rm-dist --snapshot --single-target
```

**Note:** this will automatically build for your current architecture and place the result
in the releases directory.

You may also use `make` to compile the binary, this was left over from before the fork.

Also you need to install [golint](https://github.com/golang/lint/) and [GNU
Make](https://www.gnu.org/software/make/).

Then you just need to run `make build` to compile a binary into the project
directory or `make install` go install *aws-nuke* into `$GOPATH/bin`. With
`make xc` you can cross compile *aws-nuke* for other platforms.

### Docker

You can run *aws-nuke* with Docker by using a command like this:

```bash
$ docker run \
    --rm -it \
    -v /full-path/to/nuke-config.yml:/home/aws-nuke/config.yml \
    -v /home/user/.aws:/home/aws-nuke/.aws \
    quay.io/rebuy/aws-nuke:v2.25.0 \
    --profile default \
    --config /home/aws-nuke/config.yml
```

To make it work, you need to adjust the paths for the AWS config and the
*aws-nuke* config.

Also you need to specify the correct AWS profile. Instead of mounting the AWS
directory, you can use the `--access-key-id` and `--secret-access-key` flags.

Make sure you use the latest version in the image tag. Alternatiely you can use
`main` for the latest development version, but be aware that this is more
likely to break at any time.

## Testing

### Unit Tests

To unit test *aws-nuke*, some tests require [gomock](https://github.com/golang/mock) to run.
This will run via `go generate ./...`, but is automatically run via `make test`.
To run the unit tests:

```bash
make test
```

## Contact Channels

For now GitHub issues, may open a Slack or Discord if warranted.

## Contribute

You can contribute to *aws-nuke* by forking this repository, making your
changes and creating a Pull Request against our repository. If you are unsure
how to solve a problem or have other questions about a contributions, please
create a GitHub issue.
