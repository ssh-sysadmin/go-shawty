// With brainless help from Matthew f*cking Penner
// load testers: https://github.com/denji/awesome-http-benchmark
// redis pass: o6RN4AvGBp7KcjsycbDvnLy2x
package main

import (
	"encoding/base32"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
)

var redirects = map[string]string{
	"ssh": "https://ssh.contact/",
}

var crock32 = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").
	WithPadding(base32.NoPadding)

func generateRandomSlug(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("argon2: failed to generate random bytes: %w", err)
	}
	return crock32.EncodeToString(b), nil
}

func getDestination(slug string) (string, bool) {
	v, exists := redirects[slug]
	return v, exists
}
func addDestination(slug string, destination string) error {
	redirects[slug] = destination
	return error(nil) //could be error value when networked DB
}

func handleAddRedirect(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 16*1024))
	if err != nil {
		http.Error(w, "Request body too long", http.StatusRequestEntityTooLarge)
		return
	}
	var slug string
	if len(r.URL.Path) < 2 {
		slug, err = generateRandomSlug(5)
		if err != nil {
			http.Error(w, "Failed to generate randomized slug and none provided", http.StatusInternalServerError)
			return
		}
	} else {
		slug = r.URL.Path[1:]
	}

	dest := string(body)
	err = addDestination(slug, dest)
	if err != nil {
		http.Error(w, "Unknown server error in addition of redirect mapping to store", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "/%s", slug)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	//if blank slug/path
	if len(r.URL.Path) < 1 {
		http.NotFound(w, r)
		return
	}
	dest, exists := getDestination(r.URL.Path[1:])

	if !exists {
		http.NotFound(w, r)
		return
	}
	//TODO: maybe i want to add expirys? maybe after networked DB?
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
