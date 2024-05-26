"use strict";

// constants and utilities
const B = chrome;

// receive message from backgroud script
B.runtime.onMessage.addListener((message, sender, sendResponse) => {
    console.log("message received: ", message);
    if(message.action === "search") {
        // insert at top of page a list of page
        // https://grrr.tech/posts/create-dom-node-from-html-string/
        let resultDiv = document.createElement("div");
        let htmlToInsert = message.result;
        resultDiv.innerHTML = htmlToInsert; 
        console.log("resultDiv: ", resultDiv);

        document.body.insertBefore(
            resultDiv,
            document.body.firstChild
            );
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


