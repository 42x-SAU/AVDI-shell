package main

import (
    "log"

    "diag-system/internal/server"
)

func main() {
    srv, err := server.New()
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(srv.Run())
}