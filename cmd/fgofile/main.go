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

	host := flag.String("host", "", "FTP server host/address (client)")
	cliUser := flag.String("cuser", "", "client username (or -u)")
	cliPass := flag.String("cpass", "", "client password (or -P)")
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

	if *host == "" {
		fmt.Println("Missing --host")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  Server Mode:")
		fmt.Println("    fgofile --server [--port 2121] [--root ./ftp_root] [--suser u --spass p]")
		fmt.Println()
		fmt.Println("  Client Mode:")
		fmt.Println("    fgofile --host <ip> [--port 2121] [--cuser user --cpass pass | -u user -P pass]")
		fmt.Println()
		fmt.Println("Commands inside REPL:")
		fmt.Println("  ls, cd <dir>, pwd")
		fmt.Println("  get <file>, put <file>, rm <file>, mkdir <dir>, mv <src> <dst>")
		fmt.Println("  lpwd, lcd <dir>, lls (local commands)")
		fmt.Println("  quit")
		fmt.Println()
		fmt.Println("One-shot commands (no REPL):")
		fmt.Println("  fgofile --host <ip> [--port 2121] [auth flags] put <local> [remote]")
		fmt.Println("  fgofile --host <ip> [--port 2121] [auth flags] get <remote> [local]")
		os.Exit(2)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)

	c, err := ftp.Dial(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := c.LoginInteractive(*cliUser, *cliPass); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Connected to %s\n", addr)

	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "put":
			if len(args) < 2 {
				fmt.Println("usage: fgofile --host <ip> [--port 2121] [--cuser user --cpass pass | -u user -P pass] put <local> [remote]")
				os.Exit(2)
			}
			local := args[1]
			remote := ""
			if len(args) >= 3 {
				remote = args[2]
			}
			if err := c.Put(local, remote); err != nil {
				log.Fatal(err)
			}
			return

		case "get":
			if len(args) < 2 {
				fmt.Println("usage: fgofile --host <ip> [...] get <remote> [local]")
				os.Exit(2)
			}
			remote := args[1]
			local := ""
			if len(args) >= 3 {
				local = args[2]
			}
			if err := c.Get(remote, local); err != nil {
				log.Fatal(err)
			}
			return
		default:
		}
	}

	fmt.Printf("Connected to %s\n", addr)
	c.REPL()
}
