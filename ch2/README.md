### TCP, Scanners And Proxies
Let’s begin our practical application of Go with the Transmission Control Protocol (TCP), 
the predominant standard for connection-oriented, reliable communications and the foundation 
of modern networking. TCP is everywhere, and it has well-documented libraries, code samples, 
and generally easy-to-understand packet flows. You must understand TCP to fully evaluate, 
analyze, query, and manipulate network traffic.

As an attacker, you should understand how TCP works and be able to develop usable TCP 
constructs so that you can identify open/closed ports, recognize potentially errant results 
such as false-positives—for example, syn-flood protections—and bypass egress restrictions 
through port forwarding. In this chapter, you’ll learn basic TCP communications in Go; 
build a concurrent, properly throttled port scanner; create a TCP proxy that can be used 
for port forwarding; and re-create Netcat’s `gaping security hole` feature.

### Understanding The TCP Handshake
If the port is open, a three-way handshake takes place. First, the client sends a syn packet, 
which signals the beginning of a communication. The server then responds with a syn-ack, 
or acknowledgment of the syn packet it received, prompting the client to finish with an ack, 
or acknowledgment of the server’s response. The transfer of data can then occur. If the port 
is closed, the server responds with a rst packet instead of a syn-ack. If the traffic is being
filtered by a firewall, the client will typically receive no response from the server.


For the coding exercises, check the subfolders in current directory.