
const B = browser || chrome;
// read from manifest.json
let nnp_address = "";
if (!B.management.getSelf(function(info) {
    if (info.installType !== "development") {
        nnp_address = "https://nonoiseplease.com";
    } else {
        nnp_address = B.runtime.manifest.devserver;
    } 
}));
const succMsg = document.getElementById("succ-msg");
const errMsg = document.getElementById("err-msg");
const simpleCatch = (error) => {
    console.error(error);
    errMsg.style.display = "block";
};
const postRequest = (jwt, html, url, title) => {
    return {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            html: html,
            url: url,
            title: title,
            auth_code: jwt
        }),
    }
};
const tabQuery = {currentWindow: true, active: true};
const getDocument = { code: "document.body.innerHTML"};



document.getElementById("nnpext").addEventListener("click", () => {
    const jwt = document.getElementById("jwt").value;
    B.tabs.query(tabQuery).then((tabs) => {
        const currentTab = tabs[0];
        B.tabs.executeScript(currentTab.id, getDocument ).then((result) => {
            const htmlContent = result[0];
            fetch(
                nnp_address + "/api/page-manage/load", postRequest(jwt, htmlContent, currentTab.url, currentTab.title)
            ).then((response) => {
                if (response.ok) succMsg.style.display = "block";
            }).catch(simpleCatch); // fetch
        }).catch(simpleCatch); // get content
    }).catch(simpleCatch); // get tab
});


