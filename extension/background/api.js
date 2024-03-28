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
const sendPage = async (nnp_address, userId, extensionToken, html, url, title) => {
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


const refreshToken = async (nnp_address, jwt) => {
    console.log("refreshing token");

    let res = await fetch(
        nnp_address + "/api/collections/users/auth-refresh", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: jwt
            },
        })
        .then((response) => {
            if (response.ok) {
                return response.json();
            }
            return null;
        })

    return res;
};

