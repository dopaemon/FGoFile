package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dopaemon/fgofile/internal/ftp"
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
		rootAbs, err := filepath.Abs(*root)
		if err != nil {
			log.Fatal(err)
		}
		if err := os.MkdirAll(rootAbs, 0o755); err != nil {
			log.Fatal(err)
		}

		srv, err := ftp.NewServer("0.0.0.0", *port, rootAbs, *srvUser, *srvPass)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Server listening on %s\n", srv.Addr())
		fmt.Printf("  Root : %s\n", rootAbs)
		fmt.Printf("  Auth : %v (user=%q)\n", srv.RequireAuth(), *srvUser)

		log.Fatal(srv.Serve())
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage:")
		fmt.Println("  As server:")
		fmt.Println("    fgofile --server [--port 2121] [--root ./ftp_root] [--suser u --spass p]")
		fmt.Println()
		fmt.Println("  As client:")
		fmt.Println("    fgofile <host> [--port 2121] [--cuser u --cpass p | -u u -P p]")
		fmt.Println()
		fmt.Println("Client commands (inside REPL):")
		fmt.Println("  ls, cd <dir>, pwd")
		fmt.Println("  get <file>, put <file>, rm <file>, mkdir <dir>, mv <src> <dst>")
		fmt.Println("  lpwd, lcd <dir>, lls (local directory & files)")
		fmt.Println("  quit")
		os.Exit(2)
	}

	host := flag.Arg(0)
	addr := fmt.Sprintf("%s:%d", host, *port)

	c, err := ftp.Dial(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := c.LoginInteractive(*cliUser, *cliPass); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Connected to %s\n", addr)
	c.REPL()
}
