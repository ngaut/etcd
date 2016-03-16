package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/clientv3"
)

var etcd = flag.String("etcd", "", "etcd servers")

const (
	requestTimeout = 3 * time.Second
	dialTimeout    = 3 * time.Second
)

type gateway struct {
	// etcd cluster
	servers []string
	c       *clientv3.Client
}

func (gw *gateway) start() {
	http.HandleFunc("/get", gw.handleGet)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (gw *gateway) handleGet(w http.ResponseWriter, r *http.Request) {
	log.Println("request:", r.URL.Path)
	r.ParseForm()
	key := r.PostForm.Get("key")
	if len(key) == 0 {
		http.NotFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	kv := clientv3.NewKV(gw.c)
	resp, err := kv.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		// TODO: error
		return
	}
	if len(resp.Kvs) == 0 {
		http.NotFound(w, r)
		return
	}

	for _, kv := range resp.Kvs {
		log.Printf("%+v", kv)
	}
}

func main() {
	flag.Parse()
	gw := gateway{}
	log.Printf("create etcd v3 client with endpoints %v", etcd)
	gw.servers = strings.Split(*etcd, ",")
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   gw.servers,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Panic(err)
	}
	gw.c = client
	gw.start()
}
