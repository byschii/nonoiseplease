


const postRequest = (jwt, html, url, title) => {
    // log body with shortened html
    console.debug("body: ", {
        html: html,
        url: url,
        title: title,
    });

    return {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Authorization: jwt
        },
        body: JSON.stringify({
            html: html,
            url: url,
            title: title
        }),
    }
};
const sendPage = async (nnp_address, jwt, html, url, title) => {
    console.log("sending page");
    if (!jwt) {
        console.error("no jwt");
        return false;
    }
    let res = await fetch(
        nnp_address + "/api/page-manage/load", postRequest(jwt, html, url, title)
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

/*
    right cannot apply filter to the search

    * @param {string} nnp_address
    * @param {string} jwt
    * @param {string} query - search query written on the search engine
    * @returns {Promise<[]>}
*/
const searchPage = async (nnp_address, jwt, query) => {
    console.log("searching page");

    let searchParams = new URLSearchParams({
        query: query,
        // categories: categories.join(',')
    })

    let resp = await fetch(
        nnp_address + "/api/search?"+ searchParams, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                Authorization: jwt
            }
        }).then(response => response.json());

    return resp;
}