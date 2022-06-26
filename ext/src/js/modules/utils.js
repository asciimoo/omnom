// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

const browser = chrome;

let siteUrl = '';
let debug = false;
let omnomUrl = '';
let omnomToken = '';
let defaultPublic = false;

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

async function renderSuccess(successMessage, bookmarkInfo) {
    const omnomData = await getOmnomDataFromLocal().catch(renderError);
    const burl = absoluteURL(omnomData.omnom_url, bookmarkInfo.bookmark_url);
    const surl = absoluteURL(omnomData.omnom_url, bookmarkInfo.snapshot_url);
    document.getElementById('omnom-content').innerHTML = `
<h1>${successMessage}</h1>
<a href="${burl}">view bookmark</a><br />
<a href="${surl}">view snapshot</a>
    `;
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

async function setOmnomSettings() {
    const omnomData = await getOmnomDataFromLocal().catch(renderError);
    await setSiteUrl();
    setOmnomUrl(omnomData.omnom_url || '');
    setOmnomToken(omnomData.omnom_token || '');
    setDebug(omnomData.omnom_debug || false);
    setDefaultPublic(omnomData.omnom_public || false);
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
        browser.storage.local.get(['omnom_url', 'omnom_token', 'omnom_debug', 'omnom_public'], (data) => {
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

function isOmnomDefaultPublic() {
    return defaultPublic;
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

function setDefaultPublic(isPublic) {
    defaultPublic = isPublic;
}

function setOmnomToken(token) {
    omnomToken = token;
}

function copyScript(script) {
    const newScript = document.createElement('script');
    newScript.src = script.src;
    return newScript;
}

class UrlResolver {
    constructor(rootUrl) {
        this.url = rootUrl;
        this.hasBaseUrl = false;
    }
    resolve(url) {
        if (!url) {
            return this.url;
        }
        if (url.startsWith("data:")) {
            return url;
        }
        if (this.hasBaseUrl) {
            if (!url.startsWith("/") && url.search(/^[a-zA-Z]+:\/\//) == -1) {
                return this.url+url;
            }
        }
        return new URL(url, this.url).href;
    }
    setBaseUrl(url) {
        this.hasBaseUrl = true;
        this.url = this.resolve(url);
    }
}

function base64Decode(s) {
    return decodeURIComponent(escape(atob(s)));
}

function base64Encode(s) {
    return btoa(unescape(encodeURIComponent(s)));
}

async function sha256(data) {
    // __proto__.constructor.name
    if (data.__proto__.constructor.name == 'String') {
        data = new TextEncoder().encode(data);
    }
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray
          .map((bytes) => bytes.toString(16).padStart(2, '0'))
          .join('');
    return hashHex;
}

export {
    UrlResolver,
    absoluteURL,
    arrayBufferToBase64,
    base64Decode,
    base64Encode,
    browser,
    checkStatus,
    copyScript,
    executeScriptToPromise,
    fullURL,
    getOmnomToken,
    getOmnomUrl,
    getSiteUrl,
    isDebug,
    isOmnomDefaultPublic,
    queryTabsToPromise,
    renderError,
    renderSuccess,
    setOmnomSettings,
    setSiteUrl,
    sha256
}
