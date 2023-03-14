# Spammer

Spammer takes a list of hosts and a list of paths, then it requests every combination.

**Note:** spammer shows only the requests that responded with a status code less than 400.

## Example

```
$ spammer -w tests/paths.txt -f tests/hosts.txt   

https://vitorfalcao.com/.git/index [308]
https://google.com/robots.txt [301]
https://twitter.com/robots.txt [200]
https://twitter.com/.git [200]
https://twitter.com/.git/index [200]
https://twitter.com/.git/core [200]
```

## Usage

```
Usage:
  spammer [flags]

Flags:
  -d, --debug              Enable debug mode
  -h, --help               help for spammer
  -f, --hostsfile string   File containing hosts (default "hosts.txt")
  -r, --rate-limit uint    Number of requests per second (default 100)
  -s, --silent             Disable error messages
  -t, --threads uint       Number of threads (default 10)
  -w, --wordlist string    Wordlist containing paths (default "wordlist.txt")
```

## Installation

### Using go install

```
go install https://github.com/vitorfhc/spammer
spammer -h
```

### From source

```
git clone https://github.com/vitorfhc/spammer
cd spammer
go install
```
