!SLIDE
# POSIX

the concept

!SLIDE

# Only two ways

* fork / exec
* unix socket + SOCK_DGRAM

!SLIDE
# fork / exec

but also posix_spawn(2)

!SLIDE
# dup2(2)

	int
    dup2(int fildes, int fildes2);

!SLIDE
# fcntl(2) / FD_CLOEXEC

    the given file descriptor will be auto-
    matically closed in the successor process
    image when one of the execv(2) or
    posix_spawn(2) family of system
    calls is invoked.

!SLIDE
# Fake friends

* setsockopt(2) / SO_REUSEADDR
* setsockopt(2) / SO_REUSEPORT

!SLIDE
# signal(3)

	sig_t
    signal(int sig, sig_t func);

# kill(2)

	int
    kill(pid_t pid, int sig);

!SLIDE
# socket(2)

	int
    socket(PF_LOCAL, SOCK_DGRAM, 0);

# sendmsg/recvmsg(2)

	ssize_t
    sendmsg(
    	int socket,
    	const struct msghdr *message,
    	int flags);
