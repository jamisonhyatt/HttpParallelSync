package main

import (
    "net/http"
    "time"
    "fmt"
    "log"
    "github.com/jamisonhyatt/HttpParallelSync"
    "gopkg.in/natefinch/lumberjack.v2"
)



func main() {

    log.SetOutput(&lumberjack.Logger{
        Filename:   "/var/log/http_parallel_sync/foo.log",
        MaxSize:    500, // megabytes
        MaxBackups: 3,
        MaxAge:     28, //days
    })


    client := NewCaddyClient()

    err := HttpParallelSync.Sync(client, "movies", 2)
    if (err != nil){
        log.Fatal(err)
    }


    log.Println("complete")
}

func NewCaddyClient() *HttpParallelSync.CaddyClient {
    caddy := HttpParallelSync.CaddyClient{
        Host: "localhost",
        Port: 2015,
        Ssl: false,
    }

    caddy.HttpClient = &http.Client{
        Timeout: time.Second * 25,

    }
    var protocol string
    if (caddy.Ssl) {
        protocol = "https"
    } else {
        protocol = "http"
    }
    caddy.BaseURI = fmt.Sprintf("%s://%s:%v",protocol, caddy.Host, caddy.Port)
    return &caddy
}
