package main

import (
    "log"

    "diag-system/internal/agent"
)

func main() {
    a, err := agent.NewFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(a.Run())
}