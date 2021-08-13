// streaming-api/main.go
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "sync"
    "syscall"
    "time"

    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
    "github.com/joho/godotenv"
    "github.com/pusher/pusher-http-go"
)

type cache struct {
    counter map[string]int64
    mu      sync.RWMutex
}
func (c *cache) Init(options ...string) {
    for _, v := range options {
        c.counter[strings.TrimSpace(v)] = 0
    }
}

func (c *cache) All() map[string]int64 {
    c.mu.Lock()
    defer c.mu.Unlock()

    return c.counter
}

func (c *cache) Incr(option string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.counter[strings.TrimSpace(option)]++
}

func (c *cache) Count(option string) int64 {
    c.mu.RLock()
    defer c.mu.RUnlock()

    val, ok := c.counter[strings.TrimSpace(option)]
    if !ok {
        return 0
    }

    return val
}

func main() {

    options := flag.String("options", "Messi,Suarez,Trump", "What items to search for on Twitter ?")
    httpPort := flag.Int("http.port", 1500, "What port to run HTTP on ?")
    channelsPublishInterval := flag.Duration("channels.duration", 3*time.Second, "How much duration before data is published to Pusher Channels")

    flag.Parse()

    if err := godotenv.Load(); err != nil {
        log.Fatalf("could not load .env file.. %v", err)
    }

    appID := os.Getenv("PUSHER_APP_ID")
    appKey := os.Getenv("PUSHER_APP_KEY")
    appSecret := os.Getenv("PUSHER_APP_SECRET")
    appCluster := os.Getenv("PUSHER_APP_CLUSTER")
    appIsSecure := os.Getenv("PUSHER_APP_SECURE")

    var isSecure bool
    if appIsSecure == "1" {
        isSecure = true
    }

    pusherClient := &pusher.Client{
                       AppId:   appID,
                       Key:     appKey,
                       Secret:  appSecret,
                       Cluster: appCluster,
                       Secure:  isSecure,
    }

    config := oauth1.NewConfig(os.Getenv("TWITTER_CONSUMER_KEY"), os.Getenv("TWITTER_CONSUMER_SECRET"))
    token := oauth1.NewToken(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_SECRET"))

    httpClient := config.Client(oauth1.NoContext, token)

    client := twitter.NewClient(httpClient)

    optionsCache := &cache {
        mu:      sync.RWMutex{},
        counter: make(map[string]int64),
    }

    splittedOptions := strings.Split(*options, ",")

    if n := len(splittedOptions); n < 2 {
        log.Fatalf("There must be at least 2 options... %v ", splittedOptions)
    } else if n > 3 {
        log.Fatalf("There cannot be more than 3 options... %v", splittedOptions)
    }

    optionsCache.Init(splittedOptions...)

    go func() {
        var t *template.Template
        var once sync.Once

        http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

        http.Handle("/polls", http.HandlerFunc(poll(optionsCache)))
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            once.Do(func() {
                tem, err := template.ParseFiles("index.html")
                if err != nil {
                    log.Fatal(err)
                }

                t = tem.Lookup("index.html")
            })
            t.Execute(w, nil)
        })
        http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil)
    }()

    go func(c *cache, client *pusher.Client) {
        t := time.NewTicker(*channelsPublishInterval)

        for {
           select {
            case <-t.C:
                pusherClient.Trigger("twitter-votes", "options", c.All())
            }
        }

    }(optionsCache, pusherClient)

    demux := twitter.NewSwitchDemux()
    demux.Tweet = func(tweet *twitter.Tweet) {
       fmt.Println(tweet.Text)
        for _, v := range splittedOptions {
            if strings.Contains(tweet.Text, v) {
                optionsCache.Incr(v)
            }
        }
    }

    fmt.Println("Starting Stream...")

    filterParams := &twitter.StreamFilterParams{
        Track:         splittedOptions,
        StallWarnings: twitter.Bool(true),
    }

    stream, err := client.Streams.Filter(filterParams)
    if err != nil {
        log.Fatal(err)
    }

    go demux.HandleChan(stream.Messages)

    // Wait for SIGINT and SIGTERM (HIT CTRL-C)
    ch := make(chan os.Signal)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    log.Println(<-ch)

    fmt.Println("Stopping Stream...")
    stream.Stop()
}

func poll(cache *cache) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(cache.All())
    }
}
