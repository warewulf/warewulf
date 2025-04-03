========
REST API
========

On-line documentation for the API is available at ``/api/docs``.

Authentication
==============

Authentication is managed at ``/etc/warewulf/auth.conf``. This is a
YAML formatted file with a single key: ``users:``, that is a list of user names
and passwords able to authenticate to the API.

.. warning::

   Because ``warewulfd`` runs as ``root`` by default, and because ``warewulfd``
   can run effectively arbitrary code via overlay templates, API access is
   tantamount to ``root`` access on the Warewulf server. For this reason, the
   API is only accessible via localhost by default. Still, handle API
   credentials with care.

.. code-block:: yaml

   users:
     - name: admin
       password hash: $2b$05$5QVWDpiWE7L4SDL9CYdi3O/l6HnbNOLoXgY2sa1bQQ7aSBKdSqvsC

Passwords are stored as bcrypt2 hashses, which can be generated with ``mkpasswd``.

.. code-block:: console

   $ mkpasswd --method=bcrypt
   Password: # admin
   $2b$05$5QVWDpiWE7L4SDL9CYdi3O/l6HnbNOLoXgY2sa1bQQ7aSBKdSqvsC

Node
====

* ``GET /api/nodes/``: Get nodes
* ``POST /api/nodes/overlays/build``: Build all overlays
* ``DELETE /api/nodes/{id}``: Delete an existing node
* ``GET /api/nodes/{id}``: Get a node
* ``PATCH /api/nodes/{id}``: Update an existing node
* ``PUT /api/nodes/{id}``: Add a node
* ``GET /api/nodes/{id}/fields``: Get node fields
* ``POST /api/nodes/{id}/overlays/build``: Build overlays for a node
* ``GET /api/nodes/{id}/raw``: Get a raw node

Profile
=======

* ``GET /api/profiles/``: Get node profiles
* ``DELETE /api/profiles/{id}``: Delete an existing profile
* ``GET /api/profiles/{id}``: Get a node profile
* ``PATCH /api/profiles/{id}``: Update an existing profile
* ``PUT /api/profiles/{id}``: Add a profile

Image
=====

* ``GET /api/images``: Get all images
* ``DELETE /api/images/{name}``: Delete an image
* ``GET /api/images/{name}``: Get an image
* ``PATCH /api/images/{name}``: Update or rename an image
* ``POST /api/images/{name}/build``: Build an image
* ``POST /api/images/{name}/import``: Import an image

Overlay
=======

* ``GET /api/overlays/``: Get overlays
* ``DELETE /api/overlays/{name}``: Delete an overlay
* ``GET /api/overlays/{name}``: Get an overlay
* ``PUT /api/overlays/{name}``: Create an overlay
* ``GET /api/overlays/{name}/file``: Get an overlay file
