# Notes on creating the Virtual Box environment for Guardian integration testing

- Get [Virtual Box 4.3.10](http://download.virtualbox.org/virtualbox/4.3.10/VirtualBox-4.3.10-93012-OSX.dmg) from Oracle

- Install and open vBox manager

- Create a vm (named Guardian Dev Test)
    - 2Gb RAM
    - 8Gb HDD
    - 4 processors
    - username (`spowell`)
    - password (********)
    - o/w defaults (ensure network connection -- NAT default)

- Get Ubuntu (or other distro)
    - `ubuntu-14.04-desktop-amd64.iso` (got from Glyn)

- Start vm and install from DVD
    - accept updates during boot/install
    - reboot (no need to remove DVD)

- Install VBox Guest Additions
    - Devices menu in VBox (after booting VM) attaches image
    - Accept in dialog
    - Reboot (and get largest resolution for display -- can maximise afterwards, so don't worry)

- Get git and mercurial (needed by go get)
    - `sudo apt-get install git mercurial`

- Get go
    - (Firefox download from golang -- Ubuntu 64bit version -- version 1.2.2 current)
    - After installation rename to `~/golang`
    - Edit `.profile` to include:

            export GOROOT=$HOME/golang
            export GOPATH=$HOME/go
            export PATH=$GOPATH/bin:$PATH:$GOROOT/bin

    - Check we have the right version:
        - Reboot (`.profile` not rexecuted on restart Terminal? Then modify run terminal `as login shell` in Terminal preferences.)
        - `go version` in Terminal/Shell to check
        - `env|grep GO` to check setup

- Now get the Guardian code

        go get github.com/cf-guardian/guardian

- Now build/test something

        cd ~/go/src/github.com/cf-guardian/guardian/kernel/rootfs
        go get -t ./...
        go test

    We should get gomock and glog as a byproduct of this.

### Alternative to getting the Guardian code

Share a folder with the host copy of the Guardian code.

- Create shared folder on vbox
    - point to go project source root (`GOPATH`) on host
    - choose name for it (e.g. `gohost`)
    - is mounted in `/media/sf-gohost`
    - *put local user in `vboxsf` group* or else the folder cannot be accessed (!)
- The go compiler needs to be installed on the vbox, still, and then updates to the code on
the host are immediately visible to the vbox machine.


# INTELLIJ Go support on Mac OS X

Install Community Edition of INTELLIJ, and install the `Golang.org` plugin.

Then set up `GOROOT` and `GOPATH` in the `/etc/launchd.conf` file (which may need to be created by `root`) so as to get
INTELLIJ to see these vars when started.  These should be set to the same as they are in the host
system.  The Mac can be rebooted or issue

    egrep -v '^\s*#' /etc/launchd.conf | launchctl

to make the new settings "take" straight away. Then restart INTELLIJ.

Create a *new* Go project for INTELLIJ, do not create a new `src` folder, but point to the `src` folder for guardian:

    ~/dev/go/github.com/cf-guardian/guardian

(for me).

There is more to do: the source files for the Go project need to be set:

- go to the module/project settings, and navigate to the SDK settings
- clear the stuff in the ClassPath section -- this is not needed
- make sure the `$GOROOT/src/pkg` (in full, not using the env var, I don't think) is present

