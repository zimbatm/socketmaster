#!/usr/bin/ruby

require 'thread'
require 'socket'
require 'set'

Thread.abort_on_exception = true

def log(*a)
  p [Process.pid, *a]
end

class SlowServer
  def initialize(server)
    @server = server
    @tg = ThreadGroup.new
  end

  def start
    while true
      Thread.new(@server.accept) do |client|
        @tg.add Thread.current
        Thread.current[:client] = client
        handle_client(client)
      end
    end
  rescue => ex
    log ex
  end

  def stop
    @server.close unless @server.closed?
  end

  def wait_for_clients
    while @tg.list.size > 0
      sleep 0.5
    end
  end

  protected

  def handle_client(client)
    log "New connection"

    log client.gets

    sleep 10

    client.write "HTTP/1.1 200 OK\r\n"
    client.write "Content-Type: text/plain\r\n"
    client.write "Content-Length: 2\r\n"
    client.write "\r\n"
    client.write "Hi"
  rescue Errno::EPIPE
    log "Client closed the connection"
  ensure
    client.close unless client.closed?
  end

end

conn = TCPServer.for_fd(3)

server = SlowServer.new(conn)

trap("TERM") { server.stop }
trap("QUIT") { server.stop }

log "Server waiting for connections"
server.start
log "Waiting for clients to die"
server.wait_for_clients
log "Bye"

