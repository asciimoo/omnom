// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

const browser = chrome;

let siteUrl = '';
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
    //let [tab] = await chrome.tabs.query({ active: true, currentWindow: true }
    return new Promise(resolve => {
        browser.tabs.query({currentWindow: true, active: true}).then(tabs => {
            browser.scripting.executeScript({
                target: { tabId: tabs[0].id },
                func: functionToExecute
            },
            (data) => {
                resolve(data);
            });
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
    if(document.getElementById('status')) {
        document.getElementById('status').innerHTML = errorMessage;
    } else {
        document.getElementById('omnom-content').innerHTML = `<h1 id="status" class="error">${errorMessage}</h1>`;
    }
}

function openOnNewTab(ev) {
    chrome.tabs.create({url: this.getAttribute('href')});
    ev.preventDefault();
    return false;
}

async function renderSuccess(successMessage, bookmarkInfo) {
    if(bookmarkInfo) {
        const omnomData = await getOmnomDataFromLocal().catch(renderError);
        const burl = absoluteURL(omnomData.omnom_url, bookmarkInfo.bookmark_url);
        const surl = absoluteURL(omnomData.omnom_url, bookmarkInfo.snapshot_url);
        document.getElementById('omnom-content').innerHTML = `
    <h1 id="status" class="success">${successMessage}</h1>
    <a href="${burl}" class="open-on-new-tab">view bookmark</a><br />
    <a href="${surl}" class="open-on-new-tab">view snapshot</a>
        `;
        for(let el of document.querySelectorAll(".open-on-new-tab")) {
            el.addEventListener('click', openOnNewTab);
        }
    } else {
        document.getElementById('omnom-content').innerHTML = `
    <h1 id="status" class="success">${successMessage}</h1>`
    }
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
        browser.storage.local.get(['omnom_url', 'omnom_token', 'omnom_public'], (data) => {
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

async function validateOptions(serverUrl, token) {
    let formData = new FormData();
    formData.append('token', token);
    return fetch(serverUrl + 'check_token', {
        'method': 'POST',
        'body': formData,
    });
}

function getOmnomSettings(cb) {
    chrome.storage.local.get(['omnom_url', 'omnom_token', 'omnom_public'], function(data) {
        if(!data['omnom_url']) {
            data['omnom_url'] = '';
        }
        if(!data['omnom_token']) {
            data['omnom_token'] = '';
        }
        if(!data['omnom_public']) {
            data['omnom_public'] = false;
        }
        cb(data);
    });
}

function extractVisibleTextBlocks(el) {
    if(!el) {
        el = document.body;
    }
    const sectionElements = [
        'ARTICLE',
        'ASIDE',
        'BLOCKQUOTE',
        'DIV',
        'DL',
        'DT',
        'FIGURE',
        'FOOTER',
        'H1',
        'H2',
        'H3',
        'H4',
        'H5',
        'H6',
        'LI',
        'MAIN',
        'NAV',
        'P',
        'SECTION',
        'TD',
        'TH',
    ];
    function skipInvisible(n) {
        if(n.nodeType != Node.ELEMENT_NODE) {
            return NodeFilter.FILTER_ACCEPT;
        }
        const style = window.getComputedStyle(n);
        const rect = n.getBoundingClientRect();
        // TODO calculate valid height by font size: rect.height < (style.fontSize.replace('px', '')/2))
        if(rect.width < 5 || rect.height < 5) {
            return NodeFilter.FILTER_REJECT;
        }
        if(style.display == 'none' || style.visibility == 'hidden' || style.opacity == '0') {
            return NodeFilter.FILTER_REJECT;
        }
        return NodeFilter.FILTER_ACCEPT;
    }
    const walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT, skipInvisible);
    let curTexts = [];
    let parentNode = el.tagName;
    let texts = [];
    while(walker.nextNode()) {
        let n = walker.currentNode;
        if(n.nodeType == Node.ELEMENT_NODE) {
            if(sectionElements.includes(n.tagName) && curTexts.length > 0) {
                let ct = curTexts.join('').replace(/\s+/g, " ").trim();
                if(ct) {
                    texts.push(ct);
                }
                curTexts = [];
            }
        } else if(n.nodeType == Node.TEXT_NODE) {
            curTexts.push(n.nodeValue);
        }
    }

    if(curTexts.length > 0) {
        let ct = curTexts.join('').replace(/\s+/g, " ").trim();
        if(ct) {
            texts.push(ct);
        }
    }

    return texts;
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
    extractVisibleTextBlocks,
    fullURL,
    getOmnomSettings,
    getOmnomToken,
    getOmnomUrl,
    getSiteUrl,
    isOmnomDefaultPublic,
    queryTabsToPromise,
    renderError,
    renderSuccess,
    setOmnomSettings,
    setSiteUrl,
    sha256,
    validateOptions
}
