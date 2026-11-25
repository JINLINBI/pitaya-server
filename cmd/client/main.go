package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/v3/pkg/client"
	"github.com/topfreegames/pitaya/v3/pkg/conn/message"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:3250", "connector address host:port")
	user := flag.String("user", "", "login username")
	flag.Parse()

	username := strings.TrimSpace(*user)
	if username == "" {
		fmt.Fprintln(os.Stderr, "usage: client --addr host:port --user name")
		os.Exit(2)
	}

	c := client.New(logrus.ErrorLevel, 10*time.Second)
	if err := c.ConnectTo(*addr); err != nil {
		fmt.Fprintf(os.Stderr, "connect %s: %v\n", *addr, err)
		os.Exit(1)
	}
	defer c.Disconnect()

	loginBody, _ := json.Marshal(map[string]string{"username": username})
	if err := request(c, "connector.login", loginBody); err != nil {
		fmt.Fprintf(os.Stderr, "login: %v\n", err)
		os.Exit(1)
	}
	if err := request(c, "room.join", nil); err != nil {
		fmt.Fprintf(os.Stderr, "join: %v\n", err)
		os.Exit(1)
	}

	go drainPushes(c)

	fmt.Printf("logged in as %s via %s\n", username, *addr)
	fmt.Println("type a message and press Enter; Ctrl+D to quit")

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}
		body, _ := json.Marshal(map[string]string{
			"name":    username,
			"content": text,
		})
		if err := c.SendNotify("room.message", body); err != nil {
			fmt.Fprintf(os.Stderr, "send: %v\n", err)
		}
	}
}

func request(c *client.Client, route string, data []byte) error {
	if data == nil {
		data = []byte("{}")
	}
	id, err := c.SendRequest(route, data)
	if err != nil {
		return err
	}
	deadline := time.After(10 * time.Second)
	for {
		select {
		case <-deadline:
			return fmt.Errorf("timeout waiting for %s", route)
		case msg := <-c.MsgChannel():
			if msg == nil {
				continue
			}
			if msg.Type == message.Response && msg.ID == id {
				if msg.Err {
					return fmt.Errorf("%s: %s", route, string(msg.Data))
				}
				fmt.Printf("< %s %s\n", route, string(msg.Data))
				return nil
			}
			printPush(msg)
		}
	}
}

func drainPushes(c *client.Client) {
	for msg := range c.MsgChannel() {
		if msg == nil {
			continue
		}
		printPush(msg)
	}
}

func printPush(msg *message.Message) {
	switch msg.Route {
	case "onMessage":
		var m struct {
			Name    string `json:"name"`
			Content string `json:"content"`
		}
		_ = json.Unmarshal(msg.Data, &m)
		fmt.Printf("[%s] %s\n", m.Name, m.Content)
	case "onNewUser":
		var m struct {
			Content string `json:"content"`
		}
		_ = json.Unmarshal(msg.Data, &m)
		fmt.Printf("* %s\n", m.Content)
	case "onMembers":
		fmt.Printf("* members %s\n", string(msg.Data))
	default:
		if msg.Type == message.Push || msg.Type == message.Notify {
			fmt.Printf("< push %s %s\n", msg.Route, string(msg.Data))
		}
	}
}
