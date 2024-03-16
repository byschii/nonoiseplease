

// constants and utilities
const B = browser || chrome;
const extId = B.runtime.id;


// get current state
B.storage.local.get("lastState").then((res) => {
    console.log("lastState: ", res.lastState);
    if (res.lastState) {
        const currentState = res.lastState;
        document.getElementById("nnpext-user-id").value = currentState.userId;
        document.getElementById("nnpext-extension-token").value = currentState.extensionToken;
        document.getElementById("nnpext-memory").checked = currentState.allowTemporaryMemory;
        document.getElementById("nnpext-record").checked = currentState.recordNavigation;
    }
}).catch(
    () => false
);


// event listeners
document.getElementById("nnpext-user-id").addEventListener("change", (event) => {
    console.log("user id changed");
    B.runtime.sendMessage(extId, {
        action: "status.userid",
        userid: event.target.value
    });
});
document.getElementById("nnpext-extension-token").addEventListener("change", (event) => {
    console.log("extension token changed");
    B.runtime.sendMessage(extId, {
        action: "status.extensionToken",
        extensionToken: event.target.value
    });
});
document.getElementById("nnpext-memory").addEventListener("change", (event) => {
    console.log("memory changed");
    B.runtime.sendMessage(extId, {
        action: "status.memory",
        memory: event.target.checked
    });
});
document.getElementById("nnpext-record").addEventListener("change", (event) => {
    console.log("record changed");
    B.runtime.sendMessage(extId, {
        action: "status.record",
        record: event.target.checked
    });
});
document.getElementById("nnpext-save").addEventListener("click", () => {
    console.log("save button clicked");
    B.runtime.sendMessage(extId, {
        action: "page.save"
    });
});
document.getElementById("nnpext-search").addEventListener("click", () => {
    console.log("search button clicked");
    B.runtime.sendMessage(extId, {
        action: "page.search"
    });
});

