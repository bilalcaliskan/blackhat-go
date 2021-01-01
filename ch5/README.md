# Exploiting DNS

The `Domain Name System (DNS)` locates internet domain names and translates them to IP addresses. It can be an effective 
weapon in the hands of an attacker, because organizations commonly allow the protocol to egress restricted networks and 
they frequently fail to monitor its use adequately. It takes a little knowledge, but savvy attackers can leverage these 
issues throughout nearly every step of an attack chain, including reconnaissance, command and control (C2), and even 
data exfiltration. In this chapter, you’ll learn how to write your own utilities by using Go and third-party packages 
to perform some of these capabilities. You’ll start by resolving hostnames and IP addresses to reveal the many types 
of DNS records that can be enumerated. Then you’ll use patterns illustrated in earlier chapters to build a massively 
concurrent subdomain-guessing tool. Finally, you’ll learn how to write your own DNS server and proxy, and you’ll use 
DNS tunneling to establish a C2 channel out of a restrictive network!

### Writing DNS Clients
