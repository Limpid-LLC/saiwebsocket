package main

import (
	"fmt"
	"net/http"
	"strings"
	"encoding/json"
	//~ "bytes"
	"io/ioutil"
	"../github.com/tkanos/gonfig"
	"github.com/gorilla/websocket"
	"time"
)

var clients = make(map[*websocket.Conn]string) // connected clients
var allowsTokens = make(map[string]bool)	   // allows clients
var broadcast = make(chan string)            // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    (*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func api(w http.ResponseWriter, r *http.Request ) {
	setupResponse(&w,r)
	mes := ""
	r.ParseForm()
	method := strings.Join(r.Form["method"],"")
	switch method {
		case "broadcast": {
			mes = strings.Join(r.Form["message"],"")
			broadcast <- string(mes)
			fmt.Println("broadcast:",mes)			
		}
		case "get_clients":{
			theclients := make(map[string]string);
			for key, value := range clients {
				s := fmt.Sprintf("%s", key.RemoteAddr())
				theclients[s] = value
			}			
			jsonString, err := json.Marshal(theclients)
			if err != nil {
				w.Write([]byte("{\"error\":\"json encoding error\"}"))
			} else {
				w.Write([]byte(jsonString))
			}
			fmt.Println("clients list:",clients)
		}
		case "registerToken": {
			t := strings.Join(r.Form["token"],"")
			if len(t) >= 1 { allowsTokens[t] = true; }
		}	
		case "unregisterToken": {
			t := strings.Join(r.Form["token"],"")
			if len(t) >= 1 { allowsTokens[t] = false; }
		}	
		case "registeredTokenList": {
			jsonString, err := json.Marshal(allowsTokens)
			if err != nil {
				w.Write([]byte("{\"error\":\"json encoding error\"}"))
			} else {
				w.Write([]byte(jsonString))
			}
		}	
	}
}

type saiwebsocketconfig struct {
	Host string
	Port string
	Origin string
	Responseheaders string
	AllowUnregisteredClients string
	RegisteredTokensUrl string
}

var websocketconfig saiwebsocketconfig
func main() {
 	config_err := gonfig.GetConf("saiwebsocket.config", &websocketconfig)
	if config_err != nil {
		fmt.Println("Config missed!! ")
		panic(config_err)
	}
	fmt.Println(websocketconfig)
	// Reading preregistered allows tokens
	if len(websocketconfig.RegisteredTokensUrl) > 0 {
		resp, err := http.Get(websocketconfig.RegisteredTokensUrl)
		if err != nil {
			fmt.Println("Corrupted URL. Registered tokens can not be imported")
		} else { 
			
			defer resp.Body.Close() 

			tokensJsonString, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Corrupted URL data. Registered tokens can not be imported")
			}
			if len(tokensJsonString) > 0 {
				err := json.Unmarshal(tokensJsonString, &allowsTokens)
				if err != nil {
					fmt.Println("Corrupted JSON data. Registered tokens can not be imported \n Url output example {\"abc1\":true,\"abc2\":true,\"abc3\":true,\"group/item1\":true,\"group/item2\":true}")
				}			
			}
		}
	}
	if len(allowsTokens) > 0 {
		fmt.Println("Registered tokens imported ",allowsTokens)
	}
	// Configure routes
	http.HandleFunc("/", api)
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming messages
	go handleMessages()

	// Start the server 
	fmt.Println("http server started on "+websocketconfig.Host+":"+websocketconfig.Port)
	err := http.ListenAndServe(websocketconfig.Host+":"+websocketconfig.Port, nil)
	//~ err := http.ListenAndServeTLS(websocketconfig.Host+":"+websocketconfig.Port, "localhost.crt", "localhost.key", nil)
	//~ https://godoc.org/net/http#ListenAndServeTLS
	if err != nil {
		fmt.Println("Listen&Serve: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	setConnection := false
	token := "unregstered"
	rtoken, ok := r.URL.Query()["RegisterToken"]
	if ok && len(rtoken[0]) >= 1 {
       token = strings.Join(rtoken," ")
    } 
    if websocketconfig.AllowUnregisteredClients == "yes" { 
		setConnection = true
	} else { 
		if allowsTokens[token] { 
			setConnection = true
		} 
	}
	if setConnection {
		// Upgrade initial GET request to a websocket ======================
		upgrader.CheckOrigin = func(r *http.Request) bool {
			if ( websocketconfig.Origin == "*") {return true}
			if ( websocketconfig.Origin == r.URL.String() ) {return true}
			return false
		}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
		}
		defer ws.Close()
		// ==== Upgrade initial GET request to a websocket ======================		
		clients[ws] = token  
		fmt.Println("Register client", token)	
		// loop =====
		for {
			// Read message from browser
			msgType, msg, err := ws.ReadMessage()
			_ = msgType
			if err != nil {
				return
			}
			// Print the message to the console
			fmt.Printf("%s sent: %s\n", ws.RemoteAddr(), string(msg))
			if strings.HasPrefix(string(msg), "RegisterToken:") {
				token := strings.Split(string(msg), ":")
				clients[ws] = token[1];
			} else {
				if strings.HasPrefix(string(msg), "Echo:") {
					echomessage := strings.Split(string(msg), ":")
					ws.WriteJSON(strings.TrimPrefix(strings.Join(echomessage,""), "Echo"))
				} else {
					// Send the newly received message to the broadcast channel
					broadcast <- string(msg)
				}
			}
		}
		// === loop =====			
	} else {
		fmt.Println("Connection refused")
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client,k := range clients {
			tokens := strings.Split(string(msg), "|")
			if len(tokens) >= 1 {
				if strings.Contains(tokens[0], k) {  // OR tokens[0] == "TokenToBroadcastToAllClients"
					fmt.Println("Now send ", msg, " To:", k)
					//~ err := client.WriteJSON(msg)
					err := client.WriteJSON(strings.TrimPrefix(msg, tokens[0]+"|"))
					time.Sleep(3 * time.Millisecond)
					if err != nil {
						fmt.Println("error: %v", err)
						client.Close()
						delete(clients, client)
					}
				}
			}
		}
	}
}
