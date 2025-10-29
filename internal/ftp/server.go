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
    ln       net.Listener
    root     string
    user     string
    pass     string
    require  bool
}

type session struct {
    ctrl  net.Conn
    rw    *bufio.ReadWriter
    cwd   string
    authed bool
    userGiven string
    pasv  net.Listener
}

func NewServer(host string, port int, root string, username string, password string) (*Server, error) {
    ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
    if err != nil { return nil, err }
    rootAbs, _ := filepath.Abs(root)
    return &Server{ln: ln, root: rootAbs, user: username, pass: password, require: username != ""}, nil
}

func (s *Server) Addr() string { return s.ln.Addr().String() }
func (s *Server) RequireAuth() bool { return s.require }

func (s *Server) Serve() error {
    for {
        c, err := s.ln.Accept()
        if err != nil { return err }
        go s.handle(c)
    }
}

func (s *Server) handle(c net.Conn) {
    defer c.Close()
    ss := &session{
        ctrl: c,
        rw:   bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c)),
        cwd:  s.root,
    }
    ss.write("220 Simple FTP Ready")

    for {
        line, err := ss.rw.ReadString('\n')
        if err != nil { return }
        line = strings.TrimRight(line, "\r\n")
        if line == "" { continue }
        parts := strings.SplitN(line, " ", 2)
        cmd := strings.ToUpper(parts[0])
        arg := ""; if len(parts) > 1 { arg = parts[1] }

        if cmd == "USER" {
            ss.userGiven = arg
            if s.require {
                ss.write("331 Password required")
            } else {
                ss.write("331 Password required")
            }
            continue
        }
        if cmd == "PASS" {
            if s.require {
                if ss.userGiven == s.user && arg == s.pass {
                    ss.authed = true
                    ss.write("230 Logged in")
                } else {
                    ss.write("530 Authentication failed")
                }
            } else {
                ss.authed = true
                ss.write("230 Logged in")
            }
            continue
        }

        if s.require && !ss.authed {
            ss.write("530 Please login with USER and PASS")
            continue
        }

        switch cmd {
        case "SYST": ss.write("215 UNIX Type: L8")
        case "TYPE": ss.write("200 Type set")
        case "PWD":  ss.write(fmt.Sprintf("257 \"/%s\"", relFrom(ss.cwd, s.root)))
        case "CWD":
            if arg == "" { ss.write("550 Missing path"); break }
            np := secureJoin(ss.cwd, s.root, arg)
            if !isDir(np) { ss.write("550 Failed to change directory") } else { ss.cwd = np; ss.write("250 Directory changed") }
        case "PASV":
            if ss.pasv != nil { ss.pasv.Close() }
            ln, err := net.Listen("tcp", ":0"); if err != nil { ss.write("425 Cannot open passive connection"); break }
            ss.pasv = ln
            hostIP, _, _ := net.SplitHostPort(ss.ctrl.LocalAddr().String())
            ip := toFour(hostIP)
            _, pstr, _ := net.SplitHostPort(ln.Addr().String())
            pn, _ := strconv.Atoi(pstr)
            p1, p2 := pn/256, pn%256
            ss.write(fmt.Sprintf("227 Entering Passive Mode (%d,%d,%d,%d,%d,%d)", ip[0], ip[1], ip[2], ip[3], p1, p2))
        case "LIST":
            if ss.pasv == nil { ss.write("425 Use PASV first"); break }
            ss.write("150 Opening data")
            dc, err := ss.pasv.Accept(); if err != nil { ss.write("550 LIST failed"); break }
            ents, _ := os.ReadDir(ss.cwd)
            for _, e := range ents { fmt.Fprintf(dc, "-rw-r--r-- 1 u g 0 Jan 01 00:00 %s\r\n", e.Name()) }
            dc.Close(); ss.pasv.Close(); ss.pasv = nil
            ss.write("226 Done")
        case "RETR":
            if ss.pasv == nil { ss.write("425 Use PASV first"); break }
            if arg == "" { ss.write("501 RETR needs filename"); break }
            fp := secureJoin(ss.cwd, s.root, arg)
            if !isFile(fp) { ss.write("550 File not found"); break }
            ss.write("150 Opening file")
            dc, err := ss.pasv.Accept(); if err != nil { ss.write("550 RETR failed"); break }
            f, err := os.Open(fp); if err == nil { io.Copy(dc, f); f.Close() }
            dc.Close(); ss.pasv.Close(); ss.pasv = nil
            ss.write("226 Done")
        case "STOR":
            if ss.pasv == nil { ss.write("425 Use PASV first"); break }
            if arg == "" { ss.write("501 STOR needs filename"); break }
            fp := secureJoin(ss.cwd, s.root, arg)
            ss.write("150 Ok to send data")
            dc, err := ss.pasv.Accept(); if err != nil { ss.write("550 STOR failed"); break }
            f, _ := os.Create(fp)
            io.Copy(f, dc)
            f.Close(); dc.Close(); ss.pasv.Close(); ss.pasv = nil
            ss.write("226 Done")
        case "DELE":
            if arg == "" { ss.write("501 DELE needs filename"); break }
            fp := secureJoin(ss.cwd, s.root, arg)
            if err := os.Remove(fp); err != nil { ss.write("550 Delete failed") } else { ss.write("250 Deleted") }
        case "MKD":
            if arg == "" { ss.write("501 MKD needs dirname"); break }
            fp := secureJoin(ss.cwd, s.root, arg)
            if err := os.MkdirAll(fp, 0o755); err != nil { ss.write("550 MKD failed") } else { ss.write("257 Directory created") }
        case "RNFR":
            if arg == "" { ss.write("501 RNFR needs path"); break }
            ss.userGiven = secureJoin(ss.cwd, s.root, arg)
            if _, err := os.Stat(ss.userGiven); err != nil { ss.write("550 RNFR path invalid") } else { ss.write("350 Ready for RNTO") }
        case "RNTO":
            if ss.userGiven == "" { ss.write("503 Bad sequence of commands"); break }
            newp := secureJoin(ss.cwd, s.root, arg)
            if err := os.Rename(ss.userGiven, newp); err != nil { ss.write("550 RNTO failed") } else { ss.write("250 Renamed") }
            ss.userGiven = ""
        case "QUIT": ss.write("221 Bye"); return
        default: ss.write("502 Not implemented")
        }
    }
}

func (s *session) write(line string) { fmt.Fprintf(s.rw, "%s\r\n", line); s.rw.Flush() }

func isDir(p string) bool { st, err := os.Stat(p); return err == nil && st.IsDir() }
func isFile(p string) bool { st, err := os.Stat(p); return err == nil && st.Mode().IsRegular() }

func relFrom(curr, root string) string { rel, _ := filepath.Rel(root, curr); if rel == "." { return "" }; return filepath.ToSlash(rel) }

func secureJoin(cwd, root, user string) string {
    if strings.HasPrefix(user, "/") { user = "." + user }
    p := filepath.Clean(filepath.Join(cwd, user))
    if !strings.HasPrefix(p, root) { return cwd }
    return p
}

func toFour(ip string) [4]int { var a [4]int; fmt.Sscanf(ip, "%d.%d.%d.%d", &a[0], &a[1], &a[2], &a[3]); return a }
