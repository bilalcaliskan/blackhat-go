# HTTP Clients And Remote Interaction With Tools
This chapter focuses on the client side. It will first introduce you to the basics of 
building and customizing HTTP requests and receiving their responses. Then you’ll learn 
how to parse structured response data so the client can interrogate the information to 
determine actionable or relevant data. Finally, you’ll learn how to apply these fundamentals 
by building HTTP clients that interact with a variety of security tools and resources. The 
clients you develop will query and consume the APIs of `Shodan`, `Bing`, and `Metasploit` 
and will search and parse document metadata in a manner similar to the metadata search tool [FOCA](https://github.com/ElevenPaths/FOCA).

### What is [FOCA(Fingerprinting Organizations with Collected Archives)](https://github.com/ElevenPaths/FOCA)?
FOCA is a tool used mainly to find metadata and hidden information in the documents it scans. 
These documents may be on web pages, and can be downloaded and analysed with FOCA.
It is capable of analysing a wide variety of documents, with the most common being 
Microsoft Office, Open Office, or PDF files, although it also analyses Adobe InDesign 
or SVG files, for instance.

These documents are searched for using three possible search engines: Google, Bing, 
and DuckDuckGo. The sum of the results from the three engines amounts to a lot of 
documents. 


### HTTP Fundamentals With GO
First, HTTP is a stateless protocol: the server doesn’t inherently maintain state and 
status for each request. Instead, state is tracked through a variety of means, which 
may include session identifiers, cookies, HTTP headers, and more. The client and 
servers have a responsibility to properly negotiate and validate this state.

Second, communications between clients and servers can occur either synchronously 
or asynchronously, but they operate on a request/response cycle. You can include 
several options and headers in the request in order to influence the behavior of the 
server and to create usable web applications. Most commonly, servers host files that 
a web browser renders to produce a graphical, organized, and stylish representation 
of the data. But the endpoint can serve arbitrary data types. APIs commonly communicate 
via more structured data encoding, such as XML, JSON, or MSGRPC. In some cases, the 
data retrieved may be in binary format, representing an arbitrary file type for download.

Finally, Go contains convenience functions so you can quickly and easily build and send 
HTTP requests to a server and subsequently retrieve and process the response. Through 
some of the mechanisms you’ve learned in previous chapters, you’ll find that the 
conventions for handling structured data prove extremely convenient when interacting 
with HTTP APIs.


For the coding exercises, check the subfolders in current directory.