package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/agl/xmpp"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
)

const (
	defaultHost = "localhost"
	defaultPort = "8041"
)

var con config

type config struct {
	Host string
	Port string
}

type session struct {
	talk *xmpp.Conn
	ws   *websocket.Conn
}

type message struct {
	Type string
	Data map[string]string
}

type rosterMessage struct {
	Type   string
	Roster []xmpp.RosterEntry
}

type account struct {
	userName string
	password string
	server   string
}

func chatJsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var indexTemplate = template.Must(template.ParseFiles("html/chat.js"))
		if err := indexTemplate.Execute(w, con); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var account account
		account.userName = r.FormValue("username")
		account.password = r.FormValue("password")
		account.server = r.FormValue("server")

		if account.userName == "" || account.password == "" {
			fmt.Println("vacant")
		}

		var chatTemplate = template.Must(template.ParseFiles("html/chat.html"))
		if err := chatTemplate.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		var indexTemplate = template.Must(template.ParseFiles("html/chat.html"))
		if err := indexTemplate.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func Echo(ws *websocket.Conn) {
	fmt.Println("socket open")
	var recMes message

	err := websocket.JSON.Receive(ws, &recMes)
	if err != nil {
		fmt.Println("Can't receive message")
	}

	if recMes.Type != "login" {
		fmt.Println("not login message")
		return
	}

	userName := strings.Split(recMes.Data["UserName"], "@")
	if len(userName) < 2 {
		userName = []string{userName[0], ""}
	}

	xmppConfig := xmpp.Config{nil, nil, nil, nil, false, false, false, []byte("")}
	talk, err := xmpp.Dial(recMes.Data["Server"], userName[0], userName[1], recMes.Data["Password"], &xmppConfig)
	if err != nil {
		fmt.Println("login error")
		return
	}

	senMes := message{"login", nil}

	if err := websocket.JSON.Send(ws, senMes); err != nil {
		fmt.Println("Can't send")
		return
	}

	fmt.Println("Ready")

	rosterReply, _, err := talk.RequestRoster()
	if err != nil {
		fmt.Println("Can't roster")
	}

	talk.SignalPresence("")

	ses := session{
		talk: talk,
		ws:   ws,
	}

	stanzaChan := make(chan xmpp.Stanza)
	go ses.receiveMessage(stanzaChan)

	webSocketChan := make(chan string)
	go ses.receiveWebSocket(webSocketChan)

	for {
		select {
		case rosterStanza, ok := <-rosterReply:
			var roster []xmpp.RosterEntry
			if !ok {
				fmt.Println("fail to read roaster")
				return
			}
			if roster, err = xmpp.ParseRoster(rosterStanza); err != nil {
				fmt.Println("fail to parse roaster")
				return
			}
			if err := websocket.JSON.Send(ws, rosterMessage{"roster", roster}); err != nil {
				fmt.Println("Can't send")
				return
			}
			fmt.Println("Roster information sent to the client")
		case rawStanza, ok := <-stanzaChan:
			if !ok {
				fmt.Println("stanzaChan receive failed")
				return
			}
			switch stanza := rawStanza.Value.(type) {
			case *xmpp.ClientMessage:
				fmt.Println("Message comming")
				if stanza.Body == "" {
					break
				}
				chatMes := message{"chat", map[string]string{"Remote": xmpp.RemoveResourceFromJid(stanza.From), "Text": stanza.Body}}

				if err := websocket.JSON.Send(ws, chatMes); err != nil {
					fmt.Println("Can't send")
					break
				}
			case *xmpp.ClientPresence:
				fmt.Println("ClientPresence comming")
				fmt.Println(stanza.From, stanza.Type)

				var status string
				if len(stanza.Show) > 0 {
					status = stanza.Show
				} else {
					status = "available"
				}

				preMes := message{"presence", map[string]string{"Remote": xmpp.RemoveResourceFromJid(stanza.From), "Mode": status}}
				if err := websocket.JSON.Send(ws, preMes); err != nil {
					fmt.Println("Can't send")
					return
				}
			}
		case receivedMessage, ok := <-webSocketChan:
			if !ok {
				fmt.Println("webSocketChan receive failed")
				return
			}
			var sendMes message
			if err = json.Unmarshal([]byte(receivedMessage), &sendMes); err != nil {
				fmt.Println("Can't receive")
			} else {

				if err = talk.Send(sendMes.Data["Remote"], sendMes.Data["Text"]); err != nil {
					fmt.Println("failed to send")
				}
				fmt.Println("Message sent" + sendMes.Data["Remote"] + sendMes.Data["Text"])
			}
		}
	}
	fmt.Println("function ends")
}

func (ses *session) receiveMessage(stanzaChan chan<- xmpp.Stanza) {
	defer close(stanzaChan)
	for {
		fmt.Println("xmpp reveicer waiting")
		stanza, err := ses.talk.Next()
		if err != nil {
			log.Fatal(err)
			fmt.Println("goroutine reveiveMessage closed")
			return
		}
		stanzaChan <- stanza
	}
}

func (ses *session) receiveWebSocket(webSocketChan chan<- string) {
	defer close(webSocketChan)
	for {
		fmt.Println("xmpp sender waiting")
		var receivedMessage string
		if err := websocket.Message.Receive(ses.ws, &receivedMessage); err != nil {
			fmt.Println("goroutine reveiveWebSocket closed")
			return
		}
		webSocketChan <- receivedMessage
	}
}

func main() {
	file, err := ioutil.ReadFile("config")
	if err != nil {
		fmt.Println("Host set to default (" + defaultHost + ")")
		fmt.Println("port set to default (" + defaultPort + ")")
		con.Host = defaultHost
		con.Port = defaultPort
	} else {
		if err := json.Unmarshal(file, &con); err != nil {
			fmt.Println("Host set to default (" + defaultHost + ")")
			fmt.Println("port set to default (" + defaultPort + ")")
			con.Host = defaultHost
			con.Port = defaultPort
		}
	}

	http.Handle("/", http.FileServer(http.Dir("html")))
	http.HandleFunc("/chat.html", chatHandler)
	http.HandleFunc("/chat.js", chatJsHandler)
	http.Handle("/websocket", websocket.Handler(Echo))
	http.ListenAndServe(":"+con.Port, nil)
}
