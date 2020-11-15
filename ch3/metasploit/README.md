## Interacting With Metasploit
[Metasploit](https://www.metasploit.com/) is a framework used to perform a variety of adversarial techniques, 
including reconnaissance, exploitation, command and control, persistence, lateral network movement, payload 
creation and delivery, privilege escalation, and more. Even better, the community version of the product is 
free, runs on Linux and macOS, and is actively maintained. Essential for any adversarial engagement, 
Metasploit is a fundamental tool used by penetration testers, and it exposes a `remote procedure call 
(RPC) API` to allow remote interaction with its functionality.

In this section, you’ll build a client that interacts with a remote Metasploit instance. Much like the 
[Shodan code you built](../shodan), the Metasploit client you develop won’t cover a comprehensive 
implementation of all available functionality. Rather, it will be the foundation upon which you can 
extend additional functionality as needed. We think you’ll find the implementation more complex than 
the [Shodan](../shodan) example, making the Metasploit interaction a more challenging progression.

###Setting Up Your Environment
Before you proceed with this section, download and install the Metasploit community edition if you don’t 
already have it. Start the Metasploit console as well as the RPC listener through the msgrpc module in 
Metasploit. Then set the server host—the IP on which the RPC server will listen—and a password, as shown in 
below:
```shell script
$ msfconsole
msf > load msgrpc Pass=s3cr3t ServerHost=10.0.1.6
[*] MSGRPC Service:  10.0.1.6:55552
[*] MSGRPC Username: msf
[*] MSGRPC Password: s3cr3t
[*] Successfully loaded plugin: msgrpc
```
To make the code more portable and avoid hardcoding values, set the following environment variables to 
the values you defined for your RPC instance. This is similar to what you did for the Shodan API key 
used to interact with Shodan:
```shell script
$ export MSFHOST=10.0.1.6:55552
$ export MSFPASS=s3cr3t
```
You should now have Metasploit and the RPC server running.
