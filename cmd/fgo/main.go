package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    "fgofile/internal/ftp"
)

func main() {
    // Server flags
    isServer := flag.Bool("server", false, "run as server")
    port := flag.Int("port", 2121, "port")
    root := flag.String("root", "./ftp_root", "ftp root directory (server)")

    flag.Parse()

    if *isServer {
        if err := os.MkdirAll(*root, 0o755); err != nil { log.Fatal(err) }
        srv, err := ftp.NewServer("0.0.0.0", *port, *root)
        if err != nil { log.Fatal(err) }
        fmt.Printf("Server listening on %s\n", srv.Addr())
        log.Fatal(srv.Serve())
        return
    }

    // Client mode: args: <host> [--port N]
    if flag.NArg() < 1 {
        fmt.Println("Usage:\n  fgo --server [--port 2121] [--root ./ftp_root]\n  fgo <host> [--port 2121]")
        os.Exit(2)
    }
    host := flag.Arg(0)
    addr := fmt.Sprintf("%s:%d", host, *port)

    c, err := ftp.Dial(addr)
    if err != nil { log.Fatal(err) }
    defer c.Close()

    if err := c.LoginAnonymous(); err != nil { log.Fatal(err) }
    fmt.Printf("Connected to %s\n", addr)
    c.REPL()
}
