"use strict";
function handleFetchResponseError(response) {
  return response.text().then((text) => {
      text = text || "Error";
      throw new Error(`${text}, status code: ${response.status}`);
  });
}

class Service {
  constructor() { }

  registerPlayer(player) {
      const requestInit = {
          method: 'POST',
          headers: {
              'Accept': 'application/json',
              'Content-Type': 'application/json'
          },
      };
      return fetch("/api/v1/players/register?name="+player.name+"&emotar="+player.emotar, requestInit).then((r) => {
          if (!r.ok) {
              return handleFetchResponseError(r);
          }
          return r.json().then((loggedPlayer) => {
              return loggedPlayer;
          });
      }).catch((err) => {
          console.error('Failed to login:', err);
          throw err;
      });
  }

  getEncounters() {
      const requestInit = {
          method: 'GET',
          headers: {
              'Accept': 'application/json',
              'Content-Type': 'application/json',
          },
      };
      return fetch(`/api/v1/encounters`, requestInit).then((r) => {
          if (!r.ok) {
              return handleFetchResponseError(r);
          }
          return r.json().then((encounters) => {
              return encounters;
          });
      }).catch((err) => {
          console.error('Failed to fetch encounters:', err);
          throw err;
      });
  }

  createEncounter(encounter) {
    const requestInit = {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(encounter)
    };
    return fetch(`/api/v1/encounters`, requestInit).then((r) => {
        if (!r.ok) {
            return handleFetchResponseError(r);
        }
        return r.json().then((created) => {
            return created;
        });
    }).catch((err) => {
        console.error('Failed to create encounter:', err);
        throw err;
    });
  }

  joinEncounter(encounterId) {
    const requestInit = {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
    };
    return fetch(`/api/v1/encounters/${encounterId}/join`, requestInit).then((r) => {
        if (!r.ok) {
            return handleFetchResponseError(r);
        }
        return r.json().then((encounter) => {
            return this.bindWebSocket(encounter.id).then(() => {
              return encounter;
            });
        });
    }).catch((err) => {
        console.error('Failed to join encounter:', err);
        throw err;
    });
  }

  bindWebSocket(id) {
    return fetch(`/api/v1/encounters/${id}/bind-ws`);
  }

  quitEncounter(encounter) {
    const requestInit = {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
    };
    return fetch(`/api/v1/encounters/${encounter.id}/quit`, requestInit).then((r) => {
      if (!r.ok) {
            return handleFetchResponseError(r);
        }
        return r.json().then((updated) => {
            return this.unbindWebSocket(updated.id).then(() => {
              return updated;
            });
        });
    }).catch((err) => {
        console.error('Failed to quit encounter:', err);
        throw err;
    });
  }

  unbindWebSocket(id) {
    return fetch(`/api/v1/encounters/${id}/unbind-ws`);
  }

  incrementCounter(encounterId) {
    const requestInit = {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
    };
    return fetch(`/api/v1/encounters/${encounterId}/increment`, requestInit).then((r) => {
        if (!r.ok) {
          return handleFetchResponseError(r);
        }
        return r.json().then((payload) => {
          return payload;
        });
    }).catch((err) => {
        console.error('Failed to increment counter:', err);
        throw err;
    });
  }

  resetCounters(encounterId) {
    const requestInit = {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
    };
    return fetch(`/api/v1/encounters/${encounterId}/reset`, requestInit).then((r) => {
        if (!r.ok) {
          return handleFetchResponseError(r);
        }
        return r.json().then((encounter) => {
          return encounter;
        });
    }).catch((err) => {
        console.error('Failed to reset counters:', err);
        throw err;
    });
  }

  deleteAllEncounters() {
    const requestInit = {
        method: 'DELETE',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
    };
    return fetch(`/api/v1/encounters`, requestInit).then((r) => {
      if (!r.ok) {
        return handleFetchResponseError(r);
      }
      return r.ok;
    }).catch((err) => {
      console.error('Failed to delete all encounters:', err);
      throw err;
    });
  }
}

const DefaultService = new Service();

export { DefaultService, Service };
