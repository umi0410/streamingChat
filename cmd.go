package main

import (
    "bufio"
    "context"
    "flag"
    "fmt"
    log "github.com/sirupsen/logrus"
    "github.com/umi0410/streamingChat/client"
    "github.com/umi0410/streamingChat/streamingChat"
    "os"
    "os/signal"
)

var (
    mode string
    scanner = bufio.NewScanner(os.Stdin)
)

func init(){
    flag.Parse()
    mode = flag.Arg(0)
    log.SetLevel(log.DebugLevel)
}


func main() {
    ctx := WithGracefullyShutDownContext()

    if mode == "server" {
        log.Error(streamingChat.NewChatServer().Start(ctx))
    } else if mode == "client" {
        fmt.Print("Please input your username: ")
        scanner.Scan()
        username := scanner.Text()
        c := client.NewChatClient(username)
        log.Error(c.Start(ctx))
    } else if mode == "fakeClient" {
        for _, fakeName := range []string{"Dummy", "Mike", "Coke", "Pizza", "Pasta"}{
            c := client.NewChatClient(fakeName)
            log.Error(c.Start(ctx))
        }
        c := client.NewChatClient("Chocolate")
        log.Error(c.Start(ctx))
    }  else{
        log.Error(flag.Args(), mode)
        panic("Invalid mode, given " + mode)
    }
}

func WithGracefullyShutDownContext() context.Context{
    sigInt := make(chan os.Signal, 10)
    signal.Notify(sigInt, os.Interrupt)
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        <- sigInt
        log.Warning("인터럽트 발생! context를 Cancel합니다.")
        cancel()
    }()
    return ctx
}
