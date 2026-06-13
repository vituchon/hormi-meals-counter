package repositories

import (
	"sync"
	"time"
)

type EncountersMemoryRepository struct {
	encountersById              map[int]Encounter
	encountersCreatedByPlayerId map[int]int
	idSequence                  int
	mutex                       sync.Mutex
}

func NewEncountersMemoryRepository() *EncountersMemoryRepository {
	return &EncountersMemoryRepository{
		encountersById:              make(map[int]Encounter),
		encountersCreatedByPlayerId: make(map[int]int),
		idSequence:                  0,
	}
}

func (repo *EncountersMemoryRepository) GetEncounters() ([]Encounter, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	encounters := make([]Encounter, 0, len(repo.encountersById))
	for _, encounter := range repo.encountersById {
		encounters = append(encounters, encounter)
	}
	return encounters, nil
}

func (repo *EncountersMemoryRepository) GetEncounterById(id int) (*Encounter, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	encounter, exists := repo.encountersById[id]
	if !exists {
		return nil, EntityNotExistsErr
	}
	return &encounter, nil
}

func (repo *EncountersMemoryRepository) CreateEncounter(encounter Encounter) (created *Encounter, err error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	nextId := repo.idSequence + 1
	encounter.Id = nextId
	encounter.Created = time.Now().Unix()
	if encounter.CountByPlayerId == nil {
		encounter.CountByPlayerId = make(map[int]int)
	}
	repo.encountersById[nextId] = encounter
	repo.idSequence++
	repo.encountersCreatedByPlayerId[encounter.Owner.Id]++
	return &encounter, nil
}

func (repo *EncountersMemoryRepository) UpdateEncounter(encounter Encounter) (updated *Encounter, err error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	repo.encountersById[encounter.Id] = encounter
	return &encounter, nil
}

func (repo *EncountersMemoryRepository) DeleteEncounter(id int) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	encounter := repo.encountersById[id]
	repo.encountersCreatedByPlayerId[encounter.Owner.Id]--
	delete(repo.encountersById, id)
	return nil
}

func (repo *EncountersMemoryRepository) GetEncountersCreatedCount(playerId int) int {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	return repo.encountersCreatedByPlayerId[playerId]
}
