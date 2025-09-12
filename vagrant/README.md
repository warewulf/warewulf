## Dev environment

For dev and testing purpose, a vagrant environment is provided.

For now, the only vagrant provider supported is libvirt and only usable on Linux machines.

### Initialize environment

To setup the initial environment:

```shell
vagrant init-warewulf
```

This command creates three VM nodes:
* wwctl
* wwnode1
* wwnode2

On `wwctl` node, the latest stable warewulf release available will be installed and
automatically adds `wwnode1`/`wwnode2` respectively as `n1` and `n2` in warewulf.

Once initialized, you can use `vagrant up` and `vagrant halt` to start/stop VM's as usual.

### Cleanup environment

To cleanup the environment:

```shell
vagrant cleanup-warewulf
```

This command stops VM nodes and deletes them

### Power control

`wwnode1` and `wwnode2` can be started/stopped from `wwctl` node via commands:

```shell
wwctl power on n1
wwctl power on n2
```

```shell
wwctl power off n1
wwctl power off n2
```

### Alpine boot test

For a quick test/demo, an alpine boot script is provided, you can run it with:

```shell
vagrant ssh wwctl -c "sudo ./alpine-boot.sh"
```