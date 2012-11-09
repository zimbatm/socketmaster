socketmaster: Zero-downtime restarts for your apps
==================================================

socketmaster takes care of starting your app while sharing the
file-descriptor.



Q:: How is it different than the Nginx dance ?
A:: The parent PID never changes so it works better with tools like
Upstart

Q:: What happens if I need to change the command to be executed ?
A:: Use a wrapper script:

```
#!/bin/sh
exec /path/to/your/program
```

Usage
=====

```
socketmaster -listen=tcp://:8080 path/to/wrapper/script
```

Design
======

 * socketmaster opens the port
 * socketmaster starts the given server with the socket on fd 3

On HUP:
 * socketmaster starts a new server
 * waits for X seconds
 * if the new server didn't exit, sent a SIGQUIT to the old server

On SIGINT, TERM, QUIT the signal is propagated to the clients.

All old servers are responsible to gracefuly shutdown.

TODO
====

How to handle socketmaster restarts ?
