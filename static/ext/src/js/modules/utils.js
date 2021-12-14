const browser = chrome;

const downloadStatus = {
    DOWNLOADING: 'downloading',
    DOWNLOADED: 'downloaded',
    FAILED: 'failed'
}

let downloadedCount = 0;
let downloadCount = 0;
let failedCount = 0;
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
        throw Error(res.statusText);
    }
    return res;
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

function parseCSS(styleContent) {
    const doc = document.implementation.createHTMLDocument('');
    const styleElement = document.createElement('style');

    styleElement.textContent = styleContent;
    // the style will only be parsed once it is added to a document
    doc.body.appendChild(styleElement);
    return styleElement.sheet.cssRules;
}

function queryTabsToPromise() {
    return new Promise(resolve => {
        browser.tabs.query({ active: true, currentWindow: true }, ([tab]) => resolve(tab));
    });
}

function renderError(errorMessage) {
    console.log(errorMessage);
    document.getElementById('omnom-content').innerHTML = `<h1>${errorMessage}</h1>`;
}

function renderSuccess(successMessage) {
    document.body.innerHTML = `<h1>${successMessage}</h1>`
    setTimeout(window.close, 2000);
}

function updateStatus(status) {
    switch (status) {
        case downloadStatus.DOWNLOADING: {
            downloadCount++
            break;
        }
        case downloadStatus.DOWNLOADED: {
            downloadedCount++
            break;
        }
        case downloadStatus.FAILED: {
            failedCount++;
            break;
        }
    }
    document.getElementById('omnom_status').innerHTML =
        `<h3>Downloading resources (${downloadCount}/${downloadedCount})</h3>
    <h3>Failed requests: ${failedCount}</h3>`;
}

async function setSiteUrl() {
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

async function inlineFile(url) {
    if (!url || (url || '').startsWith('data:')) {
        return url;
    }
    url = fullURL(url);
    console.log('fetching ', url);
    const options = {
        method: 'GET',
        cache: isDebug ? 'no-cache' : 'default',
    };
    const request = new Request(url, options);
    updateStatus(downloadStatus.DOWNLOADING);
    let hasError = false;
    const responseObj = await fetch(request, options)
        .then(checkStatus).catch((error) => {
            updateStatus(downloadStatus.FAILED);
            hasError = true;
        });
    if (hasError) {
        return '';
    }
    const contentType = responseObj.headers.get('Content-Type');
    updateStatus(downloadStatus.DOWNLOADED);
    if (contentType.toLowerCase().search('text') != -1) {
        // TODO use charset of the response        
        return await responseObj.text();
    }
    const buff = await responseObj.arrayBuffer()
    const base64Flag = `data:${contentType};base64,`;
    return `${base64Flag}${arrayBufferToBase64(buff)}`
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


export {
    arrayBufferToBase64,
    checkStatus,
    executeScriptToPromise,
    fullURL,
    absoluteURL,
    parseCSS,
    queryTabsToPromise,
    renderError,
    renderSuccess,
    updateStatus,
    walkDOM,
    getSiteUrl,
    inlineFile,
    setOmnomSettings,
    isDebug,
    getOmnomUrl,
    getOmnomToken,
    browser,
    downloadStatus
}