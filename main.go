package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/xyproto/simpleredis/v2"
)

var (
	redisEnabled bool
	masterPool   *simpleredis.ConnectionPool
	replicaPool  *simpleredis.ConnectionPool

	// in-memory guestbook
	guestbookEntries = make([]string, 0)

	buf    bytes.Buffer
	logger *log.Logger
)

func init() {
	logger = log.New(&buf, "logger: ", log.Lshortfile)
}

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
	_, _ = rw.Write(membersJSON)
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

// ConsumeMemory is a function that consumes memory in a loop, simulating a memory leak.
func ConsumeMemory(ctx context.Context) {
	mbStr := os.Getenv("CONSUME_MEMORY_MB")
	if mbStr == "" {
		return
	}
	var mb int64
	mb, err := strconv.ParseInt(mbStr, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid CONSUME_MEMORY_MB value: %s", mbStr))
	}
	if mb <= 0 {
		return
	}
	logger.Print("Starting memory consumption with ", mb, " MB\n")

	for i := 0; i < int(mb); i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// make a memory allocation
					// this consumes 1mb of memory
					memory := make([]byte, 1024*1024)
					time.Sleep(1 * time.Second)
					fmt.Println("Hello, World!", memory)
					memory = nil
				}
			}
		}(ctx)
	}
	<-ctx.Done()
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
	ctx := context.Background()
	go ConsumeMemory(ctx)

	r := mux.NewRouter()
	r.Path("/lrange/{key}").Methods("GET").HandlerFunc(ListRangeHandler)
	r.Path("/rpush/{key}/{value}").Methods("GET").HandlerFunc(ListPushHandler)
	r.Path("/info").Methods("GET").HandlerFunc(InfoHandler)
	r.Path("/env").Methods("GET").HandlerFunc(EnvHandler)

	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(":3000")
}
