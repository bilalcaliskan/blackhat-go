package main

func main() {
	/*
		The Shodan URL is defined as a constant value in shodan/shodan.go, that way, you can easily access and reuse
		it within your implementing functions. If Shodan ever changes the URL of its API, you’ll have to make the
		change at only this one location in order to correct your entire codebase. Next, you define a Client struct,
		used for maintaining your API token across requests. Finally, the code defines a New() helper function, taking
		the API token as input and creating and returning an initialized Client instance.
	*/
}