# Raw Packet Processing

In this chapter, you’ll learn how to capture and process network packets. You can use packet processing for many purposes, 
including to capture cleartext authentication credentials, alter the application functionality of the packets, or spoof 
and poison traffic. You can also use it for [SYN scanning](https://searchnetworking.techtarget.com/definition/SYN-scanning#:~:text=SYN%20scanning%20is%20a%20tactic,%2Dservice%20(DoS)%20attacks.) 
and for port scanning through SYN-flood protections, among other things.

> **What is SYN Flood attack?**  
> A SYN flood is a form of denial-of-service attack in which an attacker rapidly initiates a connection to a server 
> without finalizing the connection. The server has to spend resources waiting for half-opened connections, which can 
> consume enough resources to make the system unresponsive to legitimate traffic.
> 
> For more, read [here](https://en.wikipedia.org/wiki/SYN_flood#:~:text=A%20SYN%20flood%20is%20a,system%20unresponsive%20to%20legitimate%20traffic.).

We’ll introduce you to the excellent [gopacket package from Google](https://pkg.go.dev/github.com/google/gopacket), which 
will enable you to both decode packets and reassemble the stream of traffic. This package allows you to filter traffic 
by using the [Berkeley Packet Filter (BPF)](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter), also called tcpdump 
syntax; read and write .pcap files; inspect various layers and data; and manipulate packets.

### Setting Up Your Environment
We will use the directory [identify](identify) for our first coding exercise. 

Before working through the code in this chapter, you need to set up your environment. First, install gopacket by 
entering the following:
```shell script
$ go get github.com/google/gopacket
```

Now, gopacket relies on external libraries and drivers to bypass the operating system’s protocol stack. If you intend 
to compile the examples in this chapter for use on Linux or macOS, you’ll need to install `libpcap-dev`. You can do 
this with most package management utilities such as apt, yum, or brew. Here’s how you install it by using apt (the 
installation process looks similar for the other two options):
```shell script
$ sudo apt-get install libpcap-dev
```

If you intend to compile and run the examples in this chapter on Windows, you have a couple of options, based on 
whether you’re going to cross-compile or not. Setting up a development environment is simpler if you don’t cross-compile, 
but in that case, you’ll have to create a Go development environment on a Windows machine, which can be unattractive if 
you don’t want to clutter another environment. For the time being, we’ll assume you have a working environment that you 
can use to compile Windows binaries. Within this environment, you’ll need to install WinPcap. You can download an 
installer for free from https://www.winpcap.org/.

### Identifying Devices By Using The Pcap Subpackage
Before you can capture network traffic, you must identify available devices on which you can listen. You can do this 
easily using the `gopacket/pcap` subpackage, which retrieves them with the following helper function: `pcap.FindAllDevs() 
(ifs []Interface, err error)`.

[identify/main.go](identify/main.go) is the coding exercise for that:
```go
package main

import (
    "fmt"
    "log"

    "github.com/google/gopacket/pcap"
)

func main() {
 ❶ devices, err := pcap.FindAllDevs()
    if err != nil {
        log.Panicln(err)
    }
 ❷ for _, device := range devices {
        fmt.Println(device.Name❸)
     ❹ for _, address := range device.Addresses {
         ❺ fmt.Printf("    IP:      %s\n", address.IP)
            fmt.Printf("    Netmask: %s\n", address.Netmask)
        }  
    }
}
```

You enumerate your devices by calling `pcap.FindAllDevs()` ❶. Then you loop through the devices found ❷. For each 
device, you access various properties, including the `device.Name` ❸. You also access their IP addresses through 
the `Addresses` property, which is a slice of type `pcap.InterfaceAddress`. You loop through these addresses ❹, 
displaying the IP address and netmask to the screen ❺.

Executing your utility produces output similar to below:
```shell script
$ go run main.go
enp0s5
    IP:      10.0.1.20
    Netmask: ffffff00
    IP:      fe80::553a:14e7:92d2:114b
    Netmask: ffffffffffffffff0000000000000000
any
lo
    IP:      127.0.0.1
    Netmask: ff000000
    IP:      ::1
    Netmask: ffffffffffffffffffffffffffffffff
```

The output lists the available network interfaces—enp0s5, any, and lo—as well as their IPv4 and IPv6 addresses and 
netmasks. The output on your system will likely differ from these network details, but it should be similar enough 
that you can make sense of the information.

### Live Capturing And Filtering Results
Now that you know how to query available devices, you can use gopacket’s features to capture live packets off the wire. 
In doing so, you’ll also filter the set of packets by using BPF syntax. BPF allows you to limit the contents of what 
you capture and display so that you see only relevant traffic. It’s commonly used to filter traffic by protocol and 
port. For example, you could create a filter to see all TCP traffic destined for port 80. You can also filter traffic 
by destination host. A full discussion of BPF syntax is beyond the scope of this book. For additional ways to use BPF, 
take a peek at http://www.tcpdump.org/manpages/pcap-filter.7.html.

We will use the directory [filter](filter) for our second coding exercise.

 [filter/main.go](filter/main.go) shows the code, which filters traffic so that you capture only TCP traffic sent to 
 or from port 80.
 ```go
   package main
  
   import (
       "fmt"
       "log"
  
       "github.com/google/gopacket"
       "github.com/google/gopacket/pcap"
   )
  
❶ var (
       iface    = "enp0s5"
       snaplen  = int32(1600)
       promisc  = false
       timeout  = pcap.BlockForever
       filter   = "tcp and port 80"
       devFound = false
   )  
      
   func main() {
       devices, err := pcap.FindAllDevs()❷
       if err != nil {
           log.Panicln(err)
       }
      
    ❸ for _, device := range devices {
           if device.Name == iface {
               devFound = true
           }
       }
       if !devFound {
           log.Panicf("Device named '%s' does not exist\n", iface)
       }
      
     ❹ handle, err := pcap.OpenLive(iface, snaplen, promisc, timeout)
       if err != nil {
           log.Panicln(err)
       }
       defer handle.Close()
          
    ❺ if err := handle.SetBPFFilter(filter); err != nil {
           log.Panicln(err)
       }
      
    ❻ source := gopacket.NewPacketSource(handle, handle.LinkType())
       for packet := range source.Packets()❼ {
           fmt.Println(packet)
       }
   }
```

The code starts by defining several variables necessary to set up the packet capture ❶. Included among these is the 
name of the interface on which you want to capture data, the snapshot length (the amount of data to capture for each 
frame), the promisc variable (which determines whether you’ll be running promiscuous mode), and your time-out. Also, 
you define your BPF filter: `tcp and port 80`. This will make sure you capture only packets that match those criteria.

Within your main() function, you enumerate the available devices ❷, looping through them to determine whether 
your desired capture interface exists in your device list ❸. If the interface name doesn’t exist, then you panic, 
stating that it’s invalid.

What remains in the rest of the main() function is your capturing logic. From a high-level perspective, you need to 
first obtain or create a `*pcap.Handle`, which allows you to read and inject packets. Using this handle, you can 
then apply a BPF filter and create a new packet data source, from which you can read your packets.

You create your `*pcap.Handle` (named handle in the code) by issuing a call to `pcap.OpenLive()` ❹. This function 
receives an interface name, a snapshot length, a boolean value defining whether it’s promiscuous, and a time-out 
value. These input variables are all defined prior to the main() function, as we detailed previously. Call `handle.SetBPFFilter(filter)` 
to set the BPF filter for your handle ❺, and then use handle as an input while calling `gopacket.NewPacketSource(handle, handle.LinkType())` 
to create a new packet data source ❻. The second input value, `handle.LinkType()`, defines the decoder to use when 
handling packets. Lastly, you actually read packets from the wire by using a loop on `source.Packets()` ❼, which 
returns a channel.

As you might recall from previous examples in this book, looping on a channel causes the loop to block when it has no 
data to read from the channel. When a packet arrives, you read it and print its contents to screen.

The output should look like [](filter/main.go). Note that the program requires elevated privileges because we’re 
reading raw content off the network.

```shell script
$ go build -o filter && sudo ./filter
PACKET: 74 bytes, wire length 74 cap length 74 @ 2020-04-26 08:44:43.074187 -0500 CDT
- Layer 1 (14 bytes) = Ethernet   {Contents=[..14..] Payload=[..60..]
SrcMAC=00:1c:42:cf:57:11 DstMAC=90:72:40:04:33:c1 EthernetType=IPv4 Length=0}
- Layer 2 (20 bytes) = IPv4       {Contents=[..20..] Payload=[..40..] Version=4 IHL=5
TOS=0 Length=60 Id=998 Flags=DF FragOffset=0 TTL=64 Protocol=TCP Checksum=55712
SrcIP=10.0.1.20 DstIP=54.164.27.126 Options=[] Padding=[]}
- Layer 3 (40 bytes) = TCP        {Contents=[..40..] Payload=[] SrcPort=51064
DstPort=80(http) Seq=3543761149 Ack=0 DataOffset=10 FIN=false SYN=true RST=false
PSH=false ACK=false URG=false ECE=false CWR=false NS=false Window=29200
Checksum=23908 Urgent=0 Options=[..5..] Padding=[]}

PACKET: 74 bytes, wire length 74 cap length 74 @ 2020-04-26 08:44:43.086706 -0500 CDT
- Layer 1 (14 bytes) = Ethernet   {Contents=[..14..] Payload=[..60..]
SrcMAC=00:1c:42:cf:57:11 DstMAC=90:72:40:04:33:c1 EthernetType=IPv4 Length=0}
- Layer 2 (20 bytes) = IPv4       {Contents=[..20..] Payload=[..40..] Version=4 IHL=5
TOS=0 Length=60 Id=23414 Flags=DF FragOffset=0 TTL=64 Protocol=TCP Checksum=16919
SrcIP=10.0.1.20 DstIP=204.79.197.203 Options=[] Padding=[]}
- Layer 3 (40 bytes) = TCP        {Contents=[..40..] Payload=[] SrcPort=37314
DstPort=80(http) Seq=2821118056 Ack=0 DataOffset=10 FIN=false SYN=true RST=false
PSH=false ACK=false URG=false ECE=false CWR=false NS=false Window=29200
Checksum=40285 Urgent=0 Options=[..5..] Padding=[]}
```

Although the raw output isn’t very digestible, it certainly contains a nice separation of each layer. You can now 
use utility functions, such as `packet.ApplicationLayer()` and `packet.Data()`, to retrieve the raw bytes for a 
single layer or the entire packet. When you combine the output with `hex.Dump()`, you can display the contents in a 
much more readable format. Play around with this on your own.

### Sniffing And Displaying Cleartext User Credentials
Now let’s build on the code you just created. You’ll replicate some of the functionality provided by other tools to 
sniff and display cleartext user credentials.

Most organizations now operate by using switched networks, which send data directly between two endpoints rather than 
as a broadcast, making it harder to passively capture traffic in an enterprise environment. However, the following 
cleartext sniffing attack can be useful when paired with something like [Address Resolution Protocol (ARP) poisoning, 
an attack that can coerce endpoints into communicating with a malicious device on a switched network, or when you’re 
covertly sniffing outbound traffic from a compromised user workstation. In this example, we’ll assume you’ve compromised 
a user workstation and focus solely on capturing traffic that uses FTP to keep the code brief.

Here is our code in [ftp/main.go](ftp/main.go):
```go
package main

import (
    "bytes"
    "fmt"
    "log"

    "github.com/google/gopacket"
    "github.com/google/gopacket/pcap"
)

var (
    iface    = "enp0s5"
    snaplen  = int32(1600)
    promisc  = false
    timeout  = pcap.BlockForever
 ❶ filter   = "tcp and dst port 21"
    devFound = false
)

func main() {
    devices, err := pcap.FindAllDevs()
    if err != nil {
        log.Panicln(err)
    }

    for _, device := range devices {
        if device.Name == iface {
            devFound = true
        }
    }
    if !devFound {
        log.Panicf("Device named '%s' does not exist\n", iface)
    }

    handle, err := pcap.OpenLive(iface, snaplen, promisc, timeout)
    if err != nil {
        log.Panicln(err)
    }
    defer handle.Close()

    if err := handle.SetBPFFilter(filter); err != nil {
        log.Panicln(err)
    }

    source := gopacket.NewPacketSource(handle, handle.LinkType())
    for packet := range source.Packets() {
     ❷ appLayer := packet.ApplicationLayer()
        if appLayer == nil {
            continue
        }  
     ❸ payload := appLayer.Payload()
     ❹ if bytes.Contains(payload, []byte("USER")) {
            fmt.Print(string(payload))
        } else if bytes.Contains(payload, []byte("PASS")) {
            fmt.Print(string(payload))
        }  
    }
}
```

The changes you made encompass only about 10 lines of code. First, you change your BPF filter to capture only traffic 
destined for port 21 (the port commonly used for FTP traffic) ❶. The rest of the code remains the same until you 
process the packets.

To process packets, you first extract the application layer from the packet and check to see whether it actually 
exists ❷, because the application layer contains the FTP commands and data. You look for the application layer by 
examining whether the response value from `packet.ApplicationLayer()` is nil. Assuming the application layer exists in 
the packet, you extract the payload (the FTP commands/data) from the layer by calling `appLayer.Payload()` ❸. (There 
are similar methods for extracting and inspecting other layers and data, but you only need the application layer payload.) 
With your payload extracted, you then check whether the payload contains either the USER or PASS commands ❹, indicating 
that it’s part of a login sequence. If it does, display the payload to the screen.

Here’s a sample run that captures an FTP login attempt:
```shell script
$ go build -o ftp && sudo ./ftp
USER someuser
PASS passw0rd
```

Of course, you can improve this code. In this example, the payload will be displayed if the words USER or PASS exist 
anywhere in the payload. Really, the code should be searching only the beginning of the payload to eliminate false-positives 
that occur when those keywords appear as part of file contents transferred between client and server or as part of a 
longer word such as PASSAGE or ABUSER. We encourage you to make these improvements as a learning exercise.

### Port Scanning Through Syn-Flood Protections
In [Chapter 2](../ch2), you walked through the creation of a port scanner. You improved the code through multiple 
iterations until you had a high-performing implementation that produced accurate results. However, in some instances, 
that scanner can still produce incorrect results. Specifically, when an organization employs SYN-flood protections, 
typically all ports—open, closed, and filtered alike—produce the same packet exchange to indicate that the port is 
open. These protections, known as SYN cookies, prevent [SYN-flood attacks](https://www.imperva.com/learn/ddos/syn-flood/) and obfuscate the attack surface, producing 
false-positives.

> **What are [SYN Cookies](https://en.wikipedia.org/wiki/SYN_cookies)?**  
> SYN cookie is a technique used to resist SYN flood attacks. use of SYN cookies allows a server to avoid dropping 
> connections when the SYN queue fills up. Instead of storing additional connections, the SYN queue entry is encoded 
> into the sequence number sent in the SYN+ACK response. If the server then receives a subsequent ACK response from 
> the client with the incremented sequence number, the server is able to reconstruct the SYN queue entry using 
> information encoded in the TCP sequence number and proceed as usual with the connection.