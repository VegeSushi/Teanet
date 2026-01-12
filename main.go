package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":2323")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Telnet server listening on port 2323")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	asciiArt := `
             ;,'
     _o_    ;:;'
 ,-.'---` + "`" + `.__ ;
((j` + "`" + `=====',-'
 ` + "`" + `-\     /
    ` + "`" + `-=-'
`
	writer.WriteString(asciiArt + "\r\n")
	writer.WriteString("Welcome to the Teanet!\r\n")
	writer.WriteString("Type 'help' to see commands.\r\n\r\n")
	writer.Flush()

	for {
		writer.WriteString("> ")
		writer.Flush()

		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		if !handleCommand(command, writer) {
			return
		}
	}
}

func handleCommand(cmd string, w *bufio.Writer) bool {
	args := strings.Split(cmd, " ")

	switch args[0] {
	case "help":
		w.WriteString("Available commands:\r\n")
		w.WriteString("  help        Show this help\r\n")
		w.WriteString("  wiki <term> Wikipedia lookup\r\n")
		w.WriteString("  quit        Disconnect\r\n")

	case "wiki":
		if len(args) < 2 {
			w.WriteString("Usage: wiki <search term>\r\n")
		} else {
			term := strings.Join(args[1:], " ")
			result, err := fetchWikipedia(term)
			if err != nil {
				w.WriteString("Error fetching Wikipedia: " + err.Error() + "\r\n")
			} else {
				w.WriteString(result + "\r\n")
			}
		}

	case "quit", "exit":
		w.WriteString("Goodbye!\r\n")
		w.Flush()
		return false

	default:
		w.WriteString("Unknown command. Type 'help'.\r\n")
	}

	w.Flush()
	return true
}


func fetchWikipedia(term string) (string, error) {
	url := "https://en.wikipedia.org/api/rest_v1/page/summary/" + strings.ReplaceAll(term, " ", "_")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "TeanetBot/1.0 (sushiware@gmail.com)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Wikipedia returned status %d", resp.StatusCode)
	}

	var data struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Extract     string `json:"extract"`
		ContentURLs struct {
			Desktop struct {
				Page string `json:"page"`
			} `json:"desktop"`
		} `json:"content_urls"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	result := fmt.Sprintf("%s - %s\n\n%s\n\nRead more: %s",
		data.Title, data.Description, data.Extract, data.ContentURLs.Desktop.Page)

	return result, nil
}