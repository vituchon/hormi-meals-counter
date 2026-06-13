package repositories

import (
	"errors"

	"github.com/vituchon/escobita/util"
)

type Encounter struct {
	Id              int         `json:"id"`
	Name            string      `json:"name"`
	Owner           Player      `json:"owner"`
	Players         []Player    `json:"players"`
	CountByPlayerId map[int]int `json:"countByPlayerId"`
	Created         int64       `json:"created"`
}

func (encounter *Encounter) Join(player Player) error {
	joinedPlayer := util.Find(encounter.Players, func(joined Player) bool { return joined.Id == player.Id })
	playerNotJoined := joinedPlayer == nil
	if playerNotJoined {
		encounter.Players = append(encounter.Players, player)
		if encounter.CountByPlayerId == nil {
			encounter.CountByPlayerId = make(map[int]int)
		}
		if _, exists := encounter.CountByPlayerId[player.Id]; !exists {
			encounter.CountByPlayerId[player.Id] = 0
		}
	} else {
		return PlayerAlreadyJoinedErr
	}
	return nil
}

func (encounter *Encounter) IsJoined(player Player) bool {
	joinedPlayer := util.Find(encounter.Players, func(joined Player) bool { return joined.Id == player.Id })
	return joinedPlayer != nil
}

func (encounter *Encounter) Quit(player Player) error {
	var playerIndex int = -1
	for i, joined := range encounter.Players {
		if joined.Id == player.Id {
			playerIndex = i
			break
		}
	}
	playerJoined := playerIndex != -1
	if playerJoined {
		encounter.Players = append(encounter.Players[:playerIndex], encounter.Players[playerIndex+1:]...)
	} else {
		return PlayerNotJoinedErr
	}
	return nil
}

func (encounter *Encounter) Increment(playerId int) (int, error) {
	if encounter.CountByPlayerId == nil {
		encounter.CountByPlayerId = make(map[int]int)
	}
	if _, exists := encounter.CountByPlayerId[playerId]; !exists {
		return 0, PlayerNotJoinedErr
	}
	encounter.CountByPlayerId[playerId]++
	return encounter.CountByPlayerId[playerId], nil
}

func (encounter *Encounter) ResetCounts() {
	for playerId := range encounter.CountByPlayerId {
		encounter.CountByPlayerId[playerId] = 0
	}
}

func (encounter Encounter) TotalCount() int {
	total := 0
	for _, count := range encounter.CountByPlayerId {
		total += count
	}
	return total
}

func (encounter Encounter) CanPlayerDelete(player Player) bool {
	return encounter.Owner.Id == player.Id
}

var PlayerAlreadyJoinedErr error = errors.New("The player has already joined the encounter")
var PlayerNotJoinedErr error = errors.New("The player has not joined the encounter")
