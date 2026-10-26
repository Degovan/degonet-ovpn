package api

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrClientNotFound = errors.New("client not found")

// OpenVPN 2.6 menambahkan prefix protokol di Real Address, misal "tcp4-server:1.2.3.4:5678".
var protoPrefix = regexp.MustCompile(`^(tcp|udp)[46]?(-server|-client)?:`)

// Management berbicara dengan management interface OpenVPN (server.conf: management localhost 7505).
type Management struct {
	Addr string

	// Interface management OpenVPN hanya melayani satu koneksi sekaligus.
	mu sync.Mutex
}

type Client struct {
	Username       string `json:"username"`
	RealAddress    string `json:"real_address"`
	VirtualAddress string `json:"virtual_address"`
	BytesReceived  int64  `json:"bytes_received"`
	BytesSent      int64  `json:"bytes_sent"`
	ConnectedSince int64  `json:"connected_since"`
}

// command mengirim satu perintah dan mengembalikan baris balasan sampai baris terakhir
// yang memenuhi done (baris notifikasi async ">..." diabaikan).
func (m *Management) command(cmd string, done func(line string) bool) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := net.DialTimeout("tcp", m.Addr, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	if _, err := fmt.Fprintf(conn, "%s\n", cmd); err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.HasPrefix(line, ">") {
			continue
		}
		lines = append(lines, line)
		if done(line) {
			fmt.Fprint(conn, "quit\n")
			return lines, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, errors.New("management connection closed unexpectedly")
}

func (m *Management) Clients() ([]Client, error) {
	lines, err := m.command("status 3", func(line string) bool { return line == "END" })
	if err != nil {
		return nil, err
	}

	var header map[string]int
	clients := []Client{}
	for _, line := range lines {
		fields := strings.Split(line, "\t")
		switch {
		case len(fields) > 1 && fields[0] == "HEADER" && fields[1] == "CLIENT_LIST":
			header = make(map[string]int, len(fields)-1)
			for i, name := range fields[1:] {
				header[name] = i
			}
		case fields[0] == "CLIENT_LIST" && header != nil:
			get := func(name string) string {
				if i, ok := header[name]; ok && i < len(fields) {
					return fields[i]
				}
				return ""
			}
			num := func(name string) int64 {
				n, _ := strconv.ParseInt(get(name), 10, 64)
				return n
			}
			clients = append(clients, Client{
				Username:       get("Common Name"),
				RealAddress:    protoPrefix.ReplaceAllString(get("Real Address"), ""),
				VirtualAddress: get("Virtual Address"),
				BytesReceived:  num("Bytes Received"),
				BytesSent:      num("Bytes Sent"),
				ConnectedSince: num("Connected Since (time_t)"),
			})
		}
	}
	return clients, nil
}

func (m *Management) Kill(username string) error {
	quoted := strconv.Quote(username)
	lines, err := m.command("kill "+quoted, func(line string) bool {
		return strings.HasPrefix(line, "SUCCESS:") || strings.HasPrefix(line, "ERROR:")
	})
	if err != nil {
		return err
	}
	last := lines[len(lines)-1]
	if strings.HasPrefix(last, "ERROR:") {
		if strings.Contains(last, "not found") {
			return ErrClientNotFound
		}
		return errors.New(strings.TrimSpace(strings.TrimPrefix(last, "ERROR:")))
	}
	return nil
}
