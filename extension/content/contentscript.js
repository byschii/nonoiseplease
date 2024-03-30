"use strict";

// constants and utilities
const B = browser || chrome;

const createPageList = (pages) => {
    const pageList = document.createElement("ul");
    pageList.style.background = "white";
    pageList.style.display = "flex";
    pageList.style.zIndex = "1000";
    pageList.id = "nnpext-page-list";
    pages.forEach((page) => {
        const pageLink = document.createElement("a");
        const pageListItem = document.createElement("li");
        pageLink.href = page.url;
        pageLink.textContent = page.title;
        pageLink.style.color = "black";
        pageListItem.appendChild(pageLink);
        pageList.appendChild(pageListItem);
    });
    return pageList;
}

var searchAlreadyInserted = false;
var allowedDomains = ["google.com", "bing.com"];

// receive message from backgroud script
B.runtime.onMessage.addListener((message, sender, sendResponse) => {
    console.log("message received: ", message);
    if(message.action === "search" && !searchAlreadyInserted && allowedDomains.includes(new URL(message.pages[0].url).hostname)) {
        // insert at top of page a list of page
        document.body.insertAdjacentElement("afterbegin", createPageList(message.pages));
        searchAlreadyInserted = true;
    }

    if(message.action === "jwt.read") {
        let jwtText = document.getElementById("pbjwt").textContent;
        console.log("jwt = "+jwtText)
        sendResponse({jwt:jwtText})
        return;
    }
});


