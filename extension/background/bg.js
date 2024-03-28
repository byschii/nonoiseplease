console.log("bg.js loaded");

let nnp_address = "";
if (!B.management.getSelf(function(info) {
    if (info.installType !== "development") {
        nnp_address = "https://nonoiseplease.com";
    } else {
        nnp_address = B.runtime.getManifest().devserver;
    } 
}));

const postRequest = (userId, extensionToken, html, url, title) => {
    // log body with shortened html
    console.debug("body: ", JSON.stringify({
        html: html.substring(0, 40),
        url: url,
        title: title,
        extention_token: extensionToken,
        user_id: userId
    }));

    return {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            html: html,
            url: url,
            title: title,
            extention_token: extensionToken,
            user_id: userId
        }),
    }
};
const sendPage = async (userId, extensionToken, html, url, title) => {
    console.log("sending page");
    if (!userId || !extensionToken) {
        console.error("no userId, extensionToken");
        return false;
    }
    let res = await fetch(
        nnp_address + "/api/page-manage/load", postRequest(userId, extensionToken, html, url, title)
    ).then((response) => {
        console.log(response.ok? "response ok" : "response not ok")
        if (response.ok) {
            succMsg.style.display = "block";
            errMsg.style.display = "none";
        }
        return true;
    }).catch( // return false
        () => false
    );
    return res;
};

const spawnSearch = (tabId, query) => {
    B.tabs.sendMessage(tab.id, {
        action: "search",
        pages: [
            {url: "https://www.google.com", title: "Google"},
            {url: "https://www.bing.com", title: "Bing"}
        ]
    });
};

const grabJwt = async (currentTab) => {

    if (!currentTab || currentTab.url.indexOf(nnp_address) === -1) {
        console.warn("current tab is not nnp", currentTab);
        return null;
    }
    let response = await B.tabs.sendMessage(currentTab.id, {
        action: "jwt.read",
    });
    console.log("response from content script:", response);
    return response.jwt;
};

// when state is resolved
storedState.then((currentState) => {
    // Listen for a tab being updated to a complete status
    B.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
        if (changeInfo.status === 'complete' && tab.active) {
            const htmlContent = document.documentElement.innerHTML;
            if (currentState.allowTemporaryMemory) {
                currentState.pushToMemory({
                    html: htmlContent,
                    url: tab.url,
                    title: tab.title
                });
            }
            if (currentState.recordNavigation) {
                sendPage(currentState.userId, currentState.extensionToken, htmlContent, tab.url, tab.title); // too keep memory and record independent
            }
            if (currentState.memory.length > currentState.memorySize) {
                currentState.memory.shift();
            }
            if(currentState.automaticSearch){
                spawnSearch(tab.id, "test")
            }
        }
    });


    // and listen for messages from the popup
    B.runtime.onMessage.addListener((message, sender, sendResponse) => {
        console.log("currentState before -> ", currentState);
        if (message.action === "jwt.read") {
            B.tabs.query({active: true, currentWindow: true}, async function(tabs) {
                currentState.jwt = await grabJwt(tabs[0]);
                B.storage.local.set({"lastState":currentState.serialize()});
            });
        }
        else if (message.action === "jwt.delete") {
            currentState.jwt = null;
        }
        else if (message.action === "status.record") {
            if(!currentState.recordNavigation && message.record && currentState.memory.length > 0){
                // send all pages in memory
                currentState.memory.forEach((page) => {
                    sendPage(currentState.userId, currentState.extensionToken, page.html, page.url, page.title);
                });
            }
            currentState.recordNavigation = message.record;
        }
        else if (message.action === "status.memory") {
            currentState.allowTemporaryMemory = message.memory;
        }
        console.log("currentState after -> ", currentState);
        B.storage.local.set({"lastState":currentState.serialize()});

        if (message.action === "page.save") {
            B.tabs.query({active: true, currentWindow: true}, function(tabs) {
                var tabId = tabs[0].id;
                console.log("tabId: ", tabId);
                sendPage(currentState.userId, currentState.extensionToken, document.documentElement.innerHTML, tabs[0].url, tabs[0].title);
            });
        }
        if (message.action === "page.search") {
            B.tabs.query({active: true, currentWindow: true}, function(tabs) {
                spawnSearch(tabs[0].id, "test");
            });
        }
    });
});

