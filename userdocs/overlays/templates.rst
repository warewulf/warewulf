.. _templates:

=========
Templates
=========

Templates (denoted in overlays with a ``.ww`` suffix) allow you to create
dynamic configuration specifically for the node that it is applied to. Templates
have access to all metadata from the node registry (``nodes.conf``) and much of
the server configuration (``warewulf.conf``), and can also reference and import
files from the server file system.

Warewulf uses the ``text/template`` engine to facilitate implementing dynamic
content. This template format is documented at `pkg.go.dev/text/template.
<https://pkg.go.dev/text/template>`_

.. note::

   When the template is rendered within a built overlay image, the ``.ww`` will
   be dropped, so ``/etc/hosts.ww`` will end up being ``/etc/hosts``.

Non-Overlay Templates
=====================

Most Warewulf templates are included in overlays, but there are a few
non-overlay templates as well.

* ``/etc/warewulf/ipxe/``: includes iPXE script templates to direct iPXE during
  the network boot process.
* ``/etc/warewulf/grub/``: includes GRUB script templates to direct GRUB during
  the network boot process.
* ``/usr/share/warewulf/bmc/``: includes templates to generate BMC control
  commands for the ``wwctl power``, ``wwctl sensor``, and ``wwctl console``
  commands.

Template functions
==================

Warewulf templates have access to a number of functions that assist in creating
more dynamic and expressive templates.

Default functions
-----------------

``text/template`` includes a number of `default functions
<https://pkg.go.dev/text/template#hdr-Functions>`_ that are available during
Warewulf template processing.

Sprig
-----

Supplementing the default functions, Warewulf templates also have access to
`Sprig functions.`_

.. _Sprig functions.: https://masterminds.github.io/sprig/

Include
-------

Reads content from the given file into the template. If the file does not begin
with ``/`` it is considered relative to ``Paths.Sysconfdir``.

.. code-block::

   {{ Include "/root/.ssh/authorized_keys" }}

IncludeFrom
-----------

Reads content from the given file from the given image into the template.

.. code-block::

   {{ IncludeFrom $.ImageName "/etc/passwd" }}

IncludeBlock
------------

Reads content from the given file into the template, stopping when the provided
abort string is found.

.. code-block::
  
   {{ IncludeBlock "/etc/hosts" "# Do not edit after this line" }}

.. _importLink:

ImportLink
----------

Causes the processed template file to become a symlink to the same target as the
referenced symlink.

.. code-block::

   {{ ImportLink "/etc/localtime" }}

basename
--------

Returns the base name of the given path.

.. code-block::

   {{- range $type, $name := $.Tftp.IpxeBinaries }}
    if option architecture-type = {{ $type }} {
        filename "/warewulf/{{ basename $name }}";
    }
   {{- end }}

file
----

Write the content from the template to the specified file name. May be specified
more than once in a template to write content to multiple files.

.. code-block::

   {{- range $devname, $netdev := .NetDevs }}
   {{- $filename := print "ifcfg-" $devname ".conf" }}
   {{ file $filename }}
   {{/* content here */}}
   {{- end }}

.. _softlink:

softlink
--------

Causes the processed template file to become a symlink to the referenced target.

.. code-block::
  
   {{ printf "%s/%s" "/usr/share/zoneinfo" .Tags.localtime | softlink }}

.. _readlink:

readlink
--------

Equivalent to ``filepath.EvalSymlinks``. Returns the target path of a named
symlink.

.. code-block::

   {{ readlink /etc/localtime }}

IgnitionJson
------------

Generates JSON suitable for use by Ignition to create 

abort
-----

Immediately aborts processing the template and does not write a file.

.. code-block::
  
   {{ abort }}

nobackup
--------

   Disables the creation of a backup file when replacing files with the current
   template.

.. code-block::

   {{ nobackup }}

.. _UniqueField:

UniqueField
-----------

UniqueField returns a filtered version of a multi-line input string. input is
expected to be a field-separated format with one record per line (terminated by
`\n`). Order of lines is preserved, with the first matching line taking
precedence.

For example, the following template snippet has been used in the ``syncuser`` overlay
to generate a combined ``/etc/passwd``.

.. code-block::

   {{
       printf "%s\n%s" 
           (IncludeFrom $.ImageName "/etc/passwd" | trim)
           (Include (printf "%s/%s" .Paths.Sysconfdir "passwd") | trim)
       | UniqueField ":" 0 | trim
   }}

Examples
========

Many example templates are included in the distribution overlays. The ``debug``
template also includes a ``tstruct.ww`` template that includes much of the
available metadata.

.. code-block:: shell

   wwctl overlay show debug tstruct.ww
   wwctl overlay show debug tstruct.ww --render=n1

Node-Specific Files
-------------------

Sometimes there is the need to have specific files for each cluster node which
can't be generated by a template (e.g., a per-node Kerberos keytab). You can
include these files with following template:

.. code-block::

   {{ Include (printf "/srv/%s/%s" .Id "payload") }}
