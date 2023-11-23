================
Useful templates
================

The examples directory contains some useful examples for day to day cluster admninstration.

Genders
=======

The file `genders.ww` can be placed as `/etc/genders.ww` which will create a genders database which containts all nodes with their profile as key. Also for every user defined 'key' the 'tah' will be added. 

.. Note: 

A arbitrary tag with a key can be added to a node with

.. code-block:: bash
   wwctl node set --tagadd key=value n01
   wwctl node set --tagadd key2='foo=baar' n01

will result in genders file with following line


.. code-block:: bash
   n01: value,foo=baar

