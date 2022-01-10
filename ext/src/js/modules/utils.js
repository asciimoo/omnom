const browser = chrome;

let siteUrl = '';
let debug = false;
let omnomUrl = '';
let omnomToken = '';

function arrayBufferToBase64(buffer) {
    let binary = '';
    const bytes = [].slice.call(new Uint8Array(buffer));
    bytes.forEach((b) => binary += String.fromCharCode(b));

    return btoa(binary);
}

function checkStatus(res) {
    if (!res.ok) {
        return Promise.reject(res);
    }
    return Promise.resolve(res);
}

function executeScriptToPromise(functionToExecute) {
    return new Promise(resolve => {
        browser.tabs.executeScript({
            code: `(${functionToExecute})()`
        },
            (data) => {
                resolve(data);
            });
    });
}

function fullURL(url) {
    return new URL(url, siteUrl).href
}

function absoluteURL(base, url) {
    return new URL(url, base).href
}

function queryTabsToPromise() {
    return new Promise(resolve => {
        browser.tabs.query({ active: true, currentWindow: true }, ([tab]) => resolve(tab));
    });
}

function renderError(errorMessage, error) {
    if (error) {
        error.json().then(data => console.log({ error, data }));
    }
    document.getElementById('omnom-content').innerHTML = `<h1>${errorMessage}</h1>`;
}

function renderSuccess(successMessage) {
    document.getElementById('omnom-content').innerHTML = `<h1>${successMessage}</h1>`
    // setTimeout(window.close, 2000);
}

async function setSiteUrl(url) {
    if (url) {
        siteUrl = url;
        return;
    }
    const tab = await queryTabsToPromise();
    if (tab) {
        siteUrl = tab.url;
    }
}

function getSiteUrl() {
    return siteUrl;
}

async function walkDOM(node, func) {
    await func(node);
    const children = [...node.childNodes];
    return Promise.allSettled(children.map(async (node) => {
        await walkDOM(node, func)
    }));
}


async function setOmnomSettings() {
    const omnomData = await getOmnomDataFromLocal().catch(renderError);
    await setSiteUrl();
    setOmnomUrl(omnomData.omnom_url || '');
    setOmnomToken(omnomData.omnom_token || '');
    setDebug(omnomData.omnom_debug || false);
    if (omnomToken == '') {
        return Promise.reject('Token not found. Specify it in the extension\'s options');
    }
    if (omnomUrl == '') {
        return Promise.reject('Server URL not found. Specify it in the extension\'s option');
    }
    return Promise.resolve();
}

function getOmnomDataFromLocal() {
    return new Promise((resolve, reject) => {
        browser.storage.local.get(['omnom_url', 'omnom_token', 'omnom_debug'], (data) => {
            data ? resolve(data) : reject('Could not get Data');
        });
    });
}

function getOmnomUrl() {
    return omnomUrl;
}

function getOmnomToken() {
    return omnomToken;
}

function isDebug() {
    return debug;
}

function setDebug(isDebug) {
    debug = isDebug;
}

function setOmnomUrl(url) {
    omnomUrl = url;
}

function setOmnomToken(token) {
    omnomToken = token;
}

function copyScript(script) {
    const newScript = document.createElement('script');
    newScript.src = script.src;
    return newScript;
}

export {
    arrayBufferToBase64,
    checkStatus,
    executeScriptToPromise,
    fullURL,
    absoluteURL,
    queryTabsToPromise,
    renderError,
    renderSuccess,
    walkDOM,
    getSiteUrl,
    setSiteUrl,
    setOmnomSettings,
    isDebug,
    getOmnomUrl,
    getOmnomToken,
    copyScript,
    browser
}
