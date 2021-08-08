package main

import (
    "context"
    "flag"
    "fmt"
    log "github.com/sirupsen/logrus"
    "github.com/umi0410/streamingChat/client"
    "github.com/umi0410/streamingChat/streamingChat"
)

var (
    mode string
)

func init(){
    flag.Parse()
    mode = flag.Arg(0)
}


func main() {
    if mode == "server" {
        streamingChat.NewServer().Run(context.Background())
    } else if mode == "client" {
        fmt.Print("Please input your username: ")
        var username string
        fmt.Scan(&username)
        client.RunClient(username)
    } else if mode == "fakeClient" {
        go client.RunClient("DummyMummy")
        go client.RunClient("Mike")
        go client.RunClient("Michael")
        go client.RunClient("Harry")
        go client.RunClient("James")
        client.RunClient("Chocolate")
    }  else{
        log.Error(flag.Args(), mode)
        panic("Invalid mode, given " + mode)
    }
}
