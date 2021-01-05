## Subdomain Guesser
When you first start writing a new tool, you must decide which arguments the program will take. This `subdomain-guessing` 
program will take several arguments, including the target domain, the filename containing subdomains to guess, the destination 
DNS server to use, and the number of workers to launch. Go provides a useful package for parsing command line options called 
`flag` that you’ll use to handle your command line arguments. Although we don’t use the flag package across all of our 
code examples, we’ve opted to use it in this case to demonstrate more robust, elegant argument parsing. [main.go](main.go) shows 
our argument-parsing code.
```go
package main

import (
    "flag"
)

func main() {
    var (
        flDomain      = flag.String("domain", "", "The domain to perform guessing against.") ❶
        flWordlist    = flag.String("wordlist", "", "The wordlist to use for guessing.")
        flWorkerCount = flag.Int("c", 100, "The amount of workers to use.") ❷
        flServerAddr  = flag.String("server", "8.8.8.8:53", "The DNS server to use.")
    )
    flag.Parse() ❸
}
```
First, the code line declaring the `flDomain` variable ❶ takes a String argument and declares an empty string default value 
for what will be parsed as the domain option. The next pertinent line of code is the `flWorkerCount` variable declaration ❷. 
You need to provide an Integer value as the c command line option. In this case, set this to 100 default workers. But this 
value is probably too conservative, so feel free to increase the number when testing. Finally, a call to `flag.Parse()` ❸ 
populates your variables by using the provided input from the user.

If you try to build this program, you should receive an error about unused variables. Add the following code immediately 
after your call to `flag.Parse()`. This addition prints the variables to stdout along with code, ensuring that the user 
provided -domain and -wordlist:
```go
if *flDomain == "" || *flWordlist == "" {
    fmt.Println("-domain and -wordlist are required")
    os.Exit(1)
}
fmt.Println(*flWorkerCount, *flServerAddr)
```

To allow your tool to report which names were resolvable along with their respective IP addresses, you’ll create a struct 
type to store this information. Define it above the main() function:
```go
type result struct {
    IPAddress string
    Hostname string
}
```

You’ll query two main record types—A and CNAME—for this tool. You’ll perform each query in a separate function. It’s a 
good idea to keep your functions as small as possible and to have each perform one thing well. This style of development 
allows you to write smaller tests in the future.

### Querying A and CNAME Records
You’ll create two functions to perform queries: one for `A records` and the other for `CNAME records`. Both functions 
accept a FQDN as the first argument and the DNS server address as the second. Each should return a slice of strings and 
an error. Add these functions to the code you began defining in [main.go](main.go). These functions should be defined 
outside `main()`.
```go
func lookupA(fqdn, serverAddr string) ([]string, error) {
    var m dns.Msg
    var ips []string
    m.SetQuestion(dns.Fqdn(fqdn), dns.TypeA)
    in, err := dns.Exchange(&m, serverAddr)
    if err != nil {
        return ips, err
    }
    if len(in.Answer) < 1 {
        return ips, errors.New("no answer")
    }
    for _, answer := range in.Answer {
        if a, ok := answer.(*dns.A); ok {
            ips = append(ips, a.A.String())
        }
    }
    return ips, nil
}

func lookupCNAME(fqdn, serverAddr string) ([]string, error) {
    var m dns.Msg
    var fqdns []string
    m.SetQuestion(dns.Fqdn(fqdn), dns.TypeCNAME)
    in, err := dns.Exchange(&m, serverAddr)
    if err != nil {
        return fqdns, err
    }
    if len(in.Answer) < 1 {
        return fqdns, errors.New("no answer")
    }
    for _, answer := range in.Answer {
        if c, ok := answer.(*dns.CNAME); ok {
            fqdns = append(fqdns, c.Target)
        }
    }
    return fqdns, nil
}
```

This code should look familiar because it’s nearly identical to the code you wrote in the first section of this chapter. 
The first function, `lookupA`, returns a list of IP addresses, and `lookupCNAME` returns a list of hostnames.

`CNAME`, or `canonical name`, records point one FQDN to another one that serves as an alias for the first. For instance, 
say the owner of the `example.com` organization wants to host a WordPress site by using a WordPress hosting service. That 
service may have hundreds of IP addresses for balancing all of their users’ sites, so providing an individual site’s IP 
address would be infeasible. The WordPress hosting service can instead provide a canonical name (a CNAME) that the owner 
of example.com can reference. So `www.example.com` might have a CNAME pointing to `someserver.hostingcompany.org`, which 
in turn has an A record pointing to an IP address. This allows the owner of example.com to host their site on a server 
for which they have no IP information.

Often this means you’ll need to follow the trail of CNAMES to eventually end up at a valid A record. We say trail because 
you can have an endless chain of CNAMES. Place the function in the following code outside main() to see how you can use 
the trail of CNAMES to track down the valid A record:
```go
func lookup(fqdn, serverAddr string) []result {
 ❶ var results []result
 ❷ var cfqdn = fqdn // Don't modify the original.
    for {
     ❸ cnames, err := lookupCNAME(cfqdn, serverAddr)
     ❹ if err == nil && len(cnames) > 0 {
         ❺ cfqdn = cnames[0]
         ❻ continue // We have to process the next CNAME.
        }
     ❼ ips, err := lookupA(cfqdn, serverAddr)
        if err != nil {
            break // There are no A records for this hostname.
        }
     ❽ for _, ip := range ips {
            results = append(results, result{IPAddress: ip, Hostname: fqdn})
        }
     ❾ break // We have processed all the results.
    }
    return results
}
```

First, define a slice to store results ❶. Next, create a copy of the FQDN passed in as the first argument ❷, not only so 
you don’t lose the original FQDN that was guessed, but also so you can use it on the first query attempt. After starting 
an infinite loop, try to resolve the CNAMEs for the FQDN ❸. If no errors occur and at least one CNAME is returned ❹, set 
cfqdn to the CNAME returned ❺, using continue to return to the beginning of the loop ❻. This process allows you to follow 
the trail of CNAMES until a failure occurs. If there’s a failure, which indicates that you’ve reached the end of the chain, 
you can then look for A records ❼; but if there’s an error, which indicates something went wrong with the record lookup, 
then you leave the loop early. If there are valid A records, append each of the IP addresses returned to your results 
slice ❽ and break out of the loop ❾. Finally, return the results to the caller.

 Our logic associated with the name resolution seems sound. However, you haven’t accounted for performance. Let’s make 
 our example goroutine-friendly so you can add concurrency.
 
 ### Passing to a Worker Function
 You’ll create a pool of goroutines that pass work to a worker function, which performs a unit of work. You’ll do this by 
 using channels to coordinate work distribution and the gathering of results. Recall that you did something similar in 
 [Chapter 2](../../ch2), when you built a concurrent port scanner.
 
 Continue to expand the code from [main.go](main.go). First, create the `worker()` function and place it outside `main()`. 
 This function takes three channel arguments: a channel for the worker to signal whether it has closed, a channel of 
 domains on which to receive work, and a channel on which to send results. The function will need a final string argument 
 to specify the DNS server to use. The following code shows an example of our `worker()` function:
 ```go
type empty struct{} ❶

func worker(tracker chan empty, fqdns chan string, gather chan []result, serverAddr string) {
    for fqdn := range fqdns { ❷
        results := lookup(fqdn, serverAddr)
        if len(results) > 0 {
            gather <- results ❸
        }
    }
    var e empty
    tracker <- e ❹
}
```

Before introducing the `worker()` function, first define the type empty to track when the worker finishes ❶. This is a 
struct with no fields; you use an empty struct because it’s 0 bytes in size and will have little impact or overhead when 
used. Then, in the worker() function, loop over the domains channel ❷, which is used to pass in FQDNs. After getting 
results from your lookup() function and checking to ensure there is at least one result, send the results on the gather 
channel ❸, which accumulates the results back in main(). After the work loop exits because the channel has been closed, 
an empty struct is sent on the tracker channel ❹ to signal the caller that all work has been completed. Sending the empty 
struct on the tracker channel is an important last step. If you don’t do this, you’ll have a race condition, because the 
caller may exit before the gather channel receives results.

Since all of the prerequisite structure is set up at this point, let’s refocus our attention back to `main()` to complete 
the program we began in [main.go](main.go). Define some variables that will hold the results and the channels that will 
be passed to `worker()`. Then append the following code into main():
```go
var results []result
fqdns := make(chan string, *flWorkerCount)
gather := make(chan []result)
tracker := make(chan empty)
```

Create the fqdns channel as a buffered channel by using the number of workers provided by the user. This allows the workers 
to start slightly faster, as the channel can hold more than a single message before blocking the sender.

 ### Creating a Scanner with bufio
 Next, open the file provided by the user to consume as a word list. With the file open, create a new scanner by using the 
 [bufio package](https://godoc.org/bufio). The scanner allows you to read the file one line at a time. Append the following 
 code into `main()`:
 ```go
fh, err := os.Open(*flWordlist)
if err != nil {
    panic(err)
}
defer fh.Close()
scanner := bufio.NewScanner(fh)
```

The built-in function `panic()` is used here if the error returned is not nil. When you’re writing a package or program 
that others will use, you should consider presenting this information in a cleaner format.

You’ll use the new scanner to grab a line of text from the supplied word list and create a FQDN by combining the text with 
the domain the user provides. You’ll send the result on the fqdns channel. But you must start the workers first. The order 
of this is important. If you were to send your work down the fqdns channel without starting the workers, the buffered 
channel would eventually become full, and your producers would block. You’ll add the following code to your main() 
function. Its purpose is to start the worker goroutines, read your input file, and send work on your fqdns channel.
```go
❶ for i := 0; i < *flWorkerCount; i++ {
       go worker(tracker, fqdns, gather, *flServerAddr)
   }

❷ for scanner.Scan() {
       fqdns <- fmt.Sprintf("%s.%s", scanner.Text()❸, *flDomain)
   }
```
Creating the workers ❶ by using this pattern should look similar to what you did when building your concurrent port scanner: 
you used a for loop until you reached the number provided by the user. To grab each line in the file, `scanner.Scan()` is 
used in a loop ❷. This loop ends when there are no more lines to read in the file. To get a string representation of the 
text from the scanned line, use `scanner.Text()` ❸.

The work has been launched! Take a second to bask in greatness. Before reading the next code, think about where you are in 
the program and what you’ve already done in this book. Try to complete this program and then continue to the next section, 
where we’ll walk you through the rest.

 ### Gathering and Displaying the Results
 To finish up, first start an anonymous goroutine that will gather the results from the workers. Append the following code 
 into `main()`:
 ```go
go func() {
    for r := range gather {
     ❶ results = append(results, r...❷)
    }
    var e empty
 ❸ tracker <- e
}()
```

By looping over the gather channel, you append the received results onto the results slice ❶. Since you’re appending a 
slice to another slice, you must use the ... syntax ❷. After you close the gather channel and the loop ends, send an 
empty struct to the tracker channel as you did earlier ❸. This is done to prevent a race condition in case append() 
doesn’t finish by the time you eventually present the results to the user.

All that’s left is closing the channels and presenting the results. Include the following code at the bottom of main() 
in order to close the channels and present the results to the user:
```go
❶ close(fqdns)
❷ for i := 0; i < *flWorkerCount; i++ {
       <-tracker
   }
❸ close(gather)
❹ <-tracker
```
The first channel that can be closed is fqdns ❶ because you’ve already sent all the work on this channel. Next, you need 
to receive on the tracker channel one time for each of the workers ❷, allowing the workers to signal that they exited 
completely. With all of the workers accounted for, you can close the gather channel ❸ because there are no more results 
to receive. Finally, receive one more time on the tracker channel to allow the gathering goroutine to finish completely ❹.

The results aren’t yet presented to the user. Let’s fix that. If you wanted to, you could easily loop over the results 
slice and print the Hostname and IPAddress fields by using fmt.Printf(). We prefer, instead, to use one of Go’s several 
great built-in packages for presenting data; `tabwriter` is one of our favorites. It allows you to present data in nice, 
even columns broken up by tabs. Add the following code to the end of main() to use tabwriter to print your results:
```go
w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', 0)
for _, r := range results {
    fmt.Fprintf(w, "%s\t%s\n", r.Hostname, r.IPAddress)
}
w.Flush()
```

Below shows the program in its entirety:
```go
Package main

import (
    "bufio"
    "errors"
    "flag"
    "fmt"
    "os"
    "text/tabwriter"

    "github.com/miekg/dns"
)

func lookupA(fqdn, serverAddr string) ([]string, error) {
    var m dns.Msg
    var ips []string
    m.SetQuestion(dns.Fqdn(fqdn), dns.TypeA)
    in, err := dns.Exchange(&m, serverAddr)
    if err != nil {
        return ips, err
    }
    if len(in.Answer) < 1 {
        return ips, errors.New("no answer")
    }
    for _, answer := range in.Answer {
        if a, ok := answer.(*dns.A); ok {
            ips = append(ips, a.A.String())
        }
    }
    return ips, nil
}

func lookupCNAME(fqdn, serverAddr string) ([]string, error) {
    var m dns.Msg
    var fqdns []string
    m.SetQuestion(dns.Fqdn(fqdn), dns.TypeCNAME)
    in, err := dns.Exchange(&m, serverAddr)
    if err != nil {
        return fqdns, err
    }
    if len(in.Answer) < 1 {
        return fqdns, errors.New("no answer")
    }
    for _, answer := range in.Answer {
        if c, ok := answer.(*dns.CNAME); ok {
            fqdns = append(fqdns, c.Target)
        }
    }
    return fqdns, nil
}

func lookup(fqdn, serverAddr string) []result {
    var results []result
    var cfqdn = fqdn // Don't modify the original.
    For {
        cnames, err := lookupCNAME(cfqdn, serverAddr)
        if err == nil && len(cnames) > 0 {
            cfqdn = cnames[0]
            continue // We have to process the next CNAME.
        }
        ips, err := lookupA(cfqdn, serverAddr)
        if err != nil {
            break // There are no A records for this hostname.
        }
        for _, ip := range ips {
            results = append(results, result{IPAddress: ip, Hostname: fqdn})
        }
        break // We have processed all the results.
    }
    return results
}

func worker(tracker chan empty, fqdns chan string, gather chan []result, serverAddr string) {
    for fqdn := range fqdns {
        results := lookup(fqdn, serverAddr)
        if len(results) > 0 {
            gather <- results
        }
    }
    var e empty
    tracker <- e
}

type empty struct{}

type result struct {
    IPAddress string
    Hostname string
}

func main() {
    var (
        flDomain      = flag.String("domain", "", "The domain to perform guessing against.")
        flWordlist    = flag.String("wordlist", "", "The wordlist to use for guessing.")
        flWorkerCount = flag.Int("c", 100, "The amount of workers to use.")
        flServerAddr  = flag.String("server", "8.8.8.8:53", "The DNS server to use.")
    )
    flag.Parse()

    if *flDomain == "" || *flWordlist == "" {
        fmt.Println("-domain and -wordlist are required")
        os.Exit(1)
    }

    var results []result

    fqdns := make(chan string, *flWorkerCount)
    gather := make(chan []result)
    tracker := make(chan empty)

    fh, err := os.Open(*flWordlist)
    if err != nil {
        panic(err)
    }
    defer fh.Close()
    scanner := bufio.NewScanner(fh)

    for I := 0; i < *flWorkerCount; i++ {
        go worker(tracker, fqdns, gather, *flServerAddr)
    }

    go func() {
        for r := range gather {
            results = append(results, I.)
        }
        var e empty
        tracker <- e
    }()

    for scanner.Scan() {
        fqdns <- fmt.Sprintf"%s.%s", scanner.Text(), *flDomain)
    }
    // Note: We could check scanner.Err() here.

    close(fqdns)
    for i := 0; i < *flWorkerCount; i++ {
        <-tracker
    }
    close(gather)
    <-tracker

    w := tabwriter.NewWriter(os.Stdout, 0, 8' ', ' ', 0)
    for _, r := range results {
        fmt.Fprint"(w, "%s\"%s\n", r.Hostname, r.IPAddress)
    }
    w.Flush()
}
```

Your subdomain-guessing program is complete! You should now be able to build and execute your shiny new subdomain-guessing 
tool. Try it with word lists or dictionary files in open source repositories (you can find plenty with a Google search). 
Play around with the number of workers; you may find that if you go too fast, you’ll get varying results. Here’s a run 
from the authors’ system using 1000 workers:

```shell script
$ wc -l namelist.txt
1909 namelist.txt
$ time ./subdomain_guesser -domain microsoft.com -wordlist namelist.txt -c 1000
ajax.microsoft.com            72.21.81.200
buy.microsoft.com             157.56.65.82
news.microsoft.com            192.230.67.121
applications.microsoft.com    168.62.185.179
sc.microsoft.com              157.55.99.181
open.microsoft.com            23.99.65.65
ra.microsoft.com              131.107.98.31
ris.microsoft.com             213.199.139.250
smtp.microsoft.com            205.248.106.64
wallet.microsoft.com          40.86.87.229
jp.microsoft.com              134.170.185.46
ftp.microsoft.com             134.170.188.232
develop.microsoft.com         104.43.195.251
./subdomain_guesser -domain microsoft.com -wordlist namelist.txt -c 1000 0.23s user 0.67s system 22% cpu 4.040 total
```

You’ll see that the output shows several FQDNs and their IP addresses. We were able to guess the subdomain values for 
each result based off the word list provided as an input file.

Now that you’ve built your own subdomain-guessing tool and learned how to resolve hostnames and IP addresses to enumerate 
different DNS records, you’re ready to write your own DNS server and proxy.