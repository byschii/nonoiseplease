console.log("bg.js loaded");

let nnp_address = "";
if (!B.management.getSelf(function(info) {
    if (info.installType !== "development") {
        nnp_address = "https://nonoiseplease.com";
    } else {
        nnp_address = B.runtime.getManifest().devserver;
    } 
}));



const spawnSearch = (tab, jwt) => {
    console.log("spawning search on tab: ", tab);

    // 1 check tab is open on google
    let parsedUrl = new URL(tab.url);
    if(parsedUrl.hostname.includes(".google.com") === false) {
        console.warn("not searching on google");
        return;
    }

    // 2 grab the search query from the tab url
    let query = parsedUrl.searchParams.get("q");
    if(!query) {
        console.warn("no query found");
        return;
    }

    searchPageHTML(nnp_address, jwt, query).then((pages) => {
        B.tabs.sendMessage(tab.id, {
            action: "search",
            result: pages
        });
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

const popupLog = (msg, msgtype="ok") => {
    message = {
        action: "log",
        msg: msg,
        msgtype: msgtype,
    }
    B.runtime.sendMessage(message);
};

const popupLogError = (msg) => {
    popupLog(msg, "error");
};


// when state is resolved
storedState.then((currentState) => {
    // Listen for a tab being updated to a complete status
    B.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
        if (changeInfo.status === 'complete' && tab.active) {
            // Execute content script to get HTML
            B.tabs.executeScript(tabId, { code: 'document.documentElement.outerHTML' }, function(htmlContent) {
                if (currentState.allowTemporaryMemory) {
                    if (currentState.memory.length > currentState.memorySize) {
                        currentState.memory.shift();
                    }
                    currentState.pushToMemory({
                        html: htmlContent[0],
                        url: tab.url,
                        title: tab.title
                    });
                }
                if (currentState.recordNavigation) {
                    // just send current situation
                    sendPage(nnp_address, currentState.jwt, htmlContent, tab.url, tab.title); 
                    // keep memory and autosave independent
                    // someone can record without sending or send without recording locally
                }
                if(currentState.automaticSearch){
                    spawnSearch(tab, currentState.jwt);
                }
            });
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
            popupLog("jwt read");
        }
        else if (message.action === "jwt.delete") {
            currentState.jwt = null;
        }
        else if (message.action === "jwt.refresh") {
            console.log(nnp_address, currentState.jwt);
            refreshToken(nnp_address, currentState.jwt).then((newJwt) => {
                currentState.jwt = newJwt.token;
                B.storage.local.set({"lastState":currentState.serialize()});
            });
        }
        else if (message.action === "status.record") {
            if(!currentState.recordNavigation && message.record && currentState.memory.length > 0){
                // send all pages in memory
                popupLog("sending all stored pages");
                currentState.memory.forEach((page) => {
                    sendPage(nnp_address, currentState.jwt, page.html, page.url, page.title).then((res) => {
                        if(res){
                            popupLog("page sent");
                        } else {
                            popupLogError("page not sent");
                        }
                    }).catch(() => {
                        popupLogError("page not sent");
                    });                
                });
            }
            currentState.recordNavigation = message.record;
        }
        else if (message.action === "status.memory") {
            currentState.allowTemporaryMemory = message.memory;
        }
        else if (message.action === "status.autosearch") {
            currentState.automaticSearch = message.autosearch;
        }
        console.log("currentState after -> ", currentState);
        B.storage.local.set({"lastState":currentState.serialize()});

        if (message.action === "page.save") {
            B.tabs.query({active: true, currentWindow: true}, function(tabs) {
                let currentTab = tabs[0];
                B.tabs.executeScript(currentTab.id, { code: 'document.documentElement.outerHTML' }, function(htmlContent) {
                    sendPage(nnp_address, currentState.jwt, htmlContent[0], currentTab.url, currentTab.title.then((res) => {
                        if(res){
                            popupLog("page sent");
                        } else {
                            popupLogError("page not sent");
                        }
                    }).catch(() => {
                        popupLogError("page not sent");
                    }));
                });
            });
        }
        if (message.action === "page.search") {
            B.tabs.query({active: true, currentWindow: true}, function(tabs) {
                spawnSearch(tabs[0], currentState.jwt);
            });
        }
        if (message.action === "bookmark.import"){
            B.bookmarks.getTree(function(bookmarkTreeNodes) {
                let bookmarks = [];
                let bookmark = (node) => {
                    if (node.children) {
                        node.children.forEach(bookmark);
                    } else {
                        bookmarks.push(node.url);
                    }
                };
                bookmark(bookmarkTreeNodes[0]);
                console.log(bookmarks);
                sendBookmarks(nnp_address, currentState.jwt, bookmarks).then((res) => {
                    if(res){
                        popupLog("bookmarks sent");
                    } else {
                        popupLogError("bookmarks not sent");
                    }
                }).catch(() => {
                    popupLogError("bookmarks not sent");
                });
            });
        }
    });
});

