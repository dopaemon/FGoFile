package ftp

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
)

type Client struct {
    ctrl net.Conn
    rw   *bufio.ReadWriter
}

func Dial(addr string) (*Client, error) {
    c, err := net.Dial("tcp", addr)
    if err != nil { return nil, err }
    rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
    rw.ReadString('\n')
    return &Client{ctrl: c, rw: rw}, nil
}

func (c *Client) Close() error { return c.ctrl.Close() }

func (c *Client) send(cmd string) string {
    fmt.Fprintf(c.rw, "%s\r\n", cmd)
    c.rw.Flush()
    s, _ := c.rw.ReadString('\n')
    return s
}

func (c *Client) LoginInteractive(user, pass string) error {
    if user == "" {
        in := bufio.NewReader(os.Stdin)
        fmt.Print("Username (Enter for anonymous): ")
        u, _ := in.ReadString('\n')
        user = strings.TrimSpace(u)
    }
    if user == "" {
        c.send("USER anonymous")
        c.send("PASS guest")
        c.send("TYPE I")
        return nil
    }
    if pass == "" {
        in := bufio.NewReader(os.Stdin)
        fmt.Print("Password: ")
        p, _ := in.ReadString('\n')
        pass = strings.TrimSpace(p)
    }
    r1 := strings.TrimSpace(c.send("USER "+user))
    if strings.HasPrefix(r1, "530") { return fmt.Errorf(r1) }
    r2 := strings.TrimSpace(c.send("PASS "+pass))
    if !strings.HasPrefix(r2, "230") { return fmt.Errorf(r2) }
    c.send("TYPE I")
    return nil
}

var pasvRE = regexp.MustCompile(`\((\d+),(\d+),(\d+),(\d+),(\d+),(\d+)\)`)

func (c *Client) pasvDial() (net.Conn, string, error) {
    resp := c.send("PASV")
    m := pasvRE.FindStringSubmatch(resp)
    if len(m) != 7 { return nil, resp, fmt.Errorf("PASV parse error: %s", strings.TrimSpace(resp)) }
    ip := fmt.Sprintf("%s.%s.%s.%s", m[1], m[2], m[3], m[4])
    p1, _ := strconv.Atoi(m[5]); p2, _ := strconv.Atoi(m[6])
    dc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, p1*256+p2))
    return dc, resp, err
}

func (c *Client) REPL() {
    in := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("ftp> ")
        line, _ := in.ReadString('\n')
        line = strings.TrimSpace(line)
        if line == "" { continue }
        parts := strings.SplitN(line, " ", 3)
        cmd := strings.ToLower(parts[0])

        switch cmd {
        case "ls":
            dc, _, err := c.pasvDial(); if err != nil { fmt.Println(err); continue }
            fmt.Print(strings.TrimSpace(c.send("LIST")), "\n")
            io.Copy(os.Stdout, dc)
            dc.Close()
            fmt.Print(strings.TrimSpace(readLine(c.rw)), "\n")
        case "cd":
            if len(parts) < 2 { fmt.Println("usage: cd <dir>"); continue }
            arg := parts[1]
            resp := strings.TrimSpace(c.send("CWD "+arg))
            fmt.Println(resp)
            if strings.HasPrefix(resp, "250") {
                pwd := strings.TrimSpace(c.send("PWD"))
                fmt.Println(pwd)
            }
        case "pwd":
            resp := strings.TrimSpace(c.send("PWD"))
            fmt.Println(resp)
        case "rm":
            if len(parts) < 2 { fmt.Println("usage: rm <remote>"); continue }
            arg := parts[1]
            resp := strings.TrimSpace(c.send("DELE "+arg))
            fmt.Println(resp)
        case "mkdir":
            if len(parts) < 2 { fmt.Println("usage: mkdir <dir>"); continue }
            arg := parts[1]
            resp := strings.TrimSpace(c.send("MKD "+arg))
            fmt.Println(resp)
        case "mv":
            if len(parts) < 3 { fmt.Println("usage: mv <old> <new>"); continue }
            src := parts[1]; dst := parts[2]
            r := strings.TrimSpace(c.send("RNFR "+src))
            fmt.Println(r)
            if strings.HasPrefix(r, "350") { fmt.Println(strings.TrimSpace(c.send("RNTO "+dst))) }
        case "get":
            if len(parts) < 2 { fmt.Println("usage: get <remote>"); continue }
            arg := parts[1]
            dc, _, err := c.pasvDial(); if err != nil { fmt.Println(err); continue }
            fmt.Print(strings.TrimSpace(c.send("RETR "+arg)), "\n")
            f, _ := os.Create(filepath.Base(arg))
            io.Copy(f, dc)
            f.Close(); dc.Close()
            fmt.Print(strings.TrimSpace(readLine(c.rw)), "\n")
        case "put":
            if len(parts) < 2 { fmt.Println("usage: put <local>"); continue }
            arg := parts[1]
            dc, _, err := c.pasvDial(); if err != nil { fmt.Println(err); continue }
            fmt.Print(strings.TrimSpace(c.send("STOR "+filepath.Base(arg))), "\n")
            f, err := os.Open(arg); if err != nil { fmt.Println(err); dc.Close(); continue }
            io.Copy(dc, f)
            f.Close(); dc.Close()
            fmt.Print(strings.TrimSpace(readLine(c.rw)), "\n")
        case "quit", "exit":
            fmt.Print(strings.TrimSpace(c.send("QUIT")), "\n")
            return
        case "help":
            fmt.Println("Commands: ls | cd <dir> | pwd | get <file> | put <file> | rm <file> | mkdir <dir> | mv <src> <dst> | quit")
        default:
            fmt.Println("unknown command — try: ls, cd, pwd, get, put, rm, mkdir, mv, quit")
        }
    }
}

func readLine(rw *bufio.ReadWriter) string { s, _ := rw.ReadString('\n'); return s }
