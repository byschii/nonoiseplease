"use strict";

// constants and utilities
const B = browser || chrome;
var searchAlreadyInserted = false;

// receive message from backgroud script
B.runtime.onMessage.addListener((message, sender, sendResponse) => {
    console.log("message received: ", message);
    if(message.action === "search" && !searchAlreadyInserted) {
        // insert at top of page a list of page
        // https://grrr.tech/posts/create-dom-node-from-html-string/
        let resultDiv = document.createElement("div");
        resultDiv.innerHTML = message.result;
        console.log("resultDiv: ", resultDiv);

        document.body.insertBefore(
            resultDiv,
            document.body.firstChild
            );
        searchAlreadyInserted = true;
    }

    // just grab the jwt from nnp and send it background
    // now background can do calls to backend easilly
    if(message.action === "jwt.read") {
        let jwtText = document.getElementById("pbjwt").textContent;
        console.log("jwt = "+jwtText)
        sendResponse({jwt:jwtText})
        return;
    }
});


