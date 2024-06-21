# webroot + API overhaul plans

-   completely remake the bot settings page. i could probably use a lot of the same elements, but the JS will be redone. i might make all of the gRPC requests happen in JS rather than needing to go through wire-pod.
    -   typescript?
-   better looks. i usually like to make things retro-looking. IBM PC font, simplistic page. should be much more cohesive, buttons and forms should pop out more
    -   themes? i was thinking making the background dark gray, and the user can choose the accent color.
        -   vector's eye colors?
-   overhaul wire-pod settings API. use types rather than the current mess
-   allow ollama and other custom openai-supported LLM things

## things currently done

-   i have decided to stick with wing for CSS
-   add box shadow to buttons and inputs
-   change fonts to what i like
-   simplify hr, i thought the gradient looked tacky
-   dark background
-   reduce padding under each nav element
-   center everything!
-   change KG option selectors to checkboxes
-   make sure colors update when a new icon is clicked
-   use vector's teal eye color for accents
    -    i will put in a selector for all the eye colors
-   increase debug log lines to 200
-   fix "getImage" freezes
-   fix panic upon no response from LLM
-   make sure input box colors match with rest of page
-   make stimulation graph smaller
