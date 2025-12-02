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
    ctrl     net.Conn
    rw       *bufio.ReadWriter
    localCwd string
}

func Dial(addr string) (*Client, error) {
    c, err := net.Dial("tcp", addr)
    if err != nil {
        return nil, err
    }
    rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))

    _, _ = rw.ReadString('\n')

    wd, _ := os.Getwd()

    return &Client{
        ctrl:     c,
        rw:       rw,
        localCwd: wd,
    }, nil
}

func (c *Client) Close() error { return c.ctrl.Close() }

func (c *Client) send(cmd string) string {
    fmt.Fprintf(c.rw, "%s\r\n", cmd)
    _ = c.rw.Flush()
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
        _ = strings.TrimSpace(c.send("USER anonymous"))
        _ = strings.TrimSpace(c.send("PASS guest"))
        _ = strings.TrimSpace(c.send("TYPE I"))
        return nil
    }

    if pass == "" {
        in := bufio.NewReader(os.Stdin)
        fmt.Print("Password: ")
        p, _ := in.ReadString('\n')
        pass = strings.TrimSpace(p)
    }

    r1 := strings.TrimSpace(c.send("USER " + user))
    if strings.HasPrefix(r1, "530") {
        return fmt.Errorf(r1)
    }
    r2 := strings.TrimSpace(c.send("PASS " + pass))
    if !strings.HasPrefix(r2, "230") {
        return fmt.Errorf(r2)
    }
    _ = strings.TrimSpace(c.send("TYPE I"))
    return nil
}

var pasvRE = regexp.MustCompile(`\((\d+),(\d+),(\d+),(\d+),(\d+),(\d+)\)`)

func (c *Client) pasvDial() (net.Conn, string, error) {
    resp := c.send("PASV")
    m := pasvRE.FindStringSubmatch(resp)
    if len(m) != 7 {
        return nil, resp, fmt.Errorf("PASV parse error: %s", strings.TrimSpace(resp))
    }
    ip := fmt.Sprintf("%s.%s.%s.%s", m[1], m[2], m[3], m[4])
    p1, _ := strconv.Atoi(m[5])
    p2, _ := strconv.Atoi(m[6])
    dc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, p1*256+p2))
    return dc, resp, err
}

func (c *Client) REPL() {
    in := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("ftp> ")
        line, _ := in.ReadString('\n')
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        parts := strings.SplitN(line, " ", 3)
        cmd := strings.ToLower(parts[0])

        switch cmd {
        case "ls":
            dc, _, err := c.pasvDial()
            if err != nil {
                fmt.Println(err)
                continue
            }
            fmt.Print(strings.TrimSpace(c.send("LIST")), "\n")
            _, _ = io.Copy(os.Stdout, dc)
            _ = dc.Close()
            fmt.Print(strings.TrimSpace(readLine(c.rw)), "\n")

        case "lls":
            entries, err := os.ReadDir(c.localCwd)
            if err != nil {
                fmt.Println("lls error:", err)
                continue
            }
            for _, e := range entries {
                fmt.Println(e.Name())
            }

        case "cd":
            if len(parts) < 2 {
                fmt.Println("usage: cd <dir>")
                continue
            }
            arg := parts[1]
            resp := strings.TrimSpace(c.send("CWD " + arg))
            fmt.Println(resp)
            if strings.HasPrefix(resp, "250") {
                pwd := strings.TrimSpace(c.send("PWD"))
                fmt.Println(pwd)
            }

        case "pwd":
            resp := strings.TrimSpace(c.send("PWD"))
            fmt.Println(resp)

        case "rm":
            if len(parts) < 2 {
                fmt.Println("usage: rm <remote>")
                continue
            }
            arg := parts[1]
            resp := strings.TrimSpace(c.send("DELE " + arg))
            fmt.Println(resp)

        case "mkdir":
            if len(parts) < 2 {
                fmt.Println("usage: mkdir <dir>")
                continue
            }
            arg := parts[1]
            resp := strings.TrimSpace(c.send("MKD " + arg))
            fmt.Println(resp)

        case "mv":
            if len(parts) < 3 {
                fmt.Println("usage: mv <old> <new>")
                continue
            }
            src := parts[1]
            dst := parts[2]
            r := strings.TrimSpace(c.send("RNFR " + src))
            fmt.Println(r)
            if strings.HasPrefix(r, "350") {
                fmt.Println(strings.TrimSpace(c.send("RNTO " + dst)))
            }

        case "get":
            if len(parts) < 2 {
                fmt.Println("usage: get <remote>")
                continue
            }
            remote := parts[1]

            dc, _, err := c.pasvDial()
            if err != nil {
                fmt.Println(err)
                continue
            }
            fmt.Print(strings.TrimSpace(c.send("RETR "+remote)), "\n")

            localName := filepath.Base(remote)
            localPath := filepath.Join(c.localCwd, localName)

            f, err := os.Create(localPath)
            if err != nil {
                fmt.Println("create local file error:", err)
                _ = dc.Close()
                _ = readLine(c.rw)
                continue
            }
            _, _ = io.Copy(f, dc)
            _ = f.Close()
            _ = dc.Close()
            fmt.Print(strings.TrimSpace(readLine(c.rw)), "\n")

        case "put":
            if len(parts) < 2 {
                fmt.Println("usage: put <local>")
                continue
            }
            arg := parts[1]

            localPath := arg
            if !filepath.IsAbs(localPath) {
                localPath = filepath.Join(c.localCwd, arg)
            }
            localPath = filepath.Clean(localPath)

            dc, _, err := c.pasvDial()
            if err != nil {
                fmt.Println(err)
                continue
            }
            fmt.Print(strings.TrimSpace(c.send("STOR "+filepath.Base(arg))), "\n")

            f, err := os.Open(localPath)
            if err != nil {
                fmt.Println("open local file error:", err)
                _ = dc.Close()
                _ = readLine(c.rw)
                continue
            }
            _, _ = io.Copy(dc, f)
            _ = f.Close()
            _ = dc.Close()
            fmt.Print(strings.TrimSpace(readLine(c.rw)), "\n")

        case "lpwd":
            fmt.Println("Local:", c.localCwd)

        case "lcd":
            if len(parts) < 2 {
                fmt.Println("usage: lcd <dir>")
                continue
            }
            newLocal := parts[1]
            if !filepath.IsAbs(newLocal) {
                newLocal = filepath.Join(c.localCwd, newLocal)
            }
            newLocal = filepath.Clean(newLocal)
            fi, err := os.Stat(newLocal)
            if err != nil || !fi.IsDir() {
                fmt.Println("lcd: not a directory:", newLocal)
                continue
            }
            c.localCwd = newLocal
            fmt.Println("Local dir:", c.localCwd)

        case "quit", "exit":
            fmt.Print(strings.TrimSpace(c.send("QUIT")), "\n")
            return

        case "help":
            fmt.Println("Commands:")
            fmt.Println("  ls | cd <dir> | pwd")
            fmt.Println("  get <file> | put <file> | rm <file> | mkdir <dir> | mv <src> <dst>")
            fmt.Println("  lpwd | lcd <dir>  (local directory on client)")
            fmt.Println("  quit")

        default:
            fmt.Println("unknown command — try: help")
        }
    }
}

func readLine(rw *bufio.ReadWriter) string {
    s, _ := rw.ReadString('\n')
    return s
}
