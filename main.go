package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/xyproto/simpleredis/v2"
)

var (
	redisEnabled bool
	masterPool   *simpleredis.ConnectionPool
	replicaPool  *simpleredis.ConnectionPool

	// in-memory guestbook
	guestbookEntries = make([]string, 0)
)

func ListRangeHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	var membersJSON []byte
	if redisEnabled {
		list := simpleredis.NewList(replicaPool, key)
		members := HandleError(list.GetAll()).([]string)
		membersJSON = HandleError(json.MarshalIndent(members, "", "  ")).([]byte)
	} else {
		membersJSON = HandleError(json.MarshalIndent(guestbookEntries, "", "  ")).([]byte)
	}
	rw.Write(membersJSON)
}

func ListPushHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	value := mux.Vars(req)["value"]
	if redisEnabled {
		list := simpleredis.NewList(masterPool, key)
		HandleError(nil, list.Add(value))
	} else {
		guestbookEntries = append(guestbookEntries, value)
	}
	ListRangeHandler(rw, req)
}

func InfoHandler(rw http.ResponseWriter, req *http.Request) {
	var info []byte
	if redisEnabled {
		info = HandleError(masterPool.Get(0).Do("INFO")).([]byte)
	} else {
		info = []byte(`redis not enabled`)
	}
	rw.Write(info)
}

func EnvHandler(rw http.ResponseWriter, req *http.Request) {
	environment := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		key := splits[0]
		val := strings.Join(splits[1:], "=")
		environment[key] = val
	}

	envJSON := HandleError(json.MarshalIndent(environment, "", "  ")).([]byte)
	rw.Write(envJSON)
}

func HandleError(result interface{}, err error) (r interface{}) {
	if err != nil {
		panic(err)
	}
	return result
}

func main() {
	redisMaster := flag.String("redis-master", "", "Redis master (e.g. redis-master:6379)")
	redisReplica := flag.String("redis-replica", "", "Redis replica (e.g. redis-replica:6379)")

	redisEnabled = *redisMaster != "" && *redisReplica != ""
	if redisEnabled {
		masterPool = simpleredis.NewConnectionPoolHost(*redisMaster)
		defer masterPool.Close()
		replicaPool = simpleredis.NewConnectionPoolHost("redis-replica:6379")
		defer replicaPool.Close()
	}

	r := mux.NewRouter()
	r.Path("/lrange/{key}").Methods("GET").HandlerFunc(ListRangeHandler)
	r.Path("/rpush/{key}/{value}").Methods("GET").HandlerFunc(ListPushHandler)
	r.Path("/info").Methods("GET").HandlerFunc(InfoHandler)
	r.Path("/env").Methods("GET").HandlerFunc(EnvHandler)

	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(":3000")
}
