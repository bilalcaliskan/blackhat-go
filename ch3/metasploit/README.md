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

Because the details on exploitation and Metasploit use are beyond the scope of this book, let’s assume that through 
pure cunning and trickery you’ve already compromised a remote Windows system and you’ve leveraged Metasploit’s 
Meterpreter payload for advanced post-exploitation activities. Here, your efforts will instead focus on how you can 
remotely communicate with Metasploit to list and interact with established Meterpreter sessions. As we mentioned before, 
this code is a bit more cumbersome, so we’ll purposely pare it back to the bare minimum—just enough for you to take the 
code and extend it for your specific needs.

Follow the same project roadmap as the Shodan example: review the Metasploit API, lay out the project in library 
format, define data types, implement client API functions, and, finally, build a test rig that uses the library.
First, review the Metasploit API developer documentation at [Rapid7’s official website](https://metasploit.help.rapid7.com/docs/rpc-api/). 
The functionality exposed is extensive, allowing you to do just about anything remotely that you could through local 
interaction. Unlike Shodan, which uses JSON, Metasploit communicates using `MessagePack`, a compact and efficient binary 
format. Because Go doesn’t contain a standard MessagePack package, you’ll use a full-featured community implementation. 
Install it by executing the following from the command line:
```shell script
$ go get gopkg.in/vmihailenco/msgpack.v2
```

Next, create your directory structure. For this example, you use only two Go files:
```shell script
$ tree .
.
|---client
|   |---main.go
|---rpc
    |---msf.go
```
The `msf.go` file resides within the rpc package, and you’ll use `client/main.go` to implement and test the library you build.


### Defining Your Objective
Now, you need to define your objective. For the sake of brevity, implement the code to interact and issue an RPC call 
that retrieves a listing of current Meterpreter sessions—that is, the session.list method from the Metasploit developer 
documentation. The request format is defined as follows:
```
[ "session.list", "token" ]
```
This is minimal; it expects to receive the name of the method to implement and a token. The token value is a placeholder. 
If you read through the documentation, you’ll find that this is an authentication token, issued upon successful login to 
the RPC server. The response returned from Metasploit for the session.list method follows this format:
```json
{
"1" => {
    'type' => "shell",
    "tunnel_local" => "192.168.35.149:44444",
    "tunnel_peer" => "192.168.35.149:43886",
    "via_exploit" => "exploit/multi/handler",
    "via_payload" => "payload/windows/shell_reverse_tcp",
    "desc" => "Command shell",
    "info" => "",
    "workspace" => "Project1",
    "target_host" => "",
    "username" => "root",
    "uuid" => "hjahs9kw",
    "exploit_uuid" => "gcprpj2a",
    "routes" => [ ]
    }
}
```

Let’s build the Go types to handle both the request and response data. [Following piece of code](rpc/msf.go) defines 
the sessionListReq and SessionListRes:
```go
❶ type sessionListReq struct {
    ❷ _msgpack struct{} `msgpack:",asArray"`
       Method   string
       Token    string
   }

❸ type SessionListRes struct {
       ID          uint32 `msgpack:",omitempty"`❹
       Type        string `msgpack:"type"`
       TunnelLocal string `msgpack:"tunnel_local"`
       TunnelPeer  string `msgpack:"tunnel_peer"`
       ViaExploit  string `msgpack:"via_exploit"`
       ViaPayload  string `msgpack:"via_payload"`
       Description string `msgpack:"desc"`
       Info        string `msgpack:"info"`
       Workspace   string `msgpack:"workspace"`
       SessionHost string `msgpack"session_host"`
       SessionPort int    `msgpack"session_port"`
       Username    string `msgpack:"username"`
       UUID        string `msgpack:"uuid"`
       ExploitUUID string `msgpack:"exploit_uuid"`
}
```
You use the request type, sessionListReq ❶, to serialize structured data to the MessagePack format in a manner 
consistent with what the Metasploit RPC server expects—specifically, with a method name and token value. Notice 
that there aren’t any descriptors for those fields. The data is passed as an array, not a map, so rather than 
expecting data in key/value format, the RPC interface expects the data as a positional array of values. This is 
why you omit annotations for those properties—no need to define the key names. However, by default, a structure 
will be encoded as a map with the key names deduced from the property names. To disable this and force the encoding 
as a positional array, you add a special field named _msgpack that utilizes the asArray descriptor ❷, to explicitly 
instruct an encoder/decoder to treat the data as an array.

The SessionListRes type ❸ contains a one-to-one mapping between response field and struct properties. The data, as 
shown in the preceding example response, is essentially a nested map. The outer map is the session identifier to 
session details, while the inner map is the session details, represented as key/value pairs. Unlike the request, 
the response isn’t structured as a positional array, but each of the struct properties uses descriptors to 
explicitly name and map the data to and from Metasploit’s representation. The code includes the session identifier 
as a property on the struct. However, because the actual value of the identifier is the key value, this will be 
populated in a slightly different manner, so you include the omitempty descriptor ❹ to make the data optional so 
that it doesn’t impact encoding or decoding. This flattens the data so you don’t have to work with nested maps.


### Retrieving a Valid Token
Now, you have only one thing outstanding. You have to retrieve a valid token value to use for that request. To do so, 
you’ll issue a login request for the auth.login() API method, which expects the following:
```shell script
["auth.login", "username", "password"]
```
You need to replace the username and password values with what you used when loading the msfrpc module in Metasploit 
during initial setup (recall that you set them as environment variables). Assuming authentication is successful, the 
server responds with the following message, which contains an authentication token you can use for subsequent requests.
```shell script
{ "result" => "success", "token" => "a1a1a1a1a1a1a1a1" }
```
An authentication failure produces the following response:
```shell script
{
    "error" => true,
    "error_class" => "Msf::RPC::Exception",
    "error_message" => "Invalid User ID or Password"
}
```
For good measure, let’s also create functionality to expire the token by logging out. The request takes the method name, 
the authentication token, and a third optional parameter that you’ll ignore because it’s unnecessary for this scenario:
```
[ "auth.logout", "token", "logoutToken"]
```
A successful response looks like this:
```
{ "result" => "success" }
```


### Defining Request and Response Methods
Much as you structured the Go types for the session.list() method’s request and response, you need to do the same for 
both `auth.login()` and `auth.logout()`. The same reasoning applies as before, using descriptors to force requests to be 
serialized as arrays and for the responses to be treated as maps. [Here is the code](rpc/msf.go):
```go
type loginReq struct {
    _msgpack struct{} `msgpack:",asArray"`
    Method   string
    Username string
    Password string
}

type loginRes struct {
    Result       string `msgpack:"result"`
    Token        string `msgpack:"token"`
    Error        bool   `msgpack:"error"`
    ErrorClass   string `msgpack:"error_class"`
    ErrorMessage string `msgpack:"error_message"`
}

type logoutReq struct {
    _msgpack    struct{} `msgpack:",asArray"`
    Method      string
    Token       string
    LogoutToken string
}

type logoutRes struct {
    Result string `msgpack:"result"`
}
```
**It’s worth noting that Go dynamically serializes the login response, populating only the fields present, which means you can represent both successful and failed logins by using a single struct format.**

### Creating a Configuration Struct and an RPC Method
In [rpc/msf.go](rpc/msf.go), you take the defined types and actually use them, creating the necessary methods to issue RPC commands 
to Metasploit. Much as in the Shodan example, you also define an arbitrary type for maintaining pertinent configuration 
and authentication information. That way, you won’t have to explicitly and repeatedly pass in common elements such as 
host, port, and authentication token. Instead, you’ll use the type and build methods on it so that data is implicitly 
available.
```go
type Metasploit struct {
    host  string
    user  string
    pass  string
    token string
}

func New(host, user, pass string) *Metasploit {
    msf := &Metasploit{
        host: host,
        user: user,
        pass: pass,
    }

    return msf
}
```
Now you have a struct and, for convenience, a function named New() that initializes and returns a new struct.


### Performing Remote Calls
You can now build methods on your Metasploit type in order to perform the remote calls. To prevent extensive code 
duplication, in [rpc/msf.go](rpc/msf.go), you start by building a method that performs the serialization, deserialization, and 
HTTP communication logic. Then you won’t have to include this logic in every RPC function you build:
```go
func (msf *Metasploit) send(req interface{}, res interface{})❶ error {
    buf := new(bytes.Buffer)
 ❷ msgpack.NewEncoder(buf).Encode(req)
 ❸ dest := fmt.Sprintf("http://%s/api", msf.host)
    r, err := http.Post(dest, "binary/message-pack", buf)❹
    if err != nil {
        return err
    }
    defer r.Body.Close()

    if err := msgpack.NewDecoder(r.Body).Decode(&res)❺; err != nil {
        return err
    }

    return nil
}
```
The send() method receives request and response parameters of type interface{} ❶. Using this interface type allows 
you to pass any request struct into the method, and subsequently serialize and send the request to the server. 
Rather than explicitly returning the response, you’ll use the res interface{} parameter to populate its data by 
writing a decoded HTTP response to its location in memory.

Next, use the `msgpack` library to encode the request ❷. The logic to do this matches that of other standard, structured 
data types: first create an encoder via NewEncoder() and then call the Encode() method. This populates the buf 
variable with MessagePack-encoded representation of the request struct. Following the encoding, you build the 
destination URL by using the data within the Metasploit receiver, msf ❸. You use that URL and issue a POST request, 
explicitly setting the content type to `binary/message-pack` and setting the body to the serialized data ❹. Finally, 
you decode the response body ❺. As alluded to earlier, the decoded data is written to the memory location of the 
response interface that was passed into the method. The encoding and decoding of data is done without ever needing 
to explicitly know the request or response struct types, making this a flexible, reusable method.

In [rpc/msf.go](rpc/msf.go), you can see the meat of the logic in all its glory.
```go
func (msf *Metasploit) Login()❶ error {
    ctx := &loginReq{
        Method:   "auth.login",
        Username: msf.user,
        Password: msf.pass,
    }
    var res loginRes
    if err := msf.send(ctx, &res)❷; err != nil {
        return err
    }
    msf.token = res.Token
    return nil
}

func (msf *Metasploit) Logout()❸ error {
    ctx := &logoutReq{
        Method:      "auth.logout",
        Token:       msf.token,
        LogoutToken: msf.token,
    }
    var res logoutRes
    if err := msf.send(ctx, &res)❹; err != nil {
        return err
    }
    msf.token = ""
    return nil
}

func (msf *Metasploit) SessionList()❺ (map[uint32]SessionListRes, error) {
    req := &SessionListReq{Method: "session.list", Token: msf.token}
 ❻ res := make(map[uint32]SessionListRes)
    if err := msf.send(req, &res)❼; err != nil {
        return nil, err
    }

 ❽ for id, session := range res {
        session.ID = id
        res[id] = session
    }
    return res, nil
}
```
You define three methods: Login() ❶, Logout() ❸, and SessionList() ❺. Each method uses the same general flow: create 
and initialize a request struct, create the response struct, and call the helper function ❷❹❼ to send the request and 
receive the decoded response. The Login() and Logout() methods manipulate the token property. The only significant 
difference between method logic appears in the SessionList() method, where you define the response as a 
`map[uint32]SessionListRes` ❻ and loop over that response to flatten the map ❽, setting the ID property on the struct 
rather than maintaining a map of maps.

Remember that the `session.list()` RPC function requires a valid authentication token, meaning you have to log in 
before the `SessionList()` method call will succeed. Below function in [rpc/msf.go](rpc/msf.go) uses the Metasploit 
receiver struct to access a token, which isn’t a valid value yet—it’s an empty string. Since the code you’re 
developing here isn’t fully featured, you could just explicitly include a call to your `Login()` method from within the `SessionList()` 
method, but for each additional authenticated method you implement, you’d have to check for the existence of a valid 
authentication token and make an explicit call to `Login()`. This isn’t great coding practice because you’d spend a 
lot of time repeating logic that you could write, say, as part of a bootstrapping process. 

You’ve already implemented a function, `New()`, designed to be used for bootstrapping, so patch up that function to 
see what a new implementation looks like when including authentication as part of the process(see [`New()` function on rpc/msf.go](rpc/msf.go)):
```go
func New(host, user, pass string) (*Metasploit, error)❶ {
    msf := &Metasploit{
        host: host,
        user: user,
        pass: pass,
    }

    if err := msf.Login()❷; err != nil {
        return nil, err
    }

    return msf, nil
}
```
The patched-up code now includes an error as part of the return value set ❶. This is to alert on possible 
authentication failures. Also, added to the logic is an explicit call to the `Login()` method ❷. As long as 
the Metasploit struct is instantiated using this `New()` function, your authenticated method calls will now 
have access to a valid authentication token.

### Creating a Utility Program
Your last effort is to create the utility program that implements your shiny new library. Enter the below code snippet 
into [client/main.go](client/main.go), run it, and watch the magic happen.
```go
package main

import (
    "fmt"
    "log"

    "github.com/blackhat-go/bhg/ch-3/metasploit-minimal/rpc"
)

func main() {
    host := os.Getenv("MSFHOST")
    pass := os.Getenv("MSFPASS")
    user := "msf"

    if host == "" || pass == "" {
        log.Fatalln("Missing required environment variable MSFHOST or MSFPASS")
    }
    msf, err := rpc.New(host, user, pass)❶
    if err != nil {
        log.Panicln(err)
    }
 ❷ defer msf.Logout()

    sessions, err := msf.SessionList()❸
    if err != nil {
        log.Panicln(err)
    }
    fmt.Println("Sessions:")
 ❹ for _, session := range sessions {
        fmt.Printf("%5d  %s\n", session.ID, session.Info)
    }
}
```
First, bootstrap the RPC client and initialize a new Metasploit struct ❶. `Remember, you just updated this function 
to perform authentication during initialization.` Next, ensure you do proper cleanup by issuing a deferred call 
to the `Logout()` method ❷. This will run when the main function returns or exits. You then issue a call to the 
`SessionList()` method ❸ and iterate over that response to list out the available Meterpreter sessions ❹.

That was a lot of code, but fortunately, implementing other API calls should be substantially less work since 
you’ll just be defining request and response types and building the library method to issue the remote call.