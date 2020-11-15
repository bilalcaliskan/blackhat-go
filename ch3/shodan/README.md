### What is Shodan?
Shodan is a search engine that lets the user find specific types of 
computers (webcams, routers, servers, etc.) connected to the internet 
using a variety of filters. Some have also described it as a search 
engine of service banners, which are metadata that the server sends 
back to the client. This can be information about the server software, 
what options the service supports, a welcome message or anything else 
that the client can find out before interacting with the server.

Shodan collects data mostly on web servers (HTTP/HTTPS – ports 80, 8080, 443, 8443), 
as well as FTP (port 21), SSH (port 22), Telnet (port 23), SNMP (port 161), 
IMAP (ports 143, or (encrypted) 993), SMTP (port 25), SIP (port 5060), and Real Time 
Streaming Protocol (RTSP, port 554). The latter can be used to access webcams and 
their video stream.

[Shodan](https://www.shodan.io/), self-described as “the world’s first search 
engine for internet-connected devices,” facilitates passive reconnaissance by 
maintaining a searchable database of networked devices and services, including 
metadata such as product names, versions, locale, and more. Think of Shodan as a 
repository of scan data, even if it does much, much more.

### Building an HTTP client that interacts with Shodan
Prior to performing any authorized adversarial activities against an organization, any good attacker begins with
reconnaissance. Typically, this starts with passive techniques that don’t send packets to the target; that way,
detection of the activity is next to impossible. Attackers use a variety of sources and services—including social
networks, public records, and search engines—to gain potentially useful information about the target.

It’s absolutely incredible how seemingly benign information becomes critical when environmental context is applied
during a chained attack scenario. For example, a web application that discloses verbose error messages may, alone,
be considered low severity. However, if the error messages disclose the enterprise username format, and if the
organization uses single-factor authentication for its VPN, those error messages could increase the likelihood of
an internal network compromise through password-guessing attacks.

Maintaining a low profile while gathering the information ensures that the target’s awareness and security posture
remains neutral, increasing the likelihood that your attack will be successful.


### Reviewing the steps for building an API client
First, you’ll need a Shodan API key, which you get after you register on Shodan’s website. 
Now, get your API key from the site and set it as an environment variable. The following 
examples will work as-is only if you save your API key as the variable SHODAN_API_KEY.

The client you build will implement two API calls: one to query subscription credit 
information and the other to search for hosts that contain a certain string. You use 
the latter call for identifying hosts; for example, ports or operating systems matching 
a certain product.

Luckily, the Shodan API is straightforward, producing nicely structured JSON responses. 
This makes it a good starting point for learning API interaction. Here is a high-level 
overview of the typical steps for preparing and building an API client:
  - Review the service’s API documentation.
  - Design a logical structure for the code in order to reduce complexity and repetition.
  - Define request or response types, as necessary, in Go.
  - Create helper functions and types to facilitate simple initialization, authentication, and communication to
  reduce verbose or repetitive logic.
  - Build the client that interacts with the API consumer functions and types.


### Designing the project structure
When building an API client, you should structure it so that the function calls and 
logic standalone. This allows you to reuse the implementation as a library in other projects. 
That way, you won’t have to reinvent the wheel in the future. Building for reusability slightly 
changes a project’s structure. For the Shodan example, here’s the project structure:

  ```shell script
  $ tree blackhat-go/ch-3/shodan
  blackhat-go/ch-3/shodan
      |---cmd
      |   |---shodan
      |       |---main.go
      |---shodan
          |---api.go
          |---host.go
          |---shodan.go
  ```

The main.go file defines package main and is used primarily as a consumer of the API you’ll build; 
in this case, you use it primarily to interact with your client implementation.

The files in the shodan directory; `api.go`, `host.go`, and `shodan.go` define package shodan, 
which contains the types and functions necessary for communication to and from Shodan. This 
package will become your standalone library that you can import into various projects.


### Cleaning Up API Calls
When you perused the Shodan API documentation, you may have noticed that every exposed function 
requires you to send your API key. Although you certainly can pass that value around to each 
consumer function you create, that repetitive task becomes tedious. The same can be said for 
either hardcoding or handling the base URL (https://api.shodan.io/). For example, defining 
your API functions, as in the following snippet, requires you to pass in the token and URL 
to each function, which isn’t very elegant:
```go
func APIInfo(token, url string) { --snip-- }
func HostSearch(token, url string) { --snip-- }
```
Instead, opt for a more idiomatic solution that allows you to save keystrokes while arguably making your code more readable. 
To do this, create a [shodan.go](shodan/shodan.go) file and enter the code like below snippet:
```go
   package shodan

❶ const BaseURL = "https://api.shodan.io"

❷ type Client struct {
       apiKey string
   }

❸ func New(apiKey string) *Client {
       return &Client{apiKey: apiKey}
   }
```
The Shodan URL is defined as a constant value ❶; that way, you can easily access and reuse it 
within your implementing functions. If Shodan ever changes the URL of its API, you’ll have to 
make the change at only this one location in order to correct your entire codebase. Next, you 
define a Client struct, used for maintaining your API token across requests ❷. Finally, the 
code defines a `New()` helper function, taking the API token as input and creating and returning 
an initialized Client instance ❸. Now, rather than creating your API code as arbitrary functions, 
you create them as methods on the Client struct, which allows you to interrogate the instance 
directly rather than relying on overly verbose function parameters. 

### Querying Your Shodan Subscription
Now you’ll start the interaction with Shodan. Per the Shodan API documentation, the call to query your subscription 
plan information is as follows:
```
curl https://api.shodan.io/api-info?key={YOUR_API_KEY}
```
The response returned resembles the following structure. Obviously, the values will differ based on your plan details 
and remaining subscription credits.
```
{
 "query_credits": 56,
 "scan_credits": 0,
 "telnet": true,
 "plan": "edu",
 "https": true,
 "unlocked": true,
}
```
First, in [api.go](shodan/api.go), you’ll need to define a type that can be used to unmarshal the JSON response 
to a Go struct. Without it, you won’t be able to process or interrogate the response body. In 
this example, name the type APIInfo:
```
type APIInfo struct {
    QueryCredits int    `json:"query_credits"`
    ScanCredits  int    `json:"scan_credits"`
    Telnet       bool   `json:"telnet"`
    Plan         string `json:"plan"`
    HTTPS        bool   `json:"https"`
    Unlocked     bool   `json:"unlocked"`
}
```
The awesomeness that is Go makes that structure and JSON alignment a joy. For each exported 
type on the struct, you explicitly define the JSON element name with struct tags so you can 
ensure that data is mapped and parsed properly.

Next you need to implement the function in [api.go](shodan/api.go), which makes an HTTP GET 
request to Shodan and decodes the response into your APIInfo struct:
```
func (s *Client) APIInfo() (*APIInfo, error) {
    res, err := http.Get(fmt.Sprintf("%s/api-info?key=%s", BaseURL, s.apiKey))❶
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    var ret APIInfo
    if err := json.NewDecoder(res.Body).Decode(&ret)❷; err != nil {
        return nil, err
    }
    return &ret, nil
}
```
The implementation is short and sweet. You first issue an HTTP GET request to the /api-info 
resource ❶. The full URL is built using the BaseURL global constant and s.apiKey. You then 
decode the response into your APIInfo struct ❷ and return it to the caller.

Before writing code that utilizes this shiny new logic, build out a second, more useful API 
call—the host search—which you’ll add to host.go. The request and response, according to the 
API documentation, is as follows:
```shell script
$ curl https://api.shodan.io/shodan/host/search?key={YOUR_API_KEY}&query={query}&facets={facets}
{
    "matches": [
    {
        "os": null,
        "timestamp": "2014-01-15T05:49:56.283713",
        "isp": "Vivacom",
        "asn": "AS8866",
        "hostnames": [ ],
        "location": {
            "city": null,
            "region_code": null,
            "area_code": null,
            "longitude": 25,
            "country_code3": "BGR",
            "country_name": "Bulgaria",
            "postal_code": null,
            "dma_code": null,
            "country_code": "BG",
            "latitude": 43
        },
        "ip": 3579573318,
        "domains": [ ],
        "org": "Vivacom",
        "data": "@PJL INFO STATUS CODE=35078 DISPLAY="Power Saver" ONLINE=TRUE",
        "port": 9100,
        "ip_str": "213.91.244.70"
    },
    --snip--
    ],
    "facets": {
        "org": [
        {
            "count": 286,
            "value": "Korea Telecom"
        },
        --snip--
        ]
    },
    "total": 12039
}
```
Compared to the initial API call you implemented, this one is significantly more complex. Not only 
does the request take multiple parameters, but the JSON response contains nested data and arrays. 
For the following implementation, you’ll ignore the facets option and data, and instead focus on 
performing a string-based host search to process only the matches element of the response.

As you did before, start by building the Go structs to handle the response data; enter the types into 
your [host.go](shodan/host.go):
```go
type HostLocation struct {
    City         string  `json:"city"`
    RegionCode   string  `json:"region_code"`
    AreaCode     int     `json:"area_code"`
    Longitude    float32 `json:"longitude"`
    CountryCode3 string  `json:"country_code3"`
    CountryName  string  `json:"country_name"`
    PostalCode   string  `json:"postal_code"`
    DMACode      int     `json:"dma_code"`
    CountryCode  string  `json:"country_code"`
    Latitude     float32 `json:"latitude"`
}

type Host struct {
    OS        string       `json:"os"`
    Timestamp string       `json:"timestamp"`
    ISP       string       `json:"isp"`
    ASN       string       `json:"asn"`
    Hostnames []string     `json:"hostnames"`
    Location  HostLocation `json:"location"`
    IP        int64        `json:"ip"`
    Domains   []string     `json:"domains"`
    Org       string       `json:"org"`
    Data      string       `json:"data"`
    Port      int          `json:"port"`
    IPString  string       `json:"ip_str"`
}

type HostSearch struct {
    Matches []Host `json:"matches"`
}
```

The code defines three types:
  - `HostSearch` Used for parsing the matches array
  - `Host` Represents a single matches element
  - `HostLocation` Represents the location element within the host
  
`Notice that the types may not define all response fields. Go handles this elegantly, 
allowing you to define structures with only the JSON fields you care about.` Therefore, 
our code will parse the JSON just fine, while reducing the length of your code by including 
only the fields that are most relevant to the example.

To initialize and populate the struct, you’ll define the function in [host.go](shodan/host.go):
```go
func (s *Client) HostSearch(q string❶) (*HostSearch, error) {
    res, err := http.Get( ❷
        fmt.Sprintf("%s/shodan/host/search?key=%s&query=%s", BaseURL, s.apiKey, q),
    )
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    var ret HostSearch
    if err := json.NewDecoder(res.Body).Decode(&ret)❸; err != nil {
        return nil, err
    }

    return &ret, nil
}
```
The flow and logic is exactly like the APIInfo() method, except that you take the search query string 
as a parameter ❶, issue the call to the /shodan/host/search endpoint while passing the search term ❷, 
and decode the response into the HostSearch struct ❸.

You repeat this process of structure definition and function implementation for each API service you 
want to interact with. Rather than wasting precious pages here, we’ll jump ahead and show you the last 
step of the process: creating the client that uses your API code.


### Creating a Client
You’ll use a minimalistic approach to create your client: take a search term as a command line argument 
and then call the `APIInfo()` and `HostSearch()` methods, as in [main.go](cmd/shodan/main.go).
```go
func main() {
    if len(os.Args) != 2 {
        log.Fatalln("Usage: shodan searchterm")
    }
    apiKey := os.Getenv("SHODAN_API_KEY")❶
    s := shodan.New(apiKey)❷
    info, err := s.APIInfo()❸
    if err != nil {
        log.Panicln(err)
    }
    fmt.Printf(
        "Query Credits: %d\nScan Credits:  %d\n\n",
        info.QueryCredits,
        info.ScanCredits)

    hostSearch, err := s.HostSearch(os.Args[1])❹
    if err != nil {
        log.Panicln(err)
    }
 ❺ for _, host := range hostSearch.Matches {
        fmt.Printf("%18s%8d\n", host.IPString, host.Port)
    }
}
```
Start by reading your API key from the SHODAN_API_KEY environment variable ❶. Then use that value to 
initialize a new Client struct ❷, s, subsequently using it to call your APIInfo() method ❸. Call the 
HostSearch() method, passing in a search string captured as a command line argument ❹. Finally, loop 
through the results to display the IP and port values for those services matching the query string ❺. 
The following output shows a sample run, searching for the string tomcat:
```shell script
$ SHODAN_API_KEY=YOUR-KEY go run main.go tomcat
Query Credits: 100
Scan Credits:  100

    185.23.138.141    8081
   218.103.124.239    8080
     123.59.14.169    8081
      177.6.80.213    8181
    142.165.84.160   10000
--snip--
```
You’ll want to add error handling and data validation to this project, but it serves as a good example 
for fetching and displaying Shodan data with your new API. You now have a working codebase that can be 
easily extended to support and test the other Shodan functions.