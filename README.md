![GitHub Release](https://img.shields.io/github/v/release/ttl256/euivator?link=https%3A%2F%2Fgithub.com%2Fttl256%2Feuivator%2Freleases)
[![Audit](https://github.com/ttl256/euivator/actions/workflows/audit.yaml/badge.svg)](https://github.com/ttl256/euivator/actions/workflows/audit.yaml)
[![Release](https://github.com/ttl256/euivator/actions/workflows/release.yaml/badge.svg)](https://github.com/ttl256/euivator/actions/workflows/release.yaml)

# euivator

A CLI tool to work with EUIs

- Work with EUIs
  - Verify whether a given string is a valid EUI
  - Convert an EUI to a specified format: colon, dash, dot, plain
  - Produce an EUI-64 modified from an EUI-48
  - Supply an IPv6 prefix and EUI-48/EUI-64 to produce an IPv6 address
- Work with OUIs
  - Provide an EUI or just a hex prefix to look it up in IEEE registries to
    determine which company owns a particular OUI allocation

## Design features

- Regular CLI tool behavior: accept arguments from the standard input or as
  positional arguments and output results to the standard output in a form that is
  digestible by other CLI tools (like `awk`, `jq`)
- Offline OUI lookup. euivator uses [public
  registries](https://standards.ieee.org/products-programs/regauth/) to collect
  text files with OUI allocations. The files are stored locally
- Although the files in the repositories are not static, they change quite
  rarely. But when they do change it's necessary to download ~5MB of data. The
  update process is manual: `euivator oui update`. To minimize impact on the
  network euivator implements client-side caching using ETags. (there are
  [issues](#ieee-etags-and-two-smoking-nodes) with caching on the IEEE side)
- Local cache is stored according to XDG specification powered by [adrg's
  library](https://github.com/adrg/xdg) but the target directory is configurable
- The result of an OUI lookup returns a list of allocations since IEEE allows
  duplicate allocations and euivator does not play god by trying to decide which
  one is correct
- The OUI lookup performed in the longest prefix match manner

## Install

### Binary release

The following command installs the latest version to `~/.local/bin`

```sh
curl -sSfL https://raw.githubusercontent.com/ttl256/euivator/HEAD/install.sh | sh -s -- -b ~/.local/bin
```

```
Usage: install.sh [-b <bindir>] [-d] [<tag>]
  -b sets bindir or installation directory, Defaults to ./bin
  -d turns on debug logging
   <tag> is a tag from
   https://github.com/ttl256/euivator/releases
   If tag is missing, then the latest will be used.
```

### MacOS (brew)

```sh
brew tap ttl256/euivator && \
brew install euivator
```

### Linux

#### APT/DEB

```sh
echo "deb [trusted=yes] https://apt.fury.io/ttl256/ /" > /etc/apt/sources.list.d/ttl256.fury.list && \
apt update && \
apt install euivator
```

#### YUM/RPM

```sh
cat <<EOF >> /etc/yum.repos.d/ttl256.fury.repo
[fury-ttl256]
name=ttl256
baseurl=https://yum.fury.io/ttl256/
enabled=1
gpgcheck=0
EOF

yum update && \
yum install euivator
```

### Build from source

```sh
make build
```

## After installation

OUI registries are not downloaded as part of the post-install to not mess users
who just want to convert EUIs without OUI lookups.

Download OUI registries

```sh
euivator oui update
```

## Examples

```sh
# Convert EUIs
$ euivator eui convert --format dot DE:AD:BE:EF:11:22
dead.beef.1122
$ echo "DEADBEEF1122\nDE:AD:BE:EF:11:22:33:44" | euivator eui convert --format dash
de-ad-be-ef-11-22
de-ad-be-ef-11-22-33-44
# Generate EUI-64 modified
$ euivator eui modified DEADBEEF1122
dc:ad:be:ff:fe:ef:11:22
# Generate IPv6 address from a prefix and an EUI
$ euivator eui addr6 2001:db8:dead:beef::/64 00:00:00:00:00:00
2001:db8:dead:beef:200:ff:fe00:0
# Lookup OUI allocation by EUI
$ euivator oui lookup 28:6f:b9:11:22:33 | jq
{
  "input": "286FB9112233",
  "input_raw": "28:6f:b9:11:22:33",
  "records": [
    {
      "assignment": "286FB9",
      "registry": "MA-L",
      "org_name": "Nokia Shanghai Bell Co., Ltd.",
      "org_address": "No.388 Ning Qiao Road,Jin Qiao Pudong Shanghai Shanghai   CN 201206"
    }
  ]
}
# Lookup OUIs for interfaces on your machine
$ ifconfig | awk '/ether/ {print $2}' | euivator oui lookup | jq -c 'select(.records | length > 0)'
```

## IEEE, ETags and two smoking nodes

Posted the
[issue](https://www.reddit.com/r/IEEE/comments/1i82j5i/mac_address_registry_serves_csv_files_in_a_way/)
to r/IEEE. No luck so far.

TLDR; IEEE uses default nginx config to generate ETags. The algorithm takes
last-modified and content-length headers as an input to generate ETag. Multiple
nodes serve the files and last-modified differs for each of them by a few
seconds.

## Configuration

Refer to `configs/euivator.yaml.sample` to check available options. Same options
are available via environment variables in the form of `EUIVATOR_*`.

## Similar software

- [oui](https://github.com/thatmattlove/oui)
  - Does not accept arguments from the standard input
  - Output is pretty-table-formatted which makes it hard to pipeline with other
    CLI tools
  - Does not cache OUI registries
