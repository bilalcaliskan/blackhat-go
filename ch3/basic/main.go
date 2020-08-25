package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	/*
		Let’s begin the HTTP discussion by examining basic requests. Go’s net/http standard package contains several
		convenience functions to quickly and easily send POST, GET, and HEAD requests, which are arguably the most
		common HTTP verbs you’ll use. These functions take the following forms:
			- Get(url string) (resp *Response, err error)
			- Head(url string) (resp *Response, err error)
	 		- Post(url string, bodyType string, body io.Reader) (resp *Response, err error)
		Each function takes—as a parameter—the URL as a string value and uses it for the request’s destination. The Post()
		function is slightly more complex than the Get() and Head() functions. Post() takes two additional
		parameters: bodyType, which is a string value that you use for the Content-Type HTTP header (commonly
		application/x-www-form-urlencoded) of the request body, and an io.Reader.

		Note that the POST request creates the request body from form values and sets the Content-Type header. In each
		case, you must close the response body after you’re done reading data from it.
	*/


	// Read and display response body
	/*
		Inspecting various components of the HTTP response is a crucial aspect of any HTTP-related task, like reading
		the response body, accessing cookies and headers, or simply inspecting the HTTP status code.

		Following code block uses the ioutil.ReadAll() function to read data from the response body, does some error
		checking, and prints the HTTP status code and response body to stdout.

	*/
	resp, err := http.Get("https://www.google.com/robots.txt")
	if err != nil {
		log.Panicln(err)
	}
	// Print HTTP Status
	fmt.Println(resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println(string(body))
	defer resp.Body.Close()
	/*
		The Response type contains an exported Body parameter ❸, which is of type io.ReadCloser. An io.ReadCloser is
		an interface that acts as an io.Reader as well as an io.Closer, or an interface that requires the implementation
		of a Close() function to close the reader and perform any cleanup. The details are somewhat inconsequential;
		just know that after reading the data from an io.ReadCloser, you’ll need to call the Close() function ❹ on the
		response body. Using defer to close the response body is a common practice; this will ensure that the body is
		closed before you return it.
	*/


	resp, err = http.Head("https://www.google.com/robots.txt")
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)

	form := url.Values{}
	form.Add("foo", "bar")
	resp, err = http.Post("https://www.google.com/robots.txt", "application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()))
	/*
		Go has an additional POST request convenience function, called PostForm(), which removes the tediousness of
		setting those values and manually encoding every request.
			func PostForm(url string, data url.Values) (resp *Response, err error)
		If you want to substitute the PostForm() function for the Post() implementation above, you can use it:
			form := url.Values{}
			form.Add("foo", "bar")
			r3, err := http.PostForm("https://www.google.com/robots.txt", form)
			// Read response body and close.
		Unfortunately, no convenience functions exist for other HTTP verbs, such as PATCH, PUT, or DELETE. You’ll use
		these verbs mostly to interact with RESTful APIs, which employ general guidelines on how and why a server should
		use them; but nothing is set in stone, and HTTP is like the Old West when it comes to verbs. In fact, we’ve
		often toyed with the idea of creating a new web framework that exclusively uses DELETE for everything.
	*/
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)

	/*
		To generate a request with one of these verbs, you can use the NewRequest() function to create the Request
		struct, which you’ll subsequently send using the Client function’s Do() method. We promise that it’s simpler
		than it sounds. The function prototype for http.NewRequest() is as follows:
			func NewRequest(❶method, ❷url string, ❸body io.Reader) (req *Request, err error)
		You need to supply the HTTP verb ❶ and destination URL ❷ to NewRequest() as the first two string parameters.
		Much like the first POST example in Listing 3-1, you can optionally supply the request body by passing in an
		io.Reader as the third and final parameter.
	*/
	req, err := http.NewRequest("DELETE", "https://www.google.com/robots.txt", nil)
	if err != nil {
		log.Panicln(err)
	}
	var client http.Client
	resp, err = client.Do(req)
	defer resp.Body.Close()
	fmt.Println(resp.Status)

	/*
		The standard Go net/http library contains several functions that you can use to manipulate the request before
		it’s sent to the server. You’ll learn some of the more relevant and applicable variants as you work through
		practical examples throughout this chapter.
	*/

	req, err = http.NewRequest("PUT", "https://www.google.com/robots.txt", strings.NewReader(form.Encode()))
	resp, err = client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)
}