package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type NotesCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type NotesRequest struct {
	Domain string `json:"domain"`
}

// https://golangcode.com/get-the-request-ip-addr/
// https://stackoverflow.com/a/33301173
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Real-IP")
	if forwarded != "" {
		return forwarded
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func GetDomain(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-Host")
	if forwarded != "" {
		return forwarded
	}
	return r.Host
}

func main() {
	log.Println("Welcome to the IPNotes Counter, see the readme for more information")

	domainCacheTimeout := 24 * time.Hour
	userCacheTimeout := 20 * time.Minute
	// userCacheTimeout := 30 * time.Second // debugging

	cacheOfCaches := ttlcache.New(
		ttlcache.WithTTL[string, *ttlcache.Cache[string, bool]](domainCacheTimeout),
	)

	// our main endpoint, receive the domain and IP and increment the count
	http.HandleFunc("/count", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Access-Control-Allow-Origin", "*")

		ip := GetIP(req)
		domain := req.Host

		// get the IP cache of this domain, or create it if it doesn't exist
		userResp := cacheOfCaches.Get(domain)
		var userCache *ttlcache.Cache[string, bool]
		if userResp == nil {
			newCache := ttlcache.New(
				ttlcache.WithTTL[string, bool](userCacheTimeout),
			)
			cacheOfCaches.Set(domain, newCache, domainCacheTimeout)
			userCache = newCache
			go userCache.Start()
		} else {
			userCache = userResp.Value()
		}

		// track this IP under this domain
		userCache.Set(ip, true, userCacheTimeout)

		// return the count
		count := NotesCount{Domain: domain, Count: int(userCache.Len())}
		output, _ := json.MarshalIndent(count, "", "\t")
		res.Write(output)
	})

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Access-Control-Allow-Origin", "*")

		// respond with the count for this domain, irrespective of IP (for read only)
		// get the domain from the request (query param)
		domain := req.URL.Query().Get("domain")
		if domain == "" {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Please provide a domain, e.g. ?domain=example.com"))
			return
		}

		userResp := cacheOfCaches.Get(domain)
		userCount := -1
		if userResp != nil {
			userCache := userResp.Value()
			userCount = int(userCache.Len())
		}

		// return the count
		count := NotesCount{Domain: domain, Count: userCount}
		output, _ := json.MarshalIndent(count, "", "\t")
		res.Write(output)
	})

	port := 7711
	log.Printf("Listening on localhost:%d\n", port)

	go cacheOfCaches.Start()

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
