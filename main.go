// With brainless help from Matthew f*cking Penner
// load testers: https://github.com/denji/awesome-http-benchmark
// redis pass: o6RN4AvGBp7KcjsycbDvnLy2x
package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net"
	"net/http"
)

// do some context thing. not really relevant
var ctx = context.Background()

// define the redis connection settings
var rdb = redis.NewClient(&redis.Options{
	Addr:     "not-today-fucker",
	Password: "yes-ive-already-reset-this",
})

var redirects = map[string]string{
	"ssh": "https://ssh.contact/",
}

var crock32 = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").
	WithPadding(base32.NoPadding)

func generateRandomSlug(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return crock32.EncodeToString(b), nil
}

func getDestination(slug string) (string, error) {
	dest, err := rdb.Get(ctx, slug).Result()
	return dest, err
}
func addDestination(slug string, dest string) error {
	err := rdb.Set(ctx, slug, dest, 0).Err()
	return err
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
	dest, err := getDestination(r.URL.Path[1:])

	//TODO: make this fail properly and provide some data to client
	if err != nil {
		http.Error(w, "Error when fetching destination for the redirect", http.StatusInternalServerError)

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

	//check that it works
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connected to redis server.")
	}

	//dump all predefined values into redis
	for slug, dest := range redirects {
		err := rdb.Set(ctx, slug, dest, 0).Err()
		if err != nil {
			fmt.Errorf("Error in dumping predefined redirect to redis. Slug: %s, Destination: %s.", slug, dest)
		}
	}

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
