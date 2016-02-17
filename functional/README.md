## configure environment

### Debian/Ubuntu

#### Install Vagrant

```sh
sudo apt-get install -y git nfs-kernel-server
wget https://releases.hashicorp.com/vagrant/1.8.1/vagrant_1.8.1_x86_64.deb
sudo dpkg -i vagrant_1.8.1_x86_64.deb
```

#### Install VirtualBox

```sh
echo "deb http://download.virtualbox.org/virtualbox/debian $(lsb_release -sc) contrib" | sudo tee /etc/apt/sources.list.d/virtualbox.list
wget -q https://www.virtualbox.org/download/oracle_vbox.asc -O- | sudo apt-key add -
sudo apt-get update
sudo apt-get install -y build-essential dkms
sudo apt-get install -y VirtualBox-5.0
#Previous VirtualBox (if you have problems with nested virtualization, more info here: https://www.virtualbox.org/ticket/14965)
#sudo apt-get install -y VirtualBox-4.3
```

### Centos/Fedora

**NOTE**: NFS doesn't work on CentOS 6.x

#### Install Vagrant

```sh
sudo yum install -y git nfs-utils
sudo reboot
sudo service nfs start
sudo yum install -y https://releases.hashicorp.com/vagrant/1.8.1/vagrant_1.8.1_x86_64.rpm
```

#### Install VirtualBox

```sh
source /etc/os-release
for id in $ID_LIKE $ID; do break; done
OS_ID=${id:-rhel}
curl http://download.virtualbox.org/virtualbox/rpm/$OS_ID/virtualbox.repo | sudo tee /etc/yum.repos.d/virtualbox.repo
sudo yum install -y make automake gcc gcc-c++ kernel-devel-`uname -r` dkms
sudo yum install -y VirtualBox-5.0
#Previous VirtualBox (if you have problems with nested virtualization, more info here: https://www.virtualbox.org/ticket/14965)
#sudo yum install -y VirtualBox-4.3
```

## run tests in vagrant

```sh
git clone https://github.com/coreos/fleet
cd fleet/functional
# wget https://raw.githubusercontent.com/coreos/coreos-vagrant/master/Vagrantfile
# we will use default unmodified CoreOS Vagrantfile
./run-in-vagrant
```

## fleet functional tests

This functional test suite deploys a fleet cluster using nspawn containers, and asserts fleet is functioning properly.

It shares an instance of etcd deployed on the host machine with each of the nspawn containers.

It's recommended to run this in a virtual machine environment on CoreOS (e.g. using coreos-vagrant). The only dependency for the tests not provided on the CoreOS image is `go`.

The caller must do three things before running the tests:

1. Ensure an ssh-agent is running and the functional-testing identity is loaded. The `SSH_AUTH_SOCK` environment variable must be set.

```
$ ssh-agent
$ ssh-add fleet/functional/fixtures/id_rsa
$ echo $SSH_AUTH_SOCK
/tmp/ssh-kwmtTOsL7978/agent.7978
```
2. Ensure the `FLEETD_BIN` and `FLEETCTL_BIN` environment variables point to the respective fleetd and fleetctl binaries that should be used to drive the actual tests.

```
$ export FLEETD_BIN=/path/to/fleetd
$ export FLEETCTL_BIN=/path/to/fleetctl
```

3. Make sure etcd is running on the host system.

```
$ systemctl start etcd
```

Then the tests can be run with:

```
# go test github.com/coreos/fleet/functional
```

Since the tests utilize `systemd-nspawn`, this needs to be invoked as sudo/root.

An example test session using coreos-vagrant follows. This assumes that go is available in `/home/core/go` and the fleet repository in `/home/core/fleet` on the target machine (the easiest way to achieve this is to use shared folders).
```
vagrant ssh core-01 -- -A
export GOROOT="$(pwd)/go"
export PATH="${GOROOT}/bin:$PATH"
cd fleet
ssh-add functional/fixtures/id_rsa
export GOPATH="$(pwd)/gopath"
export FLEETD_BIN="$(pwd)/bin/fleetd"
export FLEETCTL_BIN="$(pwd)/bin/fleetctl"
sudo -E env PATH=$PATH go test github.com/coreos/fleet/functional -v
```

If the tests are aborted partway through, it's currently possible for them to leave residual state as a result of the systemd-nspawn operations. This can be cleaned up using the `clean.sh` script.
