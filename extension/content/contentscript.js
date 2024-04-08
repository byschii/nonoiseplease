"use strict";

// constants and utilities
const B = browser || chrome;

const createPageList = (pages) => {
    console.log("pages: ", pages);
    const pageList = document.createElement("ul");
    pageList.style.background = "white";
    pageList.style.display = "block";
    pageList.style.zIndex = "1000";
    pageList.style.position = "absolute";
    pageList.style.top = "0";
    pageList.style.left = "0";
    pageList.style.right = "0";
    pageList.style.width = "100%";
    pageList.id = "nnpext-page-list";
    pages.forEach((page) => {
        const pageLink = document.createElement("a");
        const pageListItem = document.createElement("li");
        pageLink.href = page.url;
        pageLink.textContent = page.title + " (" + page.url + ")";
        pageLink.style.color = "black";
        pageListItem.appendChild(pageLink);
        pageList.appendChild(pageListItem);
    });
    return pageList;
}

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

    if(message.action === "jwt.read") {
        let jwtText = document.getElementById("pbjwt").textContent;
        console.log("jwt = "+jwtText)
        sendResponse({jwt:jwtText})
        return;
    }
});


