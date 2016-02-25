A utility for socketmaster-managed servers that are written in go for easy zero-downtime restarts.

Use like http.ListenAndServe, except it will bind to the socketmaster-provided file descriptor.

On SIGINT, SIGTERM, SIGQUIT or SIGHUP, this will stop accepting new connections and wait for either all current connections to close or 30 seconds to pass before returning.

socketchild.ListenAndServe(nil)