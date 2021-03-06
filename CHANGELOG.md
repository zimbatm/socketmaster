
1.0.69 / 2019-10-09
==================

  * FIX: call wg.Add(1) synchronously (#15)
  * Since golang 1.8 server supports gracefull shutdown of the server (#14)
  * Merge pull request #13 from alileza/master
  * improvement swap child if the new one is still alive
  * Fix style
  * Merge pull request #9 from vivienschilis/feature-simple-http-server
  * Rename to slave
  * return after first signal
  * Add example server
  * Simple server to use

1.0.57 / 2015-04-11
===================

  * CHANGE: HUP to TERM for shutting down clients. Thanks @qzaidi
  * NEW: allow to pass command-line arguments to the running script. Thanks @qzaidi
  * FIX: Avoid double [pid] tags on the syslog output
  * Doc typos and grammar, thanks @nikai3d and @patrickberkeley

1.0.40 / 2013-06-18
==================

  * Build socketmaster with go v1.1.1 - Fixes weird issues with shared fds.
  * Avoid double [pid] tags on the syslog output
  * Add a -user option to set the child process's uid/gid
  * Add a -syslog option to log to syslog
  * Fixed a race condition in the process set
  * Forward all child process output to the socketmaster logger
  * Fixed a race condition when two signals arrive at the same time
  * Allow to disable the start-wait

1.0.13 / 2012-11-12
===================

  * Prefix output with the [pid]
  * Set EINHORN_FDS to be eninhorn-compatible
  * Adding license and changelog
  * Add a note about other related projects

1.0.7 / 2012-11-11
==================

  * First release with just the basic features

