package ftp

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

type Server struct {
    ln   net.Listener
    root string
}

func NewServer(host string, port int, root string) (*Server, error) {
    ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
    if err != nil { return nil, err }
    return &Server{ln: ln, root: abs(root)}, nil
}

func (s *Server) Addr() string { return s.ln.Addr().String() }

func (s *Server) Serve() error {
    for {
        c, err := s.ln.Accept()
        if err != nil { return err }
        go s.handle(c)
    }
}

func (s *Server) handle(c net.Conn) {
    defer c.Close()
    rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
    cwd := s.root
    var pasv net.Listener

    write := func(format string, a ...any) {
        fmt.Fprintf(rw, format+"\r\n", a...)
        rw.Flush()
    }

    write("220 Simple FTP Ready")
    for {
        line, err := rw.ReadString('\n')
        if err != nil { return }
        line = strings.TrimRight(line, "\r\n")
        if line == "" { continue }
        parts := strings.SplitN(line, " ", 2)
        cmd := strings.ToUpper(parts[0])
        arg := ""; if len(parts) > 1 { arg = parts[1] }

        switch cmd {
        case "USER": write("331 Password required")
        case "PASS": write("230 Logged in")
        case "SYST": write("215 UNIX Type: L8")
        case "TYPE": write("200 Type set") // accept TYPE I
        case "PWD":  write("257 \"/%s\"", relFrom(cwd, s.root))
        case "CWD":
            if arg == "" { write("550 Missing path"); break }
            np := secureJoin(cwd, s.root, arg)
            if !isDir(np) { write("550 Failed to change directory") } else { cwd = np; write("250 Directory changed") }
        case "PASV":
            if pasv != nil { pasv.Close() }
            ln, err := net.Listen("tcp", ":0")
            if err != nil { write("425 Cannot open passive connection"); break }
            pasv = ln
            // use control local IP for PASV reply
            hostIP, _, _ := net.SplitHostPort(c.LocalAddr().String())
            ip := toFour(hostIP)
            _, pstr, _ := net.SplitHostPort(ln.Addr().String())
            pn, _ := strconv.Atoi(pstr)
            p1, p2 := pn/256, pn%256
            write("227 Entering Passive Mode (%d,%d,%d,%d,%d,%d)", ip[0], ip[1], ip[2], ip[3], p1, p2)
        case "LIST":
            if pasv == nil { write("425 Use PASV first"); break }
            write("150 Opening data")
            dc, err := pasv.Accept(); if err != nil { write("550 LIST failed"); break }
            // very simple listing
            ents, _ := os.ReadDir(cwd)
            for _, e := range ents {
                fmt.Fprintf(dc, "-rw-r--r-- 1 u g 0 Jan 01 00:00 %s\r\n", e.Name())
            }
            dc.Close(); pasv.Close(); pasv = nil
            write("226 Done")
        case "RETR":
            if pasv == nil { write("425 Use PASV first"); break }
            if arg == "" { write("501 RETR needs filename"); break }
            fp := secureJoin(cwd, s.root, arg)
            if !isFile(fp) { write("550 File not found"); break }
            write("150 Opening file")
            dc, err := pasv.Accept(); if err != nil { write("550 RETR failed"); break }
            f, err := os.Open(fp); if err == nil { io.Copy(dc, f); f.Close() }
            dc.Close(); pasv.Close(); pasv = nil
            write("226 Done")
        case "STOR":
            if pasv == nil { write("425 Use PASV first"); break }
            if arg == "" { write("501 STOR needs filename"); break }
            fp := secureJoin(cwd, s.root, arg)
            write("150 Ok to send data")
            dc, err := pasv.Accept(); if err != nil { write("550 STOR failed"); break }
            f, _ := os.Create(fp)
            io.Copy(f, dc)
            f.Close(); dc.Close(); pasv.Close(); pasv = nil
            write("226 Done")
        case "QUIT": write("221 Bye"); return
        default: write("502 Not implemented")
        }
    }
}

func isDir(p string) bool { st, err := os.Stat(p); return err == nil && st.IsDir() }
func isFile(p string) bool { st, err := os.Stat(p); return err == nil && st.Mode().IsRegular() }

func abs(p string) string { ap, _ := filepath.Abs(p); return ap }

func relFrom(curr, root string) string {
    rel, _ := filepath.Rel(root, curr)
    if rel == "." { return "" }
    return filepath.ToSlash(rel)
}

func secureJoin(cwd, root, user string) string {
    if strings.HasPrefix(user, "/") { user = "." + user }
    p := filepath.Clean(filepath.Join(cwd, user))
    if !strings.HasPrefix(p, root) { return cwd }
    return p
}

func toFour(ip string) [4]int { var a [4]int; fmt.Sscanf(ip, "%d.%d.%d.%d", &a[0], &a[1], &a[2], &a[3]); return a }
