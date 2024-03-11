package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/txthinking/socks5"
)

func RaiseLimits() error {
	var l syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &l); err != nil {
		return err
	}
	if runtime.GOOS == "darwin" && l.Cur < 10240 {
		l.Cur = 10240
	}
	if runtime.GOOS != "darwin" && l.Cur < 60000 {
		if l.Max < 60000 {
			l.Max = 60000 // with CAP_SYS_RESOURCE capability
		}
		l.Cur = l.Max
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &l); err != nil {
		return err
	}
	return nil
}

func main() {
	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}

	appName := filepath.Base(exec)

	flag.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintf(output, "Usage: %s [...OPTIONS]\n\n", appName)
		fmt.Fprintln(output, "Options:")
		flag.PrintDefaults()
	}

	listen := flag.String("listen", "0.0.0.0:1080", "Socks5 server listen address, like: :1080 or 1.2.3.4:1080")
	socks5ServerIP := flag.String("socks5ServerIP", "", "Only if your socks5 server IP is different from listen IP")
	username := flag.String("username", "", "User name, optional")
	password := flag.String("password", "", "Password, optional")
	tcpTimeout := flag.Int("tcpTimeout", 0, "Connection deadline time (s)")
	udpTimeout := flag.Int("udpTimeout", 60, "Connection deadline time (s)")
	limitUDP := flag.Bool("limitUDP", false, "The server MAY use this information to limit access to the UDP association. This usually causes connection failures in a NAT environment, where most clients are.")

	flag.Parse()

	if *listen == "" {
		flag.Usage()
		return
	}

	host, _, err := net.SplitHostPort(*listen)
	if err != nil {
		log.Fatalln(err)
	}

	if host == "" && *socks5ServerIP == "" {
		fmt.Println("socks5 server requires a clear IP for UDP, only port is not enough. You may use public IP or lan IP or other, we can not decide for you")
	}

	var ip string

	if host != "" {
		ip = host
	}

	if *socks5ServerIP != "" {
		ip = *socks5ServerIP
	}

	if err := RaiseLimits(); err != nil {
		log.Fatalln(err)
	}

	server, err := socks5.NewClassicServer(*listen, ip, *username, *password, *tcpTimeout, *udpTimeout)
	if err != nil {
		log.Fatalln(err)
	}

	server.LimitUDP = *limitUDP
	server.ListenAndServe(nil)
}
