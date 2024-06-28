var fontMap = {
    "droidsans": "DroidSans",
    "ibmvga": "IBMVGA"
};

var colorMap = {
    "teal": {
        original: "#00ff80",
        lighter: "#00ff80"
    },
    "orange": {
        original: "#ff3d0a",
        lighter: "#ff3d0a"
    },
    "yellow": {
        original: "#ffeb00",
        lighter: "#fff766"
    },
    "lime": {
        original: "#6aff00",
        lighter: "#b3ff66"
    },
    "sapphire": {
        original: "#009aff",
        lighter: "#66cdff"
    },
    "purple": {
        original: "#cd00c1",
        lighter: "#e866dc"
    },
    "green": {
        original: "#00ff00",
        lighter: "#66ff66"
    }
};

function showUICustomizer() {
    toggleVisibility(["section-log", "section-botauth", "section-intents", "section-version", "section-intents"], "section-uicustomizer", "icon-Customizer");
}

function setUIFont() {
    let bodyFont = getValue("body-font-choose");
    document.documentElement.style.setProperty('--body-font-family', fontMap[bodyFont]);
    localStorage.setItem('bodyFont', bodyFont);
}

function setUIColor() {
    let accentColor = colorMap[getValue("accent-color-choose")].lighter;
    document.documentElement.style.setProperty('--fg-color', accentColor);
    localStorage.setItem('accentColor', getValue("accent-color-choose"));
}

function getValue(element) {
    return document.getElementById(element).value;
}

function loadSettings() {
    let savedFont = localStorage.getItem('bodyFont');
    let savedColor = localStorage.getItem('accentColor');

    if (savedFont) {
        document.documentElement.style.setProperty('--body-font-family', fontMap[savedFont]);
        if (document.getElementById("body-font-choose")) {
            document.getElementById("body-font-choose").value = savedFont;
        }
    }

    if (savedColor) {
        document.documentElement.style.setProperty('--fg-color', colorMap[savedColor].original);
        if (document.getElementById("accent-color-choose")) {
            document.getElementById("accent-color-choose").value = savedColor;
        }
    }
}

// call loadSettings
loadSettings();

