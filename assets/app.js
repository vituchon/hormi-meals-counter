var app = {
  encounter: null,
  clientPlayer: null,
  soundAllowed: true,
};

function playClickSound() {
  if (!app.soundAllowed) {
    return;
  }
  const click = new Audio('/assets/crunch.mp3');
  click.currentTime = 0;
  click.play().catch(() => {});
}

function playWelcomeSound() {
  if (!app.soundAllowed) {
    return;
  }
  const horn = new Audio('/assets/party-horn.mp3');
  horn.currentTime = 0;
  horn.play().catch(() => {});
}

function addEventIfNotRegistered(element, event, callback) {
  const attributeName = `data-event-${event}`;
  const isEventRegistered = element.getAttribute(attributeName);

  if (!isEventRegistered || isEventRegistered === "false") {
    element.addEventListener(event, callback);
    element.setAttribute(attributeName, "true");
  } else {
    console.warn("event", event, "is already registered for element", element, ". Skipping registration.");
  }
}


var Buttons;
(function (Buttons) {

  function fadeout(...ids) {
    const timeouts = ids.map(id => {
      return new Promise((resolve) => {
        let button = document.getElementById(id);
        if (button) {
          button.disabled = true;
          button.classList.remove('fade-in');
          button.classList.add('fade-out');
          setTimeout(() => {
            button.style.display = "none";
            resolve();
          }, 1000);
        } else {
          resolve();
        }
      });
    });

    return Promise.all(timeouts);
  }

  function fadein(display, ...ids) {
    const timeouts = ids.map(id => {
      return new Promise((resolve) => {
        let button = document.getElementById(id);
        if (button) {
          button.style.display = display;
          button.classList.remove('fade-out');
          button.classList.add('fade-in');
          setTimeout(() => {
            button.disabled = false;
            resolve();
          }, 1000);
        } else {
          resolve();
        }
      });
    });

    return Promise.all(timeouts);
  }

  Buttons.fadeout = fadeout;
  Buttons.fadein = fadein;
})(Buttons || (Buttons = {}));


function displayMessage(message) {
  Toastify({
    text: message,
    duration: 3000,
    newWindow: true,
    gravity: "top",
    position: 'center',
    style: {
      "font-size": "1.5em",
      "border-radius": "0.5em",
    }
  }).showToast();
}


function displayErrorMessage(message) {
  Toastify({
    text: message,
    duration: 3000,
    newWindow: true,
    gravity: "bottom",
    position: 'left',
    style: {
      "font-size": "1em",
      "border-radius": "0.5em",
      "background-color": "yellow",
      "color": "red",
    }
  }).showToast();
}

function totalCount(encounter) {
  if (!encounter || !encounter.countByPlayerId) {
    return 0;
  }
  return Object.values(encounter.countByPlayerId).reduce((a, b) => a + b, 0);
}

function exportEncounterAsJson(encounter) {
  if (!encounter) {
    return;
  }
  const payload = {
    exportedAt: new Date().toISOString(),
    encounter: encounter,
    total: totalCount(encounter),
  };
  const jsonString = JSON.stringify(payload, null, 2);
  const blob = new Blob([jsonString], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  const safeName = (encounter.name || 'encounter').replace(/[^a-z0-9_-]+/gi, '-').toLowerCase();
  const stamp = new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-');
  a.download = `hormi-meals-counter-${safeName}-${stamp}.json`;
  a.click();
  URL.revokeObjectURL(url);
}
