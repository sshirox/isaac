package main

import (
    "github.com/sshirox/isaac/internal/server"
)

func main() {
    if err := server.Run(); err != nil {
        panic(err)
    }
}
