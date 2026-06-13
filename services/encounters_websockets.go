package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"sync"

	"github.com/gorilla/websocket"
	"github.com/vituchon/hormi-meals-counter/repositories"
)

type encounterWebSockets struct {
	ConnsByEncounterId map[int][]*websocket.Conn
	mutex              sync.Mutex
}

var EncounterWebSockets encounterWebSockets = encounterWebSockets{ConnsByEncounterId: make(map[int][]*websocket.Conn)}

func (ews *encounterWebSockets) NotifyEncounterConns(encounterId int, kind string, data interface{}) {
	type Notification struct {
		Kind      string      `json:"kind"`
		BagOfCats interface{} `json:"data"`
	}

	ews.mutex.Lock()
	defer ews.mutex.Unlock()
	conns := ews.ConnsByEncounterId[encounterId]
	log.Printf("Notifying kind='%s' to encounter(id='%d') for %d conn(s)", kind, encounterId, len(conns))
	for _, conn := range conns {
		notification := Notification{Kind: kind, BagOfCats: data}
		notificationAsJson, err := json.Marshal(notification)
		if err != nil {
			log.Printf("Error on marshalling notification, skip send. Error was: '%v'\n", err)
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, notificationAsJson)
		if err != nil {
			log.Printf("Error writing kind='%s' notification to conn(remoteAddr='%s') in encounter(id='%d'): '%v'", kind, conn.RemoteAddr().String(), encounterId, err)
		} else {
			log.Printf("Sent kind='%s' notification to conn(remoteAddr='%s') in encounter(id='%d')", kind, conn.RemoteAddr().String(), encounterId)
		}
	}
}

func (ews *encounterWebSockets) BindClientWebSocketToEncounter(response http.ResponseWriter, request *http.Request, encounterId int) {
	log.Printf("Binding web socket from client(id='%d') in encounter(id='%d')...", GetClientId(request), encounterId)
	conn, _, err := WebSocketsHandler.AdquireOrRetrieve(response, request)
	if err != nil {
		log.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	ews.mutex.Lock()
	defer ews.mutex.Unlock()

	for _, existingConn := range ews.ConnsByEncounterId[encounterId] {
		if existingConn == conn {
			msg := fmt.Sprintf("Web socket(remoteAddr='%s') from client(id='%d') already bound in encounter(id='%d')", conn.RemoteAddr().String(), GetClientId(request), encounterId)
			log.Println(msg)
			http.Error(response, msg, http.StatusBadRequest)
			return
		}
	}

	ews.ConnsByEncounterId[encounterId] = append(ews.ConnsByEncounterId[encounterId], conn)
	log.Printf("Bound web socket(remoteAddr='%s') from client(id='%d') in encounter(id='%d')", conn.RemoteAddr().String(), GetClientId(request), encounterId)
}

func (ews *encounterWebSockets) UnbindAllWebSocketsInEncounter(encounterId int, request *http.Request) {
	ews.mutex.Lock()
	defer ews.mutex.Unlock()
	log.Printf("Unbinding all web sockets from encounter(id='%d')...\n", encounterId)

	for _, conn := range ews.ConnsByEncounterId[encounterId] {
		ews.doUnbindClientWebSocketInEncounter(conn, encounterId, request)
	}
	delete(ews.ConnsByEncounterId, encounterId)

	log.Printf("Unbound all web sockets from encounter(id='%d')\n", encounterId)
}

func (ews *encounterWebSockets) UnbindClientWebSocketInEncounter(conn *websocket.Conn, request *http.Request) {
	ews.mutex.Lock()
	defer ews.mutex.Unlock()
	log.Printf("Unbinding web socket(remoteAddr='%s') from a possible joined encounter...\n", conn.RemoteAddr().String())

	for encounterId, conns := range ews.ConnsByEncounterId {
		for _, _conn := range conns {
			if _conn == conn {
				ews.doUnbindClientWebSocketInEncounter(conn, encounterId, request)
				return
			}
		}
	}
	log.Printf("Web socket(remoteAddr='%s') was NOT bound to an encounter\n", conn.RemoteAddr().String())
}

func (ews *encounterWebSockets) doUnbindClientWebSocketInEncounter(givenConn *websocket.Conn, encounterId int, request *http.Request) {
	log.Printf("Unbinding web socket(remoteAddr='%s') in encounter(id='%d')...\n", givenConn.RemoteAddr().String(), encounterId)
	conns := ews.ConnsByEncounterId[encounterId]
	connsPtr := &conns
	chopped := (*connsPtr)[:0]
	for _, conn := range conns {
		if givenConn != conn {
			chopped = append(chopped, conn)
		}
	}
	*connsPtr = chopped
	ews.ConnsByEncounterId[encounterId] = *connsPtr
	log.Printf("Unbound web socket(remoteAddr='%s') in encounter(id='%d')\n", givenConn.RemoteAddr().String(), encounterId)
}

type WebSockectOutgoingAccessMsgPayload struct {
	Encounter *repositories.Encounter `json:"encounter"`
	Player    *repositories.Player    `json:"player,omitempty"`
}

type WebSockectOutgoingJoinMsgPayload = WebSockectOutgoingAccessMsgPayload

type WebSockectOutgoingQuitMsgPayload = WebSockectOutgoingAccessMsgPayload

type WebSockectOutgoingCounterMsgPayload struct {
	Encounter *repositories.Encounter `json:"encounter"`
	PlayerId  int                     `json:"playerId"`
	Count     int                     `json:"count"`
}
