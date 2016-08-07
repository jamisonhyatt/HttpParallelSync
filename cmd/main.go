package main

import (
    "net/http"
    "time"
    "fmt"
    "log"
    "github.com/jamisonhyatt/MultiDownload"
)



func main() {

    //v, err := gosnow.Default()
    //if (err != nil) {
    //    log.Fatal(err)
    //}


    client := NewCaddyClient()

    err := client.Sync("movies")
    if (err != nil){
        log.Fatal(err)
    }

    //fileFlake, err := v.Next()
    //_, err = client.GetFilePart(model.FileInfo{URL:"./Disney-RobinHood.mp4"}, 0, 0, fmt.Sprintf("1_%v", fileFlake))
    //
    //split := uint64(2389076156 / 2)
    //
    //fileFlake, err = v.Next()
    //fileName1 := fmt.Sprintf("1_%v", fileFlake)
    //_, err = client.GetFilePart(model.FileInfo{URL:"./Disney-RobinHood.mp4"}, 0, split, fileName1)
    //
    //fileFlake, err = v.Next()
    //fileName2 := fmt.Sprintf("2_%v", fileFlake)
    //_, err = client.GetFilePart(model.FileInfo{URL:"./Disney-RobinHood.mp4"}, split + 1, 0, fileName2)
    //
    //list := make([]string, 2)
    //list[0] = fileName1
    //list[1] = fileName2
    //
    //MultiDownload.Combine("RobinHood.mp4", list)
}

func NewCaddyClient() *HttpParallelSync.CaddyClient {
    caddy := HttpParallelSync.CaddyClient{
        Host: "localhost",
        Port: 2015,
        Ssl: false,
    }

    caddy.HttpClient = &http.Client{
        Timeout: time.Second * 10,

    }
    var protocol string
    if (caddy.Ssl) {
        protocol = "https"
    } else {
        protocol = "http"
    }
    caddy.BaseURI = fmt.Sprintf("%s://%s:%v",protocol, caddy.Host, caddy.Port)
    caddy.Parallelism = 2//runtime.NumCPU()
    return &caddy
}
