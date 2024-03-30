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
    console.log("spawning search");

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

    // 3 do a search on nnp
    searchPage(nnp_address, jwt, query).then((pages) => {
        console.log("search results: ", pages.pages);
        // 3 bis pages -> {url,title}
        let parsedPages = pages.pages.map((page) => {
            return {
                url: page.page.link,
                title: page.page.page_title
            }
        });

        // 4 send the results to the content script and let content script display them
        B.tabs.sendMessage(tab.id, {
            action: "search",
            pages: parsedPages
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

// when state is resolved
storedState.then((currentState) => {
    // Listen for a tab being updated to a complete status
    B.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
        if (changeInfo.status === 'complete' && tab.active) {
            // Execute content script to get HTML
            B.tabs.executeScript(tabId, { code: 'document.documentElement.outerHTML' }, function(htmlContent) {
                if (currentState.allowTemporaryMemory) {
                    currentState.pushToMemory({
                        html: htmlContent,
                        url: tab.url,
                        title: tab.title
                    });
                }
                if (currentState.recordNavigation) {
                    sendPage(nnp_address, currentState.jwt, htmlContent, tab.url, tab.title); // too keep memory and record independent
                }
                if (currentState.memory.length > currentState.memorySize) {
                    currentState.memory.shift();
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
                currentState.memory.forEach((page) => {
                    sendPage(nnp_address, currentState.jwt, page.html, page.url, page.title);
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
                    sendPage(nnp_address, currentState.jwt, htmlContent[0], currentTab.url, currentTab.title);
                });
            });
        }
        if (message.action === "page.search") {
            B.tabs.query({active: true, currentWindow: true}, function(tabs) {
                spawnSearch(tabs[0].id, "test");
            });
        }
    });
});

