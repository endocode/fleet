## fleet functional tests

This functional test suite deploys a fleet cluster using nspawn containers, and asserts fleet is functioning properly.

It shares an instance of etcd deployed on the host machine with each of the nspawn containers.

It's recommended to run this in a virtual machine environment on CoreOS (e.g. using [Vagrant][test-in-vagrant]).

Since the tests utilize [`systemd-nspawn`][systemd-nspawn], this needs to be invoked as sudo/root.

If the tests are aborted partway through, it's currently possible for them to leave residual state as a result of the `systemd-nspawn` operations. This can be cleaned up using the `clean.sh` script.

### run tests in vagrant

Vagrant will configure CoreOS instance with one-member etcd2 daemon.

```sh
$ git clone https://github.com/coreos/fleet
$ cd fleet/functional
$ # Vagrantfile based on official CoreOS Vagranfile with one extra provision string
$ ./run-in-vagrant
```

### run tests inside other CoreOS platforms (QEMU/BareMetal/libvirt/etc)

```sh
$ git clone https://github.com/coreos/fleet
```

If you didn't configure etcd2 daemon yet, just run this script:

```sh
$ sudo fleet/functional/start_etcd
```

It will configure and start one-member etcd2 daemon.

Then run the functional tests:

```sh
$ git clone https://github.com/coreos/fleet
$ sudo fleet/functional/test
```

## configure environment to run Vagrant

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

**NOTE**: NFS and Vagrant doesn't work out of the box on CentOS 6.x

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

[test-in-vagrant]: #run-tests-in-vagrant
[systemd-nspawn]: https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html
