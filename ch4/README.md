# HTTP Servers, Routing And Middleware

If you know how to write HTTP servers from scratch, you can create customized logic for social engineering, 
command-and-control (C2) transports, or APIs and frontends for your own tools, among other things. Luckily, Go has a 
brilliant standard package — `net/http` — for building HTTP servers; it’s really all you need to effectively write not 
only simple servers, but also complex, full-featured web applications.

In addition to the standard package, you can leverage third-party packages(such as `gorilla/mux`) to speed up development 
and remove some of the tedious processes, such as pattern matching. These packages will assist you with routing, building 
middleware, validating requests, and other tasks.

In this chapter, you’ll first explore many of the techniques needed to build HTTP servers using simple applications. Then 
you’ll deploy these techniques to create two social engineering applications:
  - a credential-harvesting server
  - keylogging server
  - multiplex C2 channels

### HTTP Server Basics

#### _Building a Simple Server_
The code in [hello_world/main.go](hello_world/main.go) starts a server that handles requests to a single path. The server 
should locate the name URL parameter containing a user’s name and respond with a customized greeting.

```go
package main

import (
    "fmt"
    "net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello %s\n", r.URL.Query().Get("name"))
}

func main() {
 ❶ http.HandleFunc("/hello", hello)
 ❷ http.ListenAndServe(":8000", nil)
}
```

This simple example exposes a resource at `/hello`. The resource grabs the parameter and echoes its value back to the 
client. Within the `main()` function, `http.HandleFunc()` ❶ takes two arguments: a string, which is a URL path pattern 
you’re instructing your server to look for, and a function, which will actually handle the request. You could provide 
the function definition as an anonymous inline function, if you want. In this example, you pass in the function named 
`hello()` that you defined earlier.

The `hello()` function handles requests and returns a hello message to the client. It takes two arguments itself. The 
first is `http.ResponseWriter`, which is used to write responses to the request. The second argument is a pointer to 
`http.Request`, which will allow you to read information from the incoming request. Note that you aren’t calling your 
`hello()` function from `main()`. You’re simply telling your HTTP server that any requests for `/hello` should be 
handled by a function named `hello()`.

Under the covers, what does `http.HandleFunc()` actually do? [The Go documentation](https://golang.org/pkg/net/http/#HandleFunc) 
will tell you that it places the handler on the `DefaultServerMux`. [A ServerMux](https://golang.org/pkg/net/http/#ServeMux) 
is short for a `server multiplexer`, which is just a fancy way to say that the underlying code can handle multiple 
HTTP requests for patterns and functions. `It does this using goroutines, with one goroutine per incoming request`. Importing 
the `net/http` package creates a `ServerMux` and attaches it to that package’s namespace; this is the `DefaultServerMux`.

The next line is a call to `http.ListenAndServe()` ❷, which takes a string and an `http.Handler` as arguments. This 
starts an HTTP server by using the first argument as the address. In this case, that’s :8000, which means the server 
should listen on port 8000 across all interfaces. For the second argument, the `http.Handler`, you pass in nil. As a 
result, the package uses `DefaultServerMux` as the underlying handler. Soon, you’ll be implementing your own `http.Handler` 
and will pass that in, but for now you’ll just use the default. You could also use `http.ListenAndServeTLS()`, which 
will start a server using HTTPS and TLS, as the name describes, but requires additional parameters.

Implementing the `http.Handler interface` requires a single method: `ServeHTTP(http.ResponseWriter, *http.Request)`. 
This is great because it simplifies the creation of your own custom HTTP servers. You’ll find numerous third-party 
implementations that extend the net/http functionality to add features such as middleware, authentication, response 
encoding, and more.

You can test this server by using curl:
```shell script
$ curl -i http://localhost:8000/hello?name=alice
HTTP/1.1 200 OK
Date: Sun, 12 Jan 2020 01:18:26 GMT
Content-Length: 12
Content-Type: text/plain; charset=utf-8

Hello alice
```

Excellent! The server you built reads the name URL parameter and replies with a greeting.

#### _Building a Simple Router_
Next you’ll build a simple router, shown in [simple_router/main.go](simple_router/main.go), that demonstrates how to 
dynamically handle inbound requests by inspecting the URL path. Depending on whether the URL contains the path /a, /b, 
or /c, you’ll print either the message Executing /a, Executing /b, or Executing /c. You’ll print a 404 Not Found error 
for everything else.
```go
   package main

   import (
       "fmt"
       "net/http"
   )

❶ type router struct {
   }

❷ func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    ❸ switch req.URL.Path {
       case "/a":
           fmt.Fprint(w, "Executing /a")
       case "/b":
           fmt.Fprint(w, "Executing /b")
       case "/c":
           fmt.Fprint(w, "Executing /c")
       default:
           http.Error(w, "404 Not Found", 404)
       }
   }

   func main() {
       var r router
    ❹ http.ListenAndServe(":8000", &r)
   }
```
First, you define a new type named router without any fields ❶. You’ll use this to implement the `http.Handler` 
interface. To do this, you must define the `ServeHTTP()` method ❷. The method uses a switch statement on the 
request’s URL path ❸, executing different logic depending on the path. It uses a default 404 Not Found response 
action. In `main()`, you create a new router and pass its respective pointer to `http.ListenAndServe()`❹.

Let’s take this for a spin in the ole terminal:
```shell script
$ curl http://localhost:8000/a
Executing /a
$ curl http://localhost:8000/d
404 Not Found
```

Everything works as expected; the program returns the message Executing /a for a URL that contains the /a path, and 
it returns a 404 response on a path that doesn’t exist. This is a trivial example. The third-party routers that you’ll 
use will have much more complex logic, but this should give you a basic idea of how they work.

#### _Building Simple Middleware_
Now let’s build middleware, which is a sort of wrapper that will execute on all incoming requests regardless of the 
destination function. In the example [simple_middleware/main.go](simple_middleware/main.go), you’ll create a logger 
that displays the request’s processing start and stop time.
```go
   Package main

   import (
           "fmt"
           "log"
           "net/http"
           "time"
   )

❶ type logger struct {
           Inner http.Handler
   }

❷ func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
           log.Println("start")
        ❸ l.Inner.ServeHTTP(w, r)
           log.Println("finish")
   }

   func hello(w http.ResponseWriter, r *http.Request) {
           fmt.Fprint(w, "Hello\n")
   }

   func main() {
        ❹ f := http.HandlerFunc(hello)
        ❺ l := logger{Inner: f}
        ❻ http.ListenAndServe(":8000", &l)
   }
```

What you’re essentially doing is creating an outer handler that, on every request, logs some information on the server 
and calls your hello() function. You wrap this logging logic around your function.

As with the routing example, you define a new type named logger, but this time you have a field, `Inner`, which is an 
`http.Handler` itself ❶. In your `ServeHTTP()` definition ❷, you use `log()` to print the start and finish times of 
the request, calling the `inner handler’s ServeHTTP() method` in between ❸. To the client, the request will finish 
inside the inner handler. Inside main(), you use `http.HandlerFunc()` to create an http.Handler out of a function ❹. 
You create the logger, setting Inner to your newly created handler ❺. Finally, you start the server by using a pointer 
to a logger instance ❻.

Running this and issuing a request outputs two messages containing the start and finish times of the request:
```shell script
$ go build -o simple_middleware
$ ./simple_middleware
2020/01/16 06:23:14 start
2020/01/16 06:23:14 finish
```

In the following sections, we’ll dig deeper into middleware and routing and use some of our favorite third-party 
packages, which let you create more dynamic routes and execute middleware inside a chain. We’ll also discuss some 
use cases for middleware that move into more complex scenarios.

#### _Routing with the gorilla/mux Package_
As shown in [simple_router/main.go](simple_router/main.go), you can use routing to match a request’s path to a 
function. But you can also use it to match other properties—such as the HTTP verb or host header—to a function. 
Several third-party routers are available in the Go ecosystem. Here, we’ll introduce you to one of them: the `gorilla/mux` 
package.

In the [gorilla_router/main.go](gorilla_router/main.go), we created a simple example with `gorilla/mux`.

The `gorilla/mux` package is a mature, third-party routing package that allows you to route based on both simple and 
complex patterns. It includes regular expressions, parameter matching, verb matching, and sub routing, among other 
features.

Before you can use gorilla/mux, you must go get it:
```shell script
$ go get github.com/gorilla/mux
```

Now, you can start routing. Create your router by using `mux.NewRouter()`:
```go
r := mux.NewRouter()
```

The returned type implements http.Handler but has a host of other associated methods as well. The one you’ll use 
most often is `HandleFunc()`. For example, if you wanted to define a new route to handle GET requests to the 
pattern `/foo`, you could use this:
```go
r.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprint(w, "hi foo")
}).Methods("GET")❶
```
Now, because of the call to `Methods()` ❶, only `GET` requests will match this route. All other methods will return 
a 404 response. You can chain other qualifiers on top of this, such as `Host(string)`, which matches a particular 
host header value. For example, the following will match only requests whose host header is set to www.foo.com:
```go
r.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprint(w, "hi foo")
}).Methods("GET").Host("www.foo.com")
```

Sometimes it’s helpful to match and pass in parameters within the request path (for example, when implementing a RESTful 
API). This is simple with gorilla/mux. The following will print out anything following /users/ in the request’s path:
```go
r.HandleFunc("/users/{user}", func(w http.ResponseWriter, req *http.Request) {
    user := mux.Vars(req)["user"]
    fmt.Fprintf(w, "hi %s\n", user)
}).Methods("GET")
```
In the path definition, you use braces to define a request parameter. Think of this as a named placeholder. Then, 
inside the handler function, you call `mux.Vars()`, passing it the request object, which returns a map[string]string—a 
map of request parameter names to their respective values. You provide the named placeholder user as the key. So, a 
request to `/users/bob` should produce a greeting for Bob:
```shell script
$ curl http://localhost:8000/users/bob
hi bob
```

You can take this a step further and use a regular expression to qualify the patterns passed. For example, you can 
specify that the user parameter must be lowercase letters:
```go
r.HandleFunc("/users/{user:[a-z]+}", func(w http.ResponseWriter, req *http.Request) {
    user := mux.Vars(req)["user"]
    fmt.Fprintf(w, "hi %s\n", user)
}).Methods("GET")
```

Any requests that don’t match this pattern will now return a 404 response:
```shell script
$ curl -i http://localhost:8000/users/bob1
HTTP/1.1 404 Not Found
```

#### _Building Middleware with Negroni_
The simple middleware we showed earlier logged the start and end times of the handling of the request and returned the 
response. Middleware doesn’t have to operate on every incoming request, but most of the time that will be the case. 
There are many reasons to use middleware, including logging requests, authenticating and authorizing users, and mapping 
resources.

For example, you could write middleware for performing basic authentication. It could parse an authorization header for 
each request, validate the username and password provided, and return a 401 response if the credentials are invalid. 
You could also chain multiple middleware functions together in such a way that after one is executed, the next one 
defined is run.

For the logging middleware you created earlier in this chapter, you wrapped only a single function. In practice, this 
is not very useful, because you’ll want to use more than one, and to do this, you must have logic that can execute them 
in a chain, one after another. Writing this from scratch is not incredibly difficult, but let’s not re-create the wheel. 
Here, you’ll use a mature package that is already able to do this: `urfave/negroni`.

The [negroni](https://github.com/urfave/negroni/) package, is great because it doesn’t tie you into a larger framework. 
You can easily bolt it onto other frameworks, and it provides a lot of flexibility. It also comes with default middleware 
that is useful for many applications. Before you hop in, you need to `go get negroni`:
```shell script
$ go get github.com/urfave/negroni
```

While you technically could use negroni for all application logic, doing this is far from ideal because it’s purpose-built 
to act as middleware and doesn’t include a router. Instead, it’s best to use negroni in combination with another package, 
such as `gorilla/mux` or `net/http`. Let’s use gorilla/mux to build a program that will get you acquainted with negroni and 
allow you to visualize the order of operations as they traverse the middleware chain.

Start by creating a new file called [gorilla_with_negroni/main.go](gorilla_with_negroni/main.go).
```go
package main

import (
    "net/http"

    "github.com/gorilla/mux"
    "github.com/urfave/negroni"
)

func main() {
 ❶ r := mux.NewRouter()
 ❷ n := negroni.Classic()
 ❸ n.UseHandler(r)
    http.ListenAndServe(":8000", n)
}
```

First, you create a router as you did earlier in this chapter by calling `mux.NewRouter()` ❶. Next comes your first 
interaction with the negroni package: you make a call to `negroni.Classic()` ❷. This creates a new pointer to a 
Negroni instance.

There are different ways to do this. You can either use `negroni.Classic()` or call `negroni.New()`. The first, 
`negroni.Classic()`, sets up default middleware, including a request logger, recovery middleware that will intercept 
and recover from panics, and middleware that will serve files from the public folder in the same directory. The 
`negroni.New()` function doesn’t create any default middleware.

Each type of middleware is available in the negroni package. For example, you can use the recovery package by doing 
the following:
```go
n.Use(negroni.NewRecovery())
```

Next, you add your router to the middleware stack by calling `n.UseHandler(r)` ❸. As you continue to plan and build 
out your middleware, consider the order of execution. For example, you’ll want your authentication-checking middleware 
to run prior to the handler functions that require authentication. Any middleware mounted before the router will 
execute prior to your handler functions; any middleware mounted after the router will execute after your handler 
functions. Order matters. In this case, you haven’t defined any custom middleware, but you will soon.

Go ahead and build the server you created in Listing 4-4, and then execute it. Then issue web requests to the server 
at http://localhost:8000. You should see the negroni logging middleware print information to stdout, as shown next. 
The output shows the timestamp, response code, processing time, host, and HTTP method:
```shell script
$ go build -s negroni_example
$ ./negroni_example
 [negroni] 2020-01-19T11:49:33-07:00 | 404 |      1.0002ms | localhost:8000 | GET
```

Having default middleware is great and all, but the real power comes when you create your own. With negroni, you can 
use a few methods to add middleware to the stack. Take a look at the following code. It creates trivial middleware 
that prints a message and passes execution to the next middleware in the chain:
```go
type trivial struct {
}
func (t *trivial) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) { ❶
    fmt.Println("Executing trivial middleware")
    next(w, r) ❷
}
```

This implementation is slightly different from previous examples. Before, you were implementing the `http.Handler` interface, 
which expected a `ServeHTTP()` method that accepted two parameters: http.ResponseWriter and *http.Request. In this new 
example, instead of the http.Handler interface, you’re implementing the `negroni.Handler` interface.

The slight difference is that the `negroni.Handler` interface expects you to implement a `ServeHTTP()` method that 
accepts not two, but three, parameters: `http.ResponseWriter`, `*http.Request`, and `http.HandlerFunc` ❶. The `http.HandlerFunc` 
parameter represents the next middleware function in the chain. For your purposes, you name it next. You do your 
processing within `ServeHTTP()`, and then call `next()` ❷, passing it the `http.ResponseWriter` and `*http.Request` 
values you originally received. This effectively transfers execution down the chain.

But you still have to tell negroni to use your implementation as part of the middleware chain. You can do this by 
calling negroni’s Use method and passing an instance of your negroni.Handler implementation to it:
```go
n.Use(&trivial{})
```

Writing your middleware by using this method is convenient because you can easily pass execution to the next middleware. 
There is one drawback: anything you write must use negroni. For example, if you were writing a middleware package 
that writes security headers to a response, you would want it to implement http.Handler, so you could use it in other 
application stacks, since most stacks won’t expect a negroni.Handler. The point is, regardless of your middleware’s 
purpose, compatibility issues may arise when trying to use negroni middleware in a non-negroni stack, and vice versa.

There are two other ways to tell negroni to use your middleware. `UseHandler(handler http.Handler)`, which you’re 
already familiar with, is the first. The second way is to call `UseHandleFunc(handlerFunc func(w http.ResponseWriter, r *http.Request))`. 
The latter is not something you’ll want to use often, since it doesn’t let you forgo execution of the next middleware 
in the chain. For example, if you were writing middleware to perform authentication, you would want to return a 401 
response and stop execution if any credentials or session information were invalid; with this method, there’s no way 
to do that.

#### _Adding Authentication with Negroni_
Before moving on, let’s modify our example from the previous section to demonstrate the use of `context`, which can 
easily pass variables between functions. The example in [gorilla_with_negroni/main.go](gorilla_with_negroni/main.go) uses 
negroni to add authentication middleware.
```go
import (
    "context"
    "fmt"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/urfave/negroni"
)

type badAuth struct { ❶
    Username string
    Password string
}

func (b *badAuth) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) { ❷
    username := r.URL.Query().Get("username") ❸
    password := r.URL.Query().Get("password")
    if username != b.Username || password != b.Password {
        http.Error(w, "Unauthorized", 401)
        return ❹
    }
    ctx := context.WithValue(r.Context(), "username", username) ❺
    r = r.WithContext(ctx) ❻
    next(w, r)
}

func hello(w http.ResponseWriter, r *http.Request) {
    username := r.Context().Value("username").(string) ❼
    fmt.Fprintf(w, "Hi %s\n", username)
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/hello", hello).Methods("GET")
    n := negroni.Classic()
    n.Use(&badAuth{
        Username: "admin",
        Password: "password",
    })
    n.UseHandler(r)
    http.ListenAndServe(":8000", n)
}
```
You’ve added new middleware, `badAuth`, that is going to simulate authentication, purely for demonstration purposes ❶. 
This new type has two fields, `Username` and `Password`, and implements `negroni.Handler`, since it defines the three-parameter 
version of the `ServeHTTP()` method ❷ we discussed previously. Inside the `ServeHTTP()` method, you first grab the 
username and password from the request ❸, and then compare them to the fields you have. If the username and password are 
incorrect, execution is stopped, and a 401 response is written to the requester.

Notice that you return ❹ before calling `next()`. This prevents the remainder of the middleware chain from executing. 
If the credentials are correct, you go through a rather verbose routine of adding the username to the request context. 
You first call `context.WithValue()` to initialize the context from the request, setting a variable named username on 
that context ❺. You then make sure the request uses your new context by calling `r.WithContext(ctx)` ❻. If you plan 
on writing web applications with Go, you’ll want to become familiar with this pattern, as you’ll be using it a lot.

In the `hello()` function, you get the username from the request context by using the `Context().Value(interface{})` function, 
which itself returns an interface{}. Because you know it’s a string, you can use a type assertion here ❼. If you can’t 
guarantee the type, or you can’t guarantee that the value will exist in the context, use a switch routine for conversion.

Build and execute the code from above snippet and send a few requests to the server. Send some with both correct and 
incorrect credentials. You should see the following output:
```shell script
$ curl -i http://localhost:8000/hello
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Thu, 16 Jan 2020 20:41:20 GMT
Content-Length: 13
Unauthorized
$ curl -i 'http://localhost:8000/hello?username=admin&password=password'
HTTP/1.1 200 OK
Date: Thu, 16 Jan 2020 20:41:05 GMT
Content-Length: 9
Content-Type: text/plain; charset=utf-8

Hi admin
```

Making a request without credentials results in your middleware returning a 401 Unauthorized error. Sending the same 
request with a valid set of credentials produces a super-secret greeting message accessible only to authenticated users.

That was an awful lot to digest. Up to this point, your handler functions have solely used `fmt.FPrintf()` to write your 
response to the `http.ResponseWriter` instance. In the next section, you’ll look at a more dynamic way of returning HTML 
by using Go’s templating package.

#### _Using Templates to Produce HTML Responses_
Templates allow you to dynamically generate content, including HTML, with variables from Go programs. Many languages have 
third-party packages that allow you to generate templates. Go has two templating packages, `text/template` and `html/template`. 
In this chapter, you’ll use the HTML package, because it provides the contextual encoding you need.

One of the fantastic things about Go’s package is that it’s contextually aware: it will encode your variable differently 
depending on where the variable is placed in the template. For example, if you were to supply a string as a URL to an href 
attribute, the string would be URL encoded, but the same string would be HTML encoded if it rendered within an HTML element.

To create and use templates, you first define your template, which contains a placeholder to denote the dynamic contextual 
data to render. Its syntax should look familiar to readers who have used `Jinja with Python`. When you render the template, 
you pass to it a variable that’ll be used as this context. The variable can be a complex structure with several fields, or 
it can be a primitive variable.

Let’s work through a sample, shown in [template_example/main.go](template_example/main.go), that creates a simple template 
and populates a placeholder with JavaScript. This is a contrived example that shows how to dynamically populate content 
returned to the browser.
```go
   package main

   import (
       "html/template"
       "os"
   )

❶ var x = `
   <html>
     <body>

    ❷ Hello {{.}}
     </body>
   </html>
   `

   func main() {
    ❸ t, err := template.New("hello").Parse(x)
       if err != nil {
           panic(err)
       }
    ❹ t.Execute(os.Stdout, "<script>alert('world')</script>")
   }
```

The first thing you do is create a variable, named x, to store your HTML template ❶. Here you’re using a string embedded 
in your code to define your template, but most of the time you’ll want to store your templates as separate files. Notice 
that the template is nothing more than a simple HTML page. Inside the template, you define placeholders by using the 
`{{variable-name}}` convention, where variable-name is the data element within your contextual data that you’ll want 
to render ❷. Recall that this can be a struct or another primitive. In this case, you’re using a single period, which 
tells the package that you want to render the entire context here. Since you’ll be working with a single string, this 
is fine, but if you had a larger and more complex data structure, such as a struct, you could get only the fields you 
want by calling past this period. For example, if you passed a struct with a Username field to the template, you could 
render the field by using `{{.Username}}`.

Next, in your `main()` function, you create a new template by calling `template.New(string)` ❸. Then you call `Parse(string)` 
to ensure that the template is properly formatted and to parse it. Together, these two functions return a new pointer 
to a Template.

While this example uses only a single template, it’s possible to embed templates in other templates. When using multiple 
templates, it’s important that you name them in order to be able to call them. Finally, you call `Execute(io.Writer, interface{})` ❹, 
which processes the template by using the variable passed as the second argument and writes it to the provided `io.Writer`. 
For demonstration purposes, you’ll use os.Stdout. The second variable you pass into the Execute() method is the context 
that’ll be used for rendering the template.

Running this produces HTML, and you should notice that the script tags and other nefarious characters that were provided 
as part of your context are properly encoded. Neat-o!
```shell script
$ go build -o template_example
$ ./template_example

<html>
  <body>
    Hello &lt;script&gt;alert(&#39;world&#39;)&lt;/script&gt;
  </body>
</html>
```

We could say a lot more about templates. You can use logical operators with them; you can use them with loops and other 
control structures. You can call built-in functions, and you can even define and expose arbitrary helper functions to 
greatly expand the templating capabilities. Double neat-o! We recommend you dive in and research these possibilities. 
They’re beyond the scope of this book, but are powerful.

How about you step away from the basics of creating servers and handling requests and instead focus on something more 
nefarious. Let’s create a credential harvester!

For the coding exercises with tools, check the subfolders in current directory with the following order:
  - [CREDENTIAL HARVESTING](credential-harvester/README.md)
  - [KEYLOGGING WITH THE WEBSOCKET API](websocket_keylogger/README.md)