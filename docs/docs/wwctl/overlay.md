---
id: overlay
title: wwctl overlay
---

Management interface for Warewulf overlays

build
~~~~~
This command will build a system or runtime overlay.

-s, --system  Show System Overlays as well
-a, --all  Build all overlays (runtime and system)

chmod
~~~~~
This command will allow you to change the permissions of a file within an overlay.

-s, --system  Show System Overlays as well
-n, --noupdate  Don't update overlays

create
~~~~~~
This command will create a new empty overlay.

-s, --system  Show System Overlays as well
-n, --noupdate  Don't update overlays

delete
~~~~~~
This command will delete files within an overlay or an entire overlay if no files are given to remove (use with caution).

-s, --system  Show system overlays instead of runtime
-f, --force  Force deletion of a non-empty overlay
-p, --parents  Remove empty parent directories
-n, --noupdate  Don't update overlays

edit
~~~~
This command will allow you to edit or create a new file within a given overlay. Note: when creating files ending in a '.ww' suffix this will always be parsed as a Warewulf template file, and the suffix will be removed automatically

-s, --system  Show system overlays instead of runtime
-f, --files  List files contained within a given overlay
-p, --parents  Create any necessary parent directories
-m, --mode  Permission mode for directory
-n, --noupdate  Don't update overlays

import
~~~~~~
This command will import a file into a given Warewulf overlay.

-s, --system  Show system overlays instead of runtime
-m, --mode  Permission mode for directory
-n, --noupdate  Don't update overlays

list
~~~~
This command will show you information about Warewulf overlays and the files contained within.

-s, --system  Show system overlays instead of runtime
-a, --all  List the contents of overlays
-l, --long  List 'long' of all overlay contents

mkdir
~~~~~
This command will allow you to create a new file within a given Warewulf overlay.

-s, --system  Show System Overlays as well
-m, --mode  Permission mode for directory
-n, --noupdate  Don't update overlays

show
~~~~
This command will output the contents of a file within a given

-s, --system  Show System Overlays as well
