package glutton

import (
	"bufio"
	"math/rand"
	"net"
	"regexp"
	"strings"

	"github.com/kung-foo/freki"
)

// Mirai botnet  - https://github.com/CymmetriaResearch/MTPot/blob/master/mirai_conf.json
// Hajime botnet - https://security.rapiditynetworks.com/publications/2016-10-16/hajime.pdf
var miraiCom = map[string][]string{
	"ps":                                              []string{"1 pts/21   00:00:00 init"},
	"cat /proc/mounts":                                []string{"rootfs / rootfs rw 0 0\r\n/dev/root / ext2 rw,relatime,errors=continue 0 0\r\nproc /proc proc rw,relatime 0 0\r\nsysfs /sys sysfs rw,relatime 0 0\r\nudev /dev tmpfs rw,relatime 0 0\r\ndevpts /dev/pts devpts rw,relatime,mode=600,ptmxmode=000 0 0\r\n/dev/mtdblock1 /home/hik jffs2 rw,relatime 0 0\r\ntmpfs /run tmpfs rw,nosuid,noexec,relatime,size=3231524k,mode=755 0 0\r\n"},
	"(cat .s || cp /bin/echo .s)": 			   []string{"cat: .s: No such file or directory"},
	"nc": 						   []string{"nc: command not found"},
	"wget": 					   []string{"wget: missing URL"},
	"(dd bs=52 count=1 if=.s || cat .s)": 		   []string{"\x7f\x45\x4c\x46\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00\x01\x00\x00\x00\xbc\x14\x01\x00\x34\x00\x00\x00"},
	"sh": 					           []string{"$"},
	"echo -e \\x6b\\x61\\x6d\\x69/dev > /dev/.nippon": []string{""},
	"cat /dev/.nippon":                                []string{"kami/dev"},
	"rm /dev/.nippon":                                 []string{""},
	"echo -e \\x6b\\x61\\x6d\\x69/run > /run/.nippon": []string{""},
	"cat /run/.nippon":                                []string{"kami/run"},
	"rm /run/.nippon":                                 []string{""},
	"cat /bin/sh":                                     []string{""},
	"/bin/busybox ps":                                              []string{"1 pts/21   00:00:00 init"},
	"/bin/busybox cat /proc/mounts":                                []string{"tmpfs /run tmpfs rw,nosuid,noexec,relatime,size=3231524k,mode=755 0 0"},
	"/bin/busybox echo -e \\x6b\\x61\\x6d\\x69/dev > /dev/.nippon": []string{""},
	"/bin/busybox cat /dev/.nippon":                                []string{"kami/dev"},
	"/bin/busybox rm /dev/.nippon":                                 []string{""},
	"/bin/busybox echo -e \\x6b\\x61\\x6d\\x69/run > /run/.nippon": []string{""},
	"/bin/busybox cat /run/.nippon":                                []string{"kami/run"},
	"/bin/busybox rm /run/.nippon":                                 []string{""},
	"/bin/busybox cat /bin/sh":                                     []string{""},
	"/bin/busybox cat /bin/echo":                                   []string{"/bin/busybox cat /bin/echo\r\n\x7f\x45\x4c\x46\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00\x01\x00\x00\x00\x6c\xb9\x00\x00\x34\x00\x00\x00"},
        "rm /dev/.human":                                               []string{"rm: can't remove '/.t': No such file or directory\r\nrm: can't remove '/.sh': No such file or directory\r\nrm: can't remove '/.human': No such file or directory\r\ncd /dev"},
}

func writeMsg(conn net.Conn, msg string, g *Glutton) error {
	_, err := conn.Write([]byte(msg))
	g.logger.Infof("[telnet  ] send: %q", msg)
	md := g.processor.Connections.GetByFlow(freki.NewConnKeyFromNetConn(conn))
	g.producer.LogHTTP(conn, md, msg, "write")
	return err
}

func readMsg(conn net.Conn, g *Glutton) (msg string, err error) {
	msg, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	g.logger.Infof("[telnet  ] recv: %q", msg)
	md := g.processor.Connections.GetByFlow(freki.NewConnKeyFromNetConn(conn))
	g.producer.LogHTTP(conn, md, msg, "read")
	return msg, err
}

// HandleTelnet handles telnet communication on a connection
func (g *Glutton) HandleTelnet(conn net.Conn) {
	defer conn.Close()

	// TODO (glaslos): Add device banner

	// telnet window size negotiation response
	writeMsg(conn, "\xff\xfd\x18\xff\xfd\x20\xff\xfd\x23\xff\xfd\x27", g)

	// User name prompt
	writeMsg(conn, "Username: ", g)
	_, err := readMsg(conn, g)
	if err != nil {
		g.logger.Errorf("[telnet  ] %v", err)
		return
	}
	writeMsg(conn, "Password: ", g)
	_, err = readMsg(conn, g)
	if err != nil {
		g.logger.Errorf("[telnet  ] %v", err)
		return
	}

	writeMsg(conn, "welcome\r\n> ", g)
	for {
		msg, err := readMsg(conn, g)
		if err != nil {
			g.logger.Errorf("[telnet  ] %v", err)
			return
		}
		for _, cmd := range strings.Split(msg, ";") {
			if strings.TrimRight(cmd, "") == " rm /dev/.t" {
				continue
			}
			if strings.TrimRight(cmd, "\r\n") == " rm /dev/.sh" {
				continue
			}
			if strings.TrimRight(cmd, "\r\n") == "cd /dev/" {
				writeMsg(conn, "ECCHI: applet not found\r\n", g)
				writeMsg(conn, "\r\nBusyBox v1.16.1 (2014-03-04 16:00:18 CST) built-it shell (ash)\r\nEnter 'help' for a list of built-in commands.\r\n", g)
				continue
			}

			if resp := miraiCom[strings.TrimSpace(cmd)]; len(resp) > 0 {
				writeMsg(conn, resp[rand.Intn(len(resp))]+"\r\n", g)
			} else {
				// /bin/busybox YDKBI
				re := regexp.MustCompile(`\/bin\/busybox (?P<applet>[A-Z]+)`)
				match := re.FindStringSubmatch(cmd)
				if len(match) > 1 {
					writeMsg(conn, match[1]+": applet not found\r\n", g)
					writeMsg(conn, "BusyBox v1.16.1 (2014-03-04 16:00:18 CST) built-in shell (ash)\r\nEnter 'help' for a list of built-in commands.\r\n", g)
				}
			}
		}
		writeMsg(conn, "> ", g)
	}
}
