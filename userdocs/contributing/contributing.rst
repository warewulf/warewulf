============
Contributing
============

Warewulf is an open source project, meaning we have the challenge of
limited resources. We are grateful for any support that you can
offer. Helping other users, raising issues, helping write
documentation, or contributing code are all ways to help!

Join the community
==================

This is a huge endeavor, and your help would be greatly appreciated!
Post to online communities about Warewulf, and request that your
distribution vendor, service provider, and system administrators
include Warewulf for you!

Warewulf on Slack
-----------------

Many members of the Warewulf community, including developers, communicate via Slack.
It's a great place to get help with an issue or talk about how you're using Warewulf.

An invite link is available at `https://warewulf.org/help/ <https://warewulf.org/help/>`.

Raise an Issue
==============

For general bugs/issues, you can open an issue `at the GitHub repo
<https://github.com/warewulf/warewulf/issues/new>`_.

Contribute to the code
======================

We use the traditional `GitHub Flow
<https://guides.github.com/introduction/flow>`_ to develop. This means
that you fork the main repo, create a new branch to make changes, and
submit a pull request (PR) to the master branch.

Check out our official `CONTRIBUTING.md
<https://github.com/warewulf/warewulf/blob/master/CONTRIBUTING.md>`_
document, which also includes a `code of conduct
<https://github.com/warewulf/warewulf/blob/master/CONTRIBUTING.md#code-of-conduct>`_.


Step 1. Fork the repo
---------------------

To contribute to Warewulf, you should obtain a GitHub account and fork
the `Warewulf <https://github.com/warewulf/warewulf>`_ repository. Once
forked, clone your fork of the repo to your computer. (Obviously, you
should replace ``your-username`` with your GitHub username.)

.. code-block:: bash

   git clone https://github.com/your-username/warewulf.git
   cd warewulf

Step 2. Checkout a new branch
-----------------------------

`Branches <https://guides.github.com/introduction/flow>`_ are a way
of isolating your features from the main branch. Given that we’ve just
cloned the repo, we will probably want to make a new branch from
master in which to work on our new feature. Lets call that branch
``new-feature``:

.. code-block:: bash

   git checkout master
   git checkout -b new-feature

.. note::

   You can always check which branch you are in by running ``git
   branch``.

Step 3. Make your changes
-------------------------

On your new branch, go nuts! Make changes, test them, and when you are
happy commit the changes to the branch:

.. code-block:: bash

   git add file-changed1 file-changed2...
   git commit -m "what changed?"

This commit message is important - it should describe exactly the
changes that you have made. Good commit messages read like so:

.. code-block:: bash

   git commit -m "changed function getConfig in functions.go to output csv to fix #2"
   git commit -m "updated docs about shell to close #10"

The tags ``close #10`` and ``fix #2`` are referencing issues that are
posted on the upstream repo where you will direct your pull
request. When your PR is merged into the master branch, these messages
will automatically close the issues, and further, they will link your
commits directly to the issues they intend to fix. This will help
future maintainers understand your contribution, or (hopefully not)
revert the code back to a previous version if necessary.

Step 4. Push your branch to your fork
-------------------------------------

When you are done with your commits, you should push your branch to
your fork (and you can also continuously push commits here as you
work):

.. code-block:: bash

   git push origin new-feature

Note that you should always check the status of your branches to see
what has been pushed (or not):

.. code-block:: bash

   git status

Step 5. Submit a Pull Request
-----------------------------

Once you have pushed your branch, then you can go to your fork (in the
web GUI on GitHub) and `submit a Pull Request
<https://help.github.com/articles/creating-a-pull-request>`_. Regardless
of the name of your branch, your PR should be submitted to the
``main`` branch. Submitting your PR will open a conversation thread
for the maintainers of Warewulf to discuss your contribution. At this
time, the continuous integration that is linked with the code base
will also be executed. If there is an issue, or if the maintainers
suggest changes, you can continue to push commits to your branch and
they will update the Pull Request.

Step 6. Keep your branch in sync
--------------------------------

Cloning the repo will create an exact copy of the Warewulf repository
at that moment. As you work, your branch may become out of date as
others merge changesinto the upstream master. In the event that you
need to update a branch, you will need to follow the next steps:

.. code-block:: bash

   # add a new remote named "upstream"
   git remote add upstream https://github.com/warewulf/warewulf.git
   # or another branch to be updated
   git checkout master
   git pull upstream master
   # to update your fork
   git push origin master
   git checkout new-feature
   git merge master
