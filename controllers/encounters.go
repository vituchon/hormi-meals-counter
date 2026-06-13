package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/vituchon/hormi-meals-counter/repositories"
	"github.com/vituchon/hormi-meals-counter/services"
)

var encountersRepository repositories.Encounters = repositories.NewEncountersMemoryRepository()

func GetEncounters(response http.ResponseWriter, request *http.Request) {
	encounters, err := encountersRepository.GetEncounters()
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounters: '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(response, http.StatusOK, encounters)
}

func GetEncounterById(response http.ResponseWriter, request *http.Request) {
	id, err := ParseRouteParamAsInt(request, "id")
	if err != nil {
		log.Println(err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	encounter, err := encountersRepository.GetEncounterById(id)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounter(id='%d'): '%v'", id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(response, http.StatusOK, encounter)
}

func CreateEncounter(response http.ResponseWriter, request *http.Request) {
	playerId := services.GetClientId(request)

	encounter, err := retrieveEncounterByValue(request)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounter: '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}

	player, err := playersRepository.GetPlayerById(playerId)
	if err != nil {
		msg := fmt.Sprintf("error getting player(id='%d'): '%v'", playerId, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}

	encounter.Owner = *player

	created, err := encountersRepository.CreateEncounter(*encounter)
	if err != nil {
		msg := fmt.Sprintf("error while creating encounter: '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(response, http.StatusOK, created)
}

func UpdateEncounter(response http.ResponseWriter, request *http.Request) {
	encounter, err := retrieveEncounterByValue(request)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounter: '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	updated, err := encountersRepository.UpdateEncounter(*encounter)
	if err != nil {
		msg := fmt.Sprintf("error while updating encounter(id='%d'): '%v'", encounter.Id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	services.EncounterWebSockets.NotifyEncounterConns(updated.Id, "updated", updated)
	WriteJsonResponse(response, http.StatusOK, updated)
}

func DeleteEncounter(response http.ResponseWriter, request *http.Request) {
	encounter, err := retrieveEncounterByReference(request)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounter: '%v'", err)
		log.Println(msg)
		if errors.Is(err, NoEncounterIdRouteParamErr) {
			http.Error(response, msg, http.StatusBadRequest)
		} else {
			http.Error(response, msg, http.StatusInternalServerError)
		}
		return
	}

	id := encounter.Id
	err = encountersRepository.DeleteEncounter(id)
	if err != nil {
		msg := fmt.Sprintf("error while deleting encounter(id='%d'): '%v'", id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	services.EncounterWebSockets.UnbindAllWebSocketsInEncounter(id, request)
	response.WriteHeader(http.StatusOK)
}

func DeleteEncounters(response http.ResponseWriter, request *http.Request) {
	encounters, err := encountersRepository.GetEncounters()
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounters: '%v'", err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}

	for _, encounter := range encounters {
		err = encountersRepository.DeleteEncounter(encounter.Id)
		if err != nil {
			msg := fmt.Sprintf("error while deleting encounter(id='%d'): '%v'", encounter.Id, err)
			log.Println(msg)
			http.Error(response, msg, http.StatusInternalServerError)
			return
		}
		services.EncounterWebSockets.UnbindAllWebSocketsInEncounter(encounter.Id, request)
	}
	response.WriteHeader(http.StatusOK)
}

func JoinEncounter(response http.ResponseWriter, request *http.Request) {
	encounter, err := retrieveEncounterByReference(request)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounter: '%v'", err)
		log.Println(msg)
		if errors.Is(err, NoEncounterIdRouteParamErr) {
			http.Error(response, msg, http.StatusBadRequest)
		} else {
			http.Error(response, msg, http.StatusInternalServerError)
		}
		return
	}

	playerId := services.GetClientId(request)
	player, err := playersRepository.GetPlayerById(playerId)
	if err != nil {
		msg := fmt.Sprintf("error while getting player(id='%d'): '%v'", playerId, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}

	err = encounter.Join(*player)
	if err != nil {
		msg := fmt.Sprintf("error while joining encounter(id='%d'): '%v'", encounter.Id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}
	updated, err := encountersRepository.UpdateEncounter(*encounter)
	if err != nil {
		msg := fmt.Sprintf("error while updating encounter(id='%d'): '%v'", encounter.Id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	payload := services.WebSockectOutgoingJoinMsgPayload{Encounter: updated, Player: player}
	services.EncounterWebSockets.NotifyEncounterConns(encounter.Id, "player-join", payload)
	WriteJsonResponse(response, http.StatusOK, updated)
}

func QuitEncounter(response http.ResponseWriter, request *http.Request) {
	encounter, err := retrieveEncounterByReference(request)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounter: '%v'", err)
		log.Println(msg)
		if errors.Is(err, NoEncounterIdRouteParamErr) {
			http.Error(response, msg, http.StatusBadRequest)
		} else {
			http.Error(response, msg, http.StatusInternalServerError)
		}
		return
	}

	playerId := services.GetClientId(request)
	player, err := playersRepository.GetPlayerById(playerId)
	if err != nil {
		msg := fmt.Sprintf("error while getting player(id='%d'): '%v'", playerId, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}

	err = encounter.Quit(*player)
	if err != nil {
		msg := fmt.Sprintf("error while quitting encounter(id='%d'): '%v'", encounter.Id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}
	updated, err := encountersRepository.UpdateEncounter(*encounter)
	if err != nil {
		msg := fmt.Sprintf("error while updating encounter(id='%d'): '%v'", encounter.Id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	payload := services.WebSockectOutgoingQuitMsgPayload{Encounter: updated, Player: player}
	services.EncounterWebSockets.NotifyEncounterConns(encounter.Id, "encounter-quit", payload)
	WriteJsonResponse(response, http.StatusOK, updated)
}

func IncrementCounter(response http.ResponseWriter, request *http.Request) {
	encounter, err := retrieveEncounterByReference(request)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounter: '%v'", err)
		log.Println(msg)
		if errors.Is(err, NoEncounterIdRouteParamErr) {
			http.Error(response, msg, http.StatusBadRequest)
		} else {
			http.Error(response, msg, http.StatusInternalServerError)
		}
		return
	}

	playerId := services.GetClientId(request)
	count, err := encounter.Increment(playerId)
	if err != nil {
		msg := fmt.Sprintf("error while incrementing counter for player(id='%d') in encounter(id='%d'): '%v'", playerId, encounter.Id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
		return
	}

	updated, err := encountersRepository.UpdateEncounter(*encounter)
	if err != nil {
		msg := fmt.Sprintf("error while updating encounter(id='%d') after increment: '%v'", encounter.Id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}

	payload := services.WebSockectOutgoingCounterMsgPayload{Encounter: updated, PlayerId: playerId, Count: count}
	services.EncounterWebSockets.NotifyEncounterConns(encounter.Id, "counter-incremented", payload)
	WriteJsonResponse(response, http.StatusOK, payload)
}

func ResetCounters(response http.ResponseWriter, request *http.Request) {
	encounter, err := retrieveEncounterByReference(request)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving encounter: '%v'", err)
		log.Println(msg)
		if errors.Is(err, NoEncounterIdRouteParamErr) {
			http.Error(response, msg, http.StatusBadRequest)
		} else {
			http.Error(response, msg, http.StatusInternalServerError)
		}
		return
	}

	playerId := services.GetClientId(request)
	if encounter.Owner.Id != playerId {
		msg := fmt.Sprintf("error while resetting counters: request doesn't come from owner of encounter(id='%d'), came from player(id='%d')", encounter.Id, playerId)
		log.Println(msg)
		http.Error(response, msg, http.StatusForbidden)
		return
	}

	encounter.ResetCounts()
	updated, err := encountersRepository.UpdateEncounter(*encounter)
	if err != nil {
		msg := fmt.Sprintf("error while updating encounter(id='%d') after reset: '%v'", encounter.Id, err)
		log.Println(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}
	services.EncounterWebSockets.NotifyEncounterConns(encounter.Id, "counter-reset", updated)
	WriteJsonResponse(response, http.StatusOK, updated)
}

func BindClientWebSocketToEncounter(response http.ResponseWriter, request *http.Request) {
	encounterId, err := ParseRouteParamAsInt(request, "id")
	if err != nil {
		log.Println(err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	services.EncounterWebSockets.BindClientWebSocketToEncounter(response, request, encounterId)
	response.WriteHeader(http.StatusOK)
}

func UnbindClientWebSocketInEncounter(response http.ResponseWriter, request *http.Request) {
	conn := services.WebSocketsHandler.Retrieve(request)
	if conn != nil {
		services.EncounterWebSockets.UnbindClientWebSocketInEncounter(conn, request)
		response.WriteHeader(http.StatusOK)
	} else {
		msg := fmt.Sprintf("no need to release web socket as it was not adquired (or already released) for client(id='%d')", services.GetClientId(request))
		log.Println(msg)
		http.Error(response, msg, http.StatusBadRequest)
	}
}

var NoEncounterIdRouteParamErr = errors.New("the request URL is missing the encounter ID as a route parameter")

func retrieveEncounterByReference(request *http.Request) (*repositories.Encounter, error) {
	id, err := ParseRouteParamAsInt(request, "id")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", NoEncounterIdRouteParamErr, err)
	}

	encounter, err := encountersRepository.GetEncounterById(id)
	if err != nil {
		errMsg := fmt.Sprintf("error while retrieving encounter(id='%d'): '%v'", id, err)
		return nil, errors.New(errMsg)
	}
	return encounter, nil
}

func retrieveEncounterByValue(request *http.Request) (*repositories.Encounter, error) {
	var encounter repositories.Encounter
	err := parseJsonFromReader(request.Body, &encounter)
	if err != nil {
		errMsg := fmt.Sprintf("error reading request body: '%v'", err)
		return nil, errors.New(errMsg)
	}
	return &encounter, nil
}
