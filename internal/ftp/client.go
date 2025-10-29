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
    // read banner
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

func (c *Client) LoginAnonymous() error {
    c.send("USER anonymous")
    c.send("PASS guest")
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

// REPL: ls, get <file>, put <file>, quit
func (c *Client) REPL() {
    in := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("ftp> ")
        line, _ := in.ReadString('\n')
        line = strings.TrimSpace(line)
        if line == "" { continue }
        parts := strings.SplitN(line, " ", 2)
        cmd := strings.ToLower(parts[0])
        arg := ""; if len(parts) > 1 { arg = parts[1] }

        switch cmd {
        case "ls":
            dc, _, err := c.pasvDial(); if err != nil { fmt.Println(err); continue }
            fmt.Print(strings.TrimSpace(c.send("LIST")), "\n") // 150
            io.Copy(os.Stdout, dc)
            dc.Close()
            fmt.Print(strings.TrimSpace(readLine(c.rw)), "\n") // 226
        case "cd":
            if arg == "" { fmt.Println("usage: cd <dir>"); continue }
            resp := strings.TrimSpace(c.send("CWD "+arg))
            fmt.Println(resp)
            if strings.HasPrefix(resp, "250") {
                pwd := strings.TrimSpace(c.send("PWD"))
                fmt.Println(pwd)
            }
        case "get":
            if arg == "" { fmt.Println("usage: get <remote>"); continue }
            dc, _, err := c.pasvDial(); if err != nil { fmt.Println(err); continue }
            fmt.Print(strings.TrimSpace(c.send("RETR "+arg)), "\n")
            f, _ := os.Create(filepath.Base(arg))
            io.Copy(f, dc)
            f.Close(); dc.Close()
            fmt.Print(strings.TrimSpace(readLine(c.rw)), "\n")
        case "put":
            if arg == "" { fmt.Println("usage: put <local>"); continue }
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
            fmt.Println("Commands: ls | get <file> | put <file> | quit")
        default:
            fmt.Println("unknown command — try: ls, get, put, quit")
        }
    }
}

func readLine(rw *bufio.ReadWriter) string { s, _ := rw.ReadString('\n'); return s }
