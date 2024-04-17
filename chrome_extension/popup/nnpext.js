

// constants and utilities
const B = chrome;

const extId = B.runtime.id;
const version = B.runtime.getManifest().version;

// load current state on UI
const loadStateOnUI = (delay) => {
    setTimeout(() => {
        B.storage.local.get("lastState", (res) => {
            console.log("loading lastState: ", res.lastState);
            if (res.lastState) {
                const currentState = res.lastState;
                document.getElementById("nnpext-jwt").value = currentState.jwt;
                document.getElementById("nnpext-memory").checked = currentState.allowTemporaryMemory;
                document.getElementById("nnpext-record").checked = currentState.recordNavigation;
                document.getElementById("nnpext-autosearch").checked = currentState.automaticSearch;
                document.getElementById("nnpext-version").value = version;
                document.getElementById("nnpext-msg").value = currentState.msg;
                document.getElementById("nnpext-msg").style.color = currentState.msgtype === "ok" ? "green" : "red";
            }
        });
    }, delay || 10);
};

loadStateOnUI();

// event listeners
document.getElementById("nnpext-login").addEventListener("click", (event) => {
    console.log("want to login");
    B.runtime.sendMessage(extId, {
        action: "jwt.read",
    })
    loadStateOnUI(30);
});
document.getElementById("nnpext-logout").addEventListener("click", () => {
    console.log("want to logout");
    B.runtime.sendMessage(extId, {
        action: "jwt.delete",
    });
    loadStateOnUI(10);
});
document.getElementById("nnpext-refresh-token").addEventListener("click", (event) => {
    console.log("want to refresh token");
    B.runtime.sendMessage(extId, {
        action: "jwt.refresh",
    })
    loadStateOnUI(30);
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
document.getElementById("nnpext-autosearch").addEventListener("change", (event) => {
    console.log("autosearch changed");
    B.runtime.sendMessage(extId, {
        action: "status.autosearch",
        autosearch: event.target.checked
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

