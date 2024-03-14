
// constants and utilities
const B = browser || chrome;


let currentlyActive = false;

B.runtime.onMessage.addListener((message, sender, sendResponse) => {

    if (message.action === "search") {
        console.log("searching");
        currentlyActive = true;
        const searchResults = document.querySelectorAll("a");
        searchResults.forEach((result) => {
            result.style.backgroundColor = "yellow";
        });
    }
    

});


