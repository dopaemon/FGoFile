package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    "fgofile/internal/ftp"
)

func main() {
    isServer := flag.Bool("server", false, "run as server")
    port := flag.Int("port", 2121, "port")

    root := flag.String("root", "./ftp_root", "ftp root directory (server)")
    srvUser := flag.String("suser", "", "server auth username (optional)")
    srvPass := flag.String("spass", "", "server auth password (optional)")

    cliUser := flag.String("cuser", "", "client username (or use -u)")
    cliPass := flag.String("cpass", "", "client password (or use -P)")
    flag.StringVar(cliUser, "u", *cliUser, "client username (alias)")
    flag.StringVar(cliPass, "P", *cliPass, "client password (alias)")

    flag.Parse()

    if *isServer {
        if err := os.MkdirAll(*root, 0o755); err != nil { log.Fatal(err) }
        srv, err := ftp.NewServer("0.0.0.0", *port, *root, *srvUser, *srvPass)
        if err != nil { log.Fatal(err) }
        fmt.Printf("Server listening on %s (auth: %v)\n", srv.Addr(), srv.RequireAuth())
        log.Fatal(srv.Serve())
        return
    }

    if flag.NArg() < 1 {
        fmt.Println("Usage:\n  fgofile --server [--port 2121] [--root ./ftp_root] [--suser u --spass p]\n  fgofile <host> [--port 2121] [--cuser u --cpass p | -u u -P p]")
        os.Exit(2)
    }
    host := flag.Arg(0)
    addr := fmt.Sprintf("%s:%d", host, *port)

    c, err := ftp.Dial(addr)
    if err != nil { log.Fatal(err) }
    defer c.Close()

    if err := c.LoginInteractive(*cliUser, *cliPass); err != nil { log.Fatal(err) }

    fmt.Printf("Connected to %s\n", addr)
    c.REPL()
}
