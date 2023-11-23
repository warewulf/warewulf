=========
Debugging
=========

Whether developing a new feature or fixing a bug, using the automated
test suite together with a debugger is a potent combination. This
guide here can't substitute for full documentation on a given
debugger; but it might help you get started debugging Warewulf.

Validating the code with vet
============================

The Warewulf ``Makefile`` includes a ``vet`` target which runs ``go
vet`` on the full codebase.

.. code-block:: console

   $ make vet
   

Running the full test suite
===========================

The Warewulf ``Makefile`` includes a ``test`` target which runs the
full test suite.

.. code-block:: console

   $ make test


Using delve
===========

If you have a failing test but you're having trouble tracking down
why, try using a debugger to step through the test. These instructions
use delve.


Installing delve
----------------

You can install delve as a regular user directly with Go.

.. code-block:: console

   $ go install github.com/go-delve/delve/cmd/dlv@latest

The ``dlv`` binary will be installed by default at
``$HOME/go/bin/dlv``. You can, of course, add ``$HOME/go/bin`` to your
path if you prefer.

.. code-block:: console

   $ PATH=$HOME/go/bin:$PATH


Running delve against a specific test
-------------------------------------

You can use delve to specifically run the test suite and, even more
specifically, a single failing test. In this example delve is
instructed to run the tests for Warewulf's ``node`` package, and
specifically the ``Test_GetAllNodeInfoDefaults`` test.

.. code-block:: console

   $ dlv test github.com/hpcng/warewulf/internal/pkg/node -- -test.v -test.run Test_GetAllNodeInfoDefaults
   Type 'help' for list of commands.
   (dlv) break node.Test_GetAllNodeInfoDefaults
   Breakpoint 1 set at 0x26c0d0 for github.com/hpcng/warewulf/internal/pkg/node.Test_GetAllNodeInfoDefaults() ./internal/pkg/node/nodeyaml_test.go:51

Setting a breakpoint at ``node.Test_GetAllNodeInfoDefaults`` pauses
execution once the test starts, and allows us to ``continue`` through
all the setup prior to that point.

.. code-block:: console

   (dlv) continue
   === RUN   Test_GetAllNodeInfoDefaults
   > github.com/hpcng/warewulf/internal/pkg/node.Test_GetAllNodeInfoDefaults() ./internal/pkg/node/nodeyaml_test.go:51 (hits goroutine(35):1 total:1) (PC: 0x26c0d0)
       46:		assert.Contains(t, nodeYaml.Nodes, "test_node")
       47:		assert.Equal(t, "A single node", nodeYaml.Nodes["test_node"].Comment)
       48:	}
       49:	
       50:	
   =>  51:	func Test_GetAllNodeInfoDefaults(t *testing.T) {
       52:		file, writeErr := writeTestConfigFile(`
       53:	nodes:
       54:	  test_node: {}`)
       55:		if file != nil {
       56:			defer os.Remove(file.Name())

Helpful commands from here include

``next``

  Execute the current line (marked by ``=>``) and proceed to the next
  line.

``step``

  Execute the current line (marked by ``=>``) and proceed to the next
  line, potentially moving into a function call.

``list``

  Display a contextual Go code listing, marking the next instruction.

``locals``

  Display all local variables in the current scope.

``print``

  Display (in detail) the value of a single variable from the current
  scope.

Read about other commands available within delve using the ``help``
command.


Example
-------

.. code-block:: console

   $ ~/go/bin/dlv test github.com/hpcng/warewulf/internal/pkg/node -- -test.v -test.run Test_GetAllNodeInfoDefaults
   Type 'help' for list of commands.

   (dlv) break node.Test_GetAllNodeInfoDefaults
   Breakpoint 1 set at 0x26c0d0 for github.com/hpcng/warewulf/internal/pkg/node.Test_GetAllNodeInfoDefaults() ./internal/pkg/node/nodeyaml_test.go:51

   (dlv) break nodeinfo.go:417
   Breakpoint 2 set at 0x267f18 for github.com/hpcng/warewulf/internal/pkg/node.NewNodeInfo() ./internal/pkg/node/nodeinfo.go:417

   (dlv) continue
   === RUN   Test_GetAllNodeInfoDefaults
   > github.com/hpcng/warewulf/internal/pkg/node.Test_GetAllNodeInfoDefaults() ./internal/pkg/node/nodeyaml_test.go:51 (hits goroutine(19):1 total:1) (PC: 0x26c0d0)
   46:		assert.Contains(t, nodeYaml.Nodes, "test_node")
   47:		assert.Equal(t, "A single node", nodeYaml.Nodes["test_node"].Comment)
   48:	}
   49:	
   50:	
   =>  51:	func Test_GetAllNodeInfoDefaults(t *testing.T) {
   52:		file, writeErr := writeTestConfigFile(`
   53:	nodes:
   54:	  test_node: {}`)
   55:		if file != nil {
   56:			defer os.Remove(file.Name())

   (dlv) continue
   WARN   : Error reading UNDEF/warewulf/defaults.conf: open UNDEF/warewulf/defaults.conf: no such file or directory
   > github.com/hpcng/warewulf/internal/pkg/node.NewNodeInfo() ./internal/pkg/node/nodeinfo.go:417 (hits goroutine(19):1 total:1) (PC: 0x267f18)
   412:			defaultNodeConf.NetDevs = nil
   413:			nodeInfo.SetDefFrom(defaultNodeConf)
   414:		}
   415:	
   416:		// Load normal attributes
   => 417:		if nodeConf != nil {
   418:			// If no profiles are included, automatically include the
   419:			// default profile.
   420:			if len(nodeConf.Profiles) == 0 {
   421:				nodeInfo.Profiles.SetSlice([]string{"default"})
   422:			} else {

   (dlv) next
   > github.com/hpcng/warewulf/internal/pkg/node.NewNodeInfo() ./internal/pkg/node/nodeinfo.go:420 (PC: 0x267f24)
   415:	
   416:		// Load normal attributes
   417:		if nodeConf != nil {
   418:			// If no profiles are included, automatically include the
   419:			// default profile.
   => 420:			if len(nodeConf.Profiles) == 0 {
   421:				nodeInfo.Profiles.SetSlice([]string{"default"})
   422:			} else {
   423:				nodeInfo.Profiles.SetSlice(nodeConf.Profiles)
   424:			}
   425:	

   (dlv) next
   > github.com/hpcng/warewulf/internal/pkg/node.NewNodeInfo() ./internal/pkg/node/nodeinfo.go:421 (PC: 0x267f3c)
   416:		// Load normal attributes
   417:		if nodeConf != nil {
   418:			// If no profiles are included, automatically include the
   419:			// default profile.
   420:			if len(nodeConf.Profiles) == 0 {
   => 421:				nodeInfo.Profiles.SetSlice([]string{"default"})
   422:			} else {
   423:				nodeInfo.Profiles.SetSlice(nodeConf.Profiles)
   424:			}
   425:	
   426:			nodeInfo.SetFrom(nodeConf)

   (dlv) next
   > github.com/hpcng/warewulf/internal/pkg/node.NewNodeInfo() ./internal/pkg/node/nodeinfo.go:426 (PC: 0x267fec)
   421:				nodeInfo.Profiles.SetSlice([]string{"default"})
   422:			} else {
   423:				nodeInfo.Profiles.SetSlice(nodeConf.Profiles)
   424:			}
   425:	
   => 426:			nodeInfo.SetFrom(nodeConf)
   427:		}
   428:	
   429:		// Load default attributes for each NetDev
   430:		if defaultNetDevConf != nil {
   431:			for _, netdev := range nodeInfo.NetDevs {

   (dlv) next
   > github.com/hpcng/warewulf/internal/pkg/node.NewNodeInfo() ./internal/pkg/node/nodeinfo.go:430 (PC: 0x268000)
   425:	
   426:			nodeInfo.SetFrom(nodeConf)
   427:		}
   428:	
   429:		// Load default attributes for each NetDev
   => 430:		if defaultNetDevConf != nil {
   431:			for _, netdev := range nodeInfo.NetDevs {
   432:				netdev.SetDefFrom(defaultNetDevConf)
   433:			}
   434:		}
   435:	

   (dlv) print nodeInfo
   github.com/hpcng/warewulf/internal/pkg/node.NodeInfo {
   Id: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 1, cap: 1, [
   "test_node",
   ],
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 0, cap: 0, nil,},
   Comment: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 0, cap: 0, nil,},
   ClusterName: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 0, cap: 0, nil,},
   ContainerName: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 0, cap: 0, nil,},
   Ipxe: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 1, cap: 1, ["default"],},
   RuntimeOverlay: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 1, cap: 1, ["generic"],},
   SystemOverlay: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 1, cap: 1, ["wwinit"],},
   Root: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 1, cap: 1, [
   "initramfs",
   ],},
   Discoverable: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 0, cap: 0, nil,},
   Init: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 1, cap: 1, [
   "/sbin/init",
   ],},
   AssetKey: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 0, cap: 0, nil,},
   Kernel: *github.com/hpcng/warewulf/internal/pkg/node.KernelEntry {
   Override: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x4000158370),
   Args: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001583c8),},
   Ipmi: *github.com/hpcng/warewulf/internal/pkg/node.IpmiEntry {
   Ipaddr: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001b6600),
   Netmask: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001b6658),
   Port: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001b66b0),
   Gateway: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001b6708),
   UserName: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001b6760),
   Password: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001b67b8),
   Interface: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001b6810),
   Write: (*"github.com/hpcng/warewulf/internal/pkg/node.Entry")(0x40001b6868),
   Tags: map[string]*github.com/hpcng/warewulf/internal/pkg/node.Entry [],},
   Profiles: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 1, cap: 1, ["default"],
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 1, cap: 1, ["default"],},
   PrimaryNetDev: github.com/hpcng/warewulf/internal/pkg/node.Entry {
   value: []string len: 0, cap: 0, nil,
   altvalue: []string len: 0, cap: 0, nil,
   from: "",
   def: []string len: 0, cap: 0, nil,},
   NetDevs: map[string]*github.com/hpcng/warewulf/internal/pkg/node.NetDevEntry [],
   Tags: map[string]*github.com/hpcng/warewulf/internal/pkg/node.Entry [],}
