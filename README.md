=======
gopistrano
==========

Automatic Deployment Tool in Golang

## Requirements

* GoLang >= 1.3.3

## Installation

Run this if you want to install gopistrano binary

``` go
go install github.com/alanchavez88/gopistrano
```

That will compile and install gopistrano in your $GOPATH

To deploy a project, you need to create a Gopfile. A Gopfile is just a configuration file in plain-text that will contain the credentials to SSH into your server, and the path of the directory where you want to deploy your project to.

This is a sample Gopfile
```
username = yourusername
password = yourpassword
# private_key = /home/user/.ssh/id_rsa
hostname = example.com
port = 22
repository = https://github.com/alanchavez88/theHarvester.git
keep_releases = 5
path = /home7/alanchav/gopistrano
use_sudo = false
webserver_user = nobody

```
The file above will clone the git repository above into the path specified in the Gopfile.

Currently gopistrano only supports git, other version controls will be added in the future.

It also only supports username and password authentication, the next update will provide authenticate via PEM files and SSH Keys

To deploy you have to run

``` sh
gopistrano deploy:setup

```

and then:

``` sh
gopistrano deploy
```

## Support

Need help with Gopistrano? Shoot me an email at alan@alanchavez.com
