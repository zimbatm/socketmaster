socketmaster: Zero-downtime restarts for your apps
==================================================

socketmaster takes care of starting your app while sharing the
file-descriptor.


* Q:: How is it different than the Nginx dance ?
* A:: The parent PID never changes so it works better with tools like
Upstart

* Q:: What happens if I need to change the command to be executed ?
* A:: Use a wrapper script:

```bash
#!/bin/sh
exec /path/to/your/program
```

Usage
=====

```
socketmaster -listen=tcp://:8080 -command=path/to/wrapper/script
```

listen supports the following socket types: tcp, tcp4, tcp6, unix, fd

Design
======

 * socketmaster opens the port
 * socketmaster starts the given server with the socket on fd 3

On HUP:
 * socketmaster starts a new server
 * waits for X seconds
 * sends a SIGQUIT to the old server

On SIGINT, TERM, QUIT the signal is propagated to the clients.

All old servers are responsible to gracefuly shutdown.

Related projects
================

 * [einhorn](https://github.com/stripe/einhorn): Is much more complete
 in terms of features and is written in ruby.

TODO
====

How to handle socketmaster restarts ?

License (MIT)
=============

Copyright © 2012 PandaStream Ltd.

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the
“Software”), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS
OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
