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

/**
 * Horizontal dock (#hub-nav / #dock):
 * - .dock-fits when all tabs fit → CSS centers (setup + hub)
 * - .dock-overflow when not → flex-start + scroll (center+overflow
 *   is a classic flex bug that hides the rightmost tabs)
 * - edge fade classes for swipe affordance
 * - vertical wheel → horizontal scroll when the rail overflows
 *   (narrow desktop windows ~428px)
 */
function updateDockScrollFade(el) {
    if (!el) return;

    // Measure with start alignment so scrollWidth is trustworthy
    var prevLeft = el.scrollLeft;
    el.classList.add("dock-measuring");
    el.classList.remove("dock-fits", "dock-overflow");
    // force reflow
    void el.offsetWidth;

    var maxScroll = el.scrollWidth - el.clientWidth;
    var overflow = maxScroll > 2;

    el.classList.remove("dock-measuring");
    el.classList.toggle("dock-overflow", overflow);
    el.classList.toggle("dock-fits", !overflow);

    if (!overflow) {
        el.classList.remove("dock-fade-left", "dock-fade-right");
        el.scrollLeft = 0;
        return;
    }

    // restore scroll position within new max
    el.scrollLeft = Math.min(prevLeft, Math.max(0, maxScroll));
    maxScroll = el.scrollWidth - el.clientWidth;
    var sl = el.scrollLeft;
    el.classList.toggle("dock-fade-left", sl > 2);
    el.classList.toggle("dock-fade-right", sl < maxScroll - 2);
}

function bindDockScrollFades(el) {
    if (!el || el.dataset.dockFadeBound === "1") return;
    el.dataset.dockFadeBound = "1";

    var update = function () {
        updateDockScrollFade(el);
    };

    el.addEventListener("scroll", update, { passive: true });
    window.addEventListener("resize", update);

    // Map vertical wheel to horizontal scroll when the rail overflows
    // (desktop trackpad/mouse users on a narrow window).
    el.addEventListener(
        "wheel",
        function (e) {
            if (el.scrollWidth <= el.clientWidth + 2) return;
            // Already horizontal gesture — leave it alone
            if (Math.abs(e.deltaX) > Math.abs(e.deltaY)) return;
            if (e.deltaY === 0) return;
            el.scrollLeft += e.deltaY;
            e.preventDefault();
            updateDockScrollFade(el);
        },
        { passive: false }
    );

    if (typeof ResizeObserver !== "undefined") {
        try {
            var ro = new ResizeObserver(update);
            ro.observe(el);
        } catch (err) {
            /* ignore */
        }
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", update);
    } else {
        requestAnimationFrame(update);
    }
    // icon kit / fonts can change tab widths after first paint
    setTimeout(update, 300);
    setTimeout(update, 1000);
}

function initDockScrollFades() {
    bindDockScrollFades(document.getElementById("hub-nav"));
    bindDockScrollFades(document.getElementById("dock"));
}

initDockScrollFades();

