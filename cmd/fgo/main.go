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

    cliUser := flag.String("u", "", "client username (alias of --username)")
    cliPass := flag.String("P", "", "client password (alias of --password)")

    flag.StringVar(cliUser, "username", "", "client username")
    flag.StringVar(cliPass, "password", "", "client password")

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
        fmt.Println("Usage:\n  fgo --server [--port 2121] [--root ./ftp_root] [--username u --password p]\n  fgo <host> [--port 2121] [--username u --password p]")
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
