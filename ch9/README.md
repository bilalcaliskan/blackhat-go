# Writing And Porting Exploit Code
In the majority of the previous chapters, you used Go to create network-based attacks. You’ve 
explored raw TCP, HTTP, DNS, SMB, database interaction, and passive packet capturing.

This chapter focuses instead on identifying and exploiting vulnerabilities. First, you’ll learn how to create a 
vulnerability fuzzer to discover an application’s security weaknesses. Then you’ll learn how to port existing 
exploits to Go. Finally, we’ll show you how to use popular tools to create Go-friendly shellcode. By the end of 
the chapter, you should have a basic understanding of how to use Go to discover flaws while also using it to write 
and deliver various payloads.

### Creating A Fuzzer
[Fuzzing](https://owasp.org/www-community/Fuzzing) is a technique that sends extensive amounts of data to an 
application in an attempt to force the application to produce abnormal behavior. This behavior can reveal coding 
errors or security deficiencies, which you can later exploit.

`Fuzzing an application` can also produce undesirable side effects, such as resource exhaustion, memory corruption, 
and service interruption. Some of these side effects are necessary for bug hunters and exploit developers to do 
their jobs but bad for the stability of the application. Therefore, it’s crucial that you always perform fuzzing in a 
controlled lab environment. As with most of the techniques we discuss in this book, don’t fuzz applications or systems 
without explicit authorization from the owner.

In this section, you’ll build two fuzzers. The first will check the capacity of an input in an attempt to crash a 
service and identify a [buffer overflow](https://owasp.org/www-community/vulnerabilities/Buffer_Overflow). The second 
fuzzer will replay an HTTP request, cycling through potential input values to detect [SQL injection](https://owasp.org/www-community/attacks/SQL_Injection).

Buffer Overflow readings:
  - https://owasp.org/www-community/vulnerabilities/Buffer_Overflow
  - https://owasp.org/www-community/attacks/Buffer_overflow_attack

SQL Injection readings:
  - https://owasp.org/www-community/attacks/SQL_Injection

#### _Buffer Overflow Fuzzing_
Buffer overflows occur when a user submits more data in an input than the application has allocated memory space for. 
For example, a user could submit 5,000 characters when the application expects to receive only 5. If a program uses the 
wrong techniques, this could allow the user to write that surplus data to parts of memory that aren’t intended for that 
purpose. This “overflow” corrupts the data stored within adjacent memory locations, allowing a malicious user to 
potentially crash the program or alter its logical flow.

Buffer overflows are particularly impactful for network-based programs that receive data from clients. Using buffer 
overflows, a client can disrupt server availability or possibly achieve remote code execution.

##### How Buffer Overflow Fuzzing Works
Fuzzing to create a buffer overflow generally involves submitting increasingly longer inputs, such that each subsequent 
request includes an input value whose length is one character longer than the previous attempt. A contrived example 
using the A character as input would execute according to the pattern shown in below diagram.

By sending numerous inputs to a vulnerable function, you’ll eventually reach a point where the length of your input 
exceeds the function’s defined buffer size, which will corrupt the program’s control elements, such as its return and 
instruction pointers. At this point, the application or system will crash.

By sending incrementally larger requests for each attempt, you can precisely determine the expected input size, which 
is important for exploiting the application later. You can then inspect the crash or resulting core dump to better 
understand the vulnerability and attempt to develop a working exploit. We won’t go into debugger usage and exploit 
development here; instead, let’s focus on writing the fuzzer.

You can also read more about Core dump in below links:
  - https://wiki.archlinux.org/index.php/Core_dump
  - https://en.wikipedia.org/wiki/Core_dump

```
Attempt     Input Value
1           A
2           AA
3           AAA
N           A repeated N times
```

If you’ve done any manual fuzzing using modern, interpreted languages, you’ve probably used a construct to create 
strings of specific lengths. For example, the following Python code, run within the interpreter console, shows how 
simple it is to create a string of 25 A characters:
```python
>>> x = "A"*25
>>> x
'AAAAAAAAAAAAAAAAAAAAAAAAA'
```

Unfortunately, Go has no such construct to conveniently build strings of arbitrary length. You’ll have to do that the 
old-fashioned way—using a loop—which would look something like this:
```go
var (
        n int
        s string
)
for n = 0; n < 25; n++ {
    s += "A"
}
```

Sure, it’s a little more verbose than the Python alternative, but not overwhelming.

The other consideration you’ll need to make is the delivery mechanism for your payload. This will depend on the target 
application or system. In some instances, this could involve writing a file to a disk. In other cases, you might 
communicate over TCP/UDP with an HTTP, SMTP, SNMP, FTP, Telnet, or other networked service.

In the following example, you’ll perform fuzzing against a `remote FTP server`. You can tweak a lot of the logic we 
present fairly quickly to operate against other protocols, so it should act as a good basis for you to develop custom 
fuzzers against other services.

Although Go’s standard packages include support for some common protocols, such as HTTP and SMTP, they don’t include 
support for client-server FTP interactions. Instead, you could use a third-party package that already performs FTP 
communications, so you don’t have to reinvent the wheel and write something from the ground up. However, for maximum 
control (and to appreciate the protocol), you’ll instead build the basic FTP functionality using raw TCP communications. 
If you need a refresher on how this works, refer to [Chapter 2](../ch2).

##### Building The Buffer Overflow Fuzzer
[ftp-fuzz/main.go](ftp-fuzz/main.go) shows the fuzzer code. We’ve hardcoded some values, such as the target IP and 
port, as well as the maximum length of your input. The code itself fuzzes the `USER` property. Since this property 
occurs before a user is authenticated, it represents a commonly testable point on the attack surface. You could 
certainly extend this code to test other pre-authentication commands, such as `PASS`, but keep in mind that if you 
supply a legitimate username and then keep submitting inputs for PASS, you might get locked out eventually.
```go
func main() {
  ❶ for i := 0; i < 2500; i++ {
      ❷ conn, err := net.Dial("tcp", "test.rebex.net:22")
         if err != nil {
          ❸ log.Fata lf("[!] Error at offset %d: %s\n", i, err)
         }  
      ❹ bufio.NewReader(conn).ReadString('\n')

         user := ""
      ❺ for n := 0; n <= i; n++ {
             user += "A"
          }  

         raw := "USER %s\n"
      ❻ fmt.Fprintf(conn, raw, user)
         bufio.NewReader(conn).ReadString('\n')

         raw = "PASS password\n"
         fmt.Fprint(conn, raw)
         bufio.NewReader(conn).ReadString('\n')

         if err := conn.Close()❼; err != nil {
          ❽ log.Println("[!] Error at offset %d: %s\n", i, err)
         }  
    }  
}
```
The code is essentially one large loop, beginning at ❶. Each time the program loops, it adds another character to the 
username you’ll supply. In this case, you’ll send usernames from 1 to 2,500 characters in length.

For each iteration of the loop, you establish a TCP connection to the destination FTP server ❷. Any time you interact 
with the FTP service, whether it’s the initial connection or the subsequent commands, you explicitly read the response 
from the server as a single line ❹. This allows the code to block while waiting for the TCP responses so you don’t 
send your commands prematurely, before packets have made their round trip. You then use another for loop to build the 
string of As in the manner we showed previously ❺. You use the index i of the outer loop to build the string length 
dependent on the current iteration of the loop, so that it increases by one each time the program starts over. You use 
this value to write the USER command by using fmt.Fprintf(conn, raw, user) ❻.

Although you could end your interaction with the FTP server at this point (after all, you’re fuzzing only the USER 
command), you proceed to send the PASS command to complete the transaction. Lastly, you close your connection cleanly ❼.

It’s worth noting that there are two points, ❸ and ❽, where abnormal connectivity behavior could indicate a service 
disruption, implying a potential buffer overflow: when the connection is first established and when the connection 
closes. If you can’t establish a connection the next time the program loops, it’s likely that something went wrong. 
You’ll then want to check whether the service crashed as a result of a buffer overflow.

If you can’t close a connection after you’ve established it, this may indicate the abnormal behavior of the remote 
FTP service abruptly disconnecting, but it probably isn’t caused by a buffer overflow. The anomalous condition is 
logged, but the program will continue.

A packet capture, illustrated in below picture, shows that each subsequent USER command grows in length, confirming that 
your code works as desired.
![Figure 9-1: A Wireshark capture depicting the USER command growing by one letter each time the program loops](resources/wireshark.jpg)

You could improve the code in several ways for flexibility and convenience. For example, you’d probably want to remove 
the hardcoded IP, port, and iteration values, and instead include them via command line arguments or a configuration 
file. We invite you to perform these usability updates as an exercise. Furthermore, you could extend the code so it 
fuzzes commands after authentication. Specifically, you could update the tool to fuzz the CWD/CD command. Various tools 
have historically been susceptible to buffer overflows related to the handling of this command, making it a good target 
for fuzzing.

#### _SQL Injection Fuzzing_
In this section, you’ll explore SQL injection fuzzing. Instead of changing the length of each input, this variation on 
the attack cycles through a defined list of inputs to attempt to cause SQL injection. In other words, you’ll fuzz the 
username parameter of a website login form by attempting a list of inputs consisting of various SQL meta-characters and 
syntax that, if handled insecurely by the backend database, will yield abnormal behavior by the application.

To keep things simple, you’ll be probing only for error-based SQL injection, ignoring other forms, such as boolean-, 
time-, and union-based. That means that instead of looking for subtle differences in response content or response time, 
you’ll look for an error message in the HTTP response to indicate a SQL injection. This implies that you expect the 
web server to remain operational, so you can no longer rely on connection establishment as a litmus test for whether 
you’ve succeeded in creating abnormal behavior. Instead, you’ll need to search the response body for a database error 
message.

##### How SQL Injection WOrks
At its core, SQL injection allows an attacker to insert SQL meta-characters into a statement, potentially manipulating 
the query to produce unintended behavior or return restricted, sensitive data. The problem occurs when developers 
blindly concatenate untrusted user data to their SQL queries, as in the following pseudocode:
```go
username = HTTP_GET["username"]
query = "SELECT * FROM users WHERE user = '" + username + "'"
result = db.execute(query)
if(len(result) > 0) {
    return AuthenticationSuccess()
} else {
    return AuthenticationFailed()
}
```
In our pseudocode, the username variable is read directly from an HTTP parameter. The value of the username variable 
isn’t sanitized or validated. You then build a query string by using the value, concatenating it onto the SQL query 
syntax directly. The program executes the query against the database and inspects the result. If it finds at least 
one matching record, you’d consider the authentication successful. The code should behave appropriately so long as 
the supplied username consists of alphanumeric and a certain subset of special characters. For example, supplying 
a username of alice results in the following safe query:
```sql
SELECT * FROM users WHERE user = 'alice'
```

However, what happens when the user supplies a username containing an apostrophe? Supplying a username of o'doyle 
produces the following query:
```sql
SELECT * FROM users WHERE user = 'o'doyle'
```

The problem here is that the backend database now sees an unbalanced number of single quotation marks. Notice the 
emphasized portion of the preceding query, doyle; the backend database interprets this as SQL syntax, since it’s 
outside the enclosing quotes. This, of course, is invalid SQL syntax, and the backend database won’t be able to 
process it. For error-based SQL injection, this produces an error message in the HTTP response. The message itself 
will vary based on the database. In the case of MySQL, you’ll receive an error similar to the following, possibly 
with additional details disclosing the query itself:
```shell script
You have an error in your SQL syntax
```

Although we won’t go too deeply into exploitation, you could now manipulate the username input to produce a valid SQL 
query that would bypass the authentication in our example. The username input ' OR 1=1# does just that when placed in 
the following SQL statement:
```sql
SELECT * FROM users WHERE user = '' OR 1=1#'
```

This input appends a logical OR onto the end of the query. This OR statement always evaluates to true, because 1 always 
equals 1. You then use a MySQL comment (#) to force the backend database to ignore the remainder of the query. This 
results in a valid SQL statement that, assuming one or more rows exist in the database, you can use to bypass authentication 
in the preceding pseudocode example.

##### Building the SQL Injection Fuzzer