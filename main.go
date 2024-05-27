// With brainless help from Matthew f*cking Penner
// load testers: https://github.com/denji/awesome-http-benchmark
// redis pass: o6RN4AvGBp7KcjsycbDvnLy2x
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

var redirects = map[string]string{
	"ssh": "https://ssh.contact/",
}

func getDestination(path string) (string, bool) {
	v, exists := redirects[path]
	return v, exists
}
func addDestination(path string, destination string) error {
	redirects[path] = destination
	return error(nil) //could be error value when networked DB
}

func handleAddRedirect(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 16*1024))
	if err != nil {
		http.Error(w, "Request body too long", http.StatusRequestEntityTooLarge)
		return
	}
	if len(r.URL.Path) < 1 {

	}
	path := r.URL.Path[1:]
	dest := string(body)
	err = addDestination(path, dest)
	if err != nil {
		http.Error(w, "Unknown server error in addition of redirect mapping to store", http.StatusInternalServerError)
		return
	}

}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	dest, exists := getDestination(r.URL.Path[1:])

	if len(r.URL.Path) < 1 {
		http.NotFound(w, r)
		return
	}
	if !exists {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, dest, http.StatusTemporaryRedirect)

}

func handleStatus(w http.ResponseWriter, r *http.Request) {

	userAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	fmt.Fprintf(w, "I'm alive, your IP is %s", userAddr)

}

func main() {
	http.HandleFunc("/*", handleRedirect)
	http.HandleFunc("POST /*", handleAddRedirect)
	http.HandleFunc("/status", handleStatus)

	fmt.Println("Server is listening on port 42069")

	err := http.ListenAndServe(":42069", nil)
	if err != nil {
		log.Printf(err.Error())
		return
	}

}
