// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Utility functions for the Omnom browser extension.
 * Provides URL handling, encoding/decoding, storage access, and browser API wrappers.
 */

/**
 * Browser API reference (Chrome)
 * @type {Object}
 */
const browser = chrome;

/**
 * Current site URL being processed
 * @type {string}
 */
let siteUrl = '';

/**
 * Omnom server URL
 * @type {string}
 */
let omnomUrl = '';

/**
 * Omnom authentication token
 * @type {string}
 */
let omnomToken = '';

/**
 * Default bookmark visibility setting
 * @type {boolean}
 */
let defaultPublic = false;

/**
 * Converts an ArrayBuffer to a base64-encoded string
 * @param {ArrayBuffer} buffer - The buffer to convert
 * @returns {string} Base64-encoded string
 */
function arrayBufferToBase64(buffer) {
    let binary = '';
    const bytes = [].slice.call(new Uint8Array(buffer));
    bytes.forEach((b) => binary += String.fromCharCode(b));

    return btoa(binary);
}

/**
 * Checks if an HTTP response is successful
 * @param {Response} res - The fetch response object
 * @returns {Promise<Response>} Resolved promise if OK, rejected if not
 */
function checkStatus(res) {
    if (!res.ok) {
        return Promise.reject(res);
    }
    return Promise.resolve(res);
}

/**
 * Executes a script in the active tab and returns result as a promise
 * @param {Function} functionToExecute - The function to execute in the tab
 * @returns {Promise} Promise that resolves with script execution result
 */
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

/**
 * Converts a relative URL to absolute URL using the current site URL
 * @param {string} url - The URL to convert
 * @returns {string} Absolute URL
 */
function fullURL(url) {
    return new URL(url, siteUrl).href
}

/**
 * Converts a relative URL to absolute URL using a specified base URL
 * @param {string} base - The base URL
 * @param {string} url - The URL to convert
 * @returns {string} Absolute URL
 */
function absoluteURL(base, url) {
    return new URL(url, base).href
}

/**
 * Queries for the active tab and returns it as a promise
 * @returns {Promise<Tab>} Promise resolving to the active tab
 */
function queryTabsToPromise() {
    return new Promise(resolve => {
        browser.tabs.query({ active: true, currentWindow: true }, ([tab]) => resolve(tab));
    });
}

/**
 * Renders an error message in the UI
 * @param {string} errorMessage - The error message to display
 * @param {Response} [error] - Optional response object for additional error info
 */
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

/**
 * Opens a link in a new tab (event handler)
 * @param {Event} ev - The click event
 * @returns {boolean} Always returns false to prevent default
 */
function openOnNewTab(ev) {
    chrome.tabs.create({url: this.getAttribute('href')});
    ev.preventDefault();
    return false;
}

/**
 * Renders a success message in the UI with optional bookmark links
 * @async
 * @param {string} successMessage - The success message to display
 * @param {Object} [bookmarkInfo] - Optional bookmark information with URLs
 */
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

/**
 * Sets the current site URL
 * @async
 * @param {string} [url] - The URL to set, or queries active tab if not provided
 */
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

/**
 * Gets the current site URL
 * @returns {string} The current site URL
 */
function getSiteUrl() {
    return siteUrl;
}

/**
 * Loads and sets Omnom settings from storage
 * @async
 * @returns {Promise} Resolves on success, rejects with error message on failure
 */
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

/**
 * Retrieves Omnom settings from local storage
 * @returns {Promise<Object>} Promise resolving to settings object
 */
function getOmnomDataFromLocal() {
    return new Promise((resolve, reject) => {
        browser.storage.local.get(['omnom_url', 'omnom_token', 'omnom_public'], (data) => {
            data ? resolve(data) : reject('Could not get Data');
        });
    });
}

/**
 * Gets the Omnom server URL
 * @returns {string} The Omnom server URL
 */
function getOmnomUrl() {
    return omnomUrl;
}

/**
 * Gets the Omnom authentication token
 * @returns {string} The authentication token
 */
function getOmnomToken() {
    return omnomToken;
}

/**
 * Checks if bookmarks are public by default
 * @returns {boolean} True if default is public
 */
function isOmnomDefaultPublic() {
    return defaultPublic;
}

/**
 * Sets the Omnom server URL
 * @param {string} url - The server URL
 */
function setOmnomUrl(url) {
    omnomUrl = url;
}

/**
 * Sets the default bookmark visibility
 * @param {boolean} isPublic - Whether bookmarks are public by default
 */
function setDefaultPublic(isPublic) {
    defaultPublic = isPublic;
}

/**
 * Sets the Omnom authentication token
 * @param {string} token - The authentication token
 */
function setOmnomToken(token) {
    omnomToken = token;
}

/**
 * Creates a copy of a script element
 * @param {HTMLScriptElement} script - The script element to copy
 * @returns {HTMLScriptElement} New script element with same source
 */
function copyScript(script) {
    const newScript = document.createElement('script');
    newScript.src = script.src;
    return newScript;
}

/**
 * Class for resolving relative URLs to absolute URLs
 * @class
 */
class UrlResolver {
    /**
     * Creates a UrlResolver instance
     * @param {string} rootUrl - The root URL for resolution
     */
    constructor(rootUrl) {
        this.url = rootUrl;
        this.hasBaseUrl = false;
    }
    
    /**
     * Resolves a relative URL to absolute
     * @param {string} url - The URL to resolve
     * @returns {string} Absolute URL
     */
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
    
    /**
     * Sets a new base URL for resolution
     * @param {string} url - The new base URL
     */
    setBaseUrl(url) {
        this.hasBaseUrl = true;
        this.url = this.resolve(url);
    }
}

/**
 * Decodes a base64 string with UTF-8 support
 * @param {string} s - Base64 string to decode
 * @returns {string} Decoded string
 */
function base64Decode(s) {
    return decodeURIComponent(escape(atob(s)));
}

/**
 * Encodes a string to base64 with UTF-8 support
 * @param {string} s - String to encode
 * @returns {string} Base64-encoded string
 */
function base64Encode(s) {
    return btoa(unescape(encodeURIComponent(s)));
}

/**
 * Computes SHA-256 hash of data
 * @async
 * @param {string|ArrayBuffer} data - Data to hash
 * @returns {Promise<string>} Hex string of hash
 */
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

/**
 * Validates Omnom server URL and token by checking with server
 * @async
 * @param {string} serverUrl - The server URL to validate
 * @param {string} token - The authentication token to validate
 * @returns {Promise<Response>} Fetch response from validation endpoint
 */
async function validateOptions(serverUrl, token) {
    let formData = new FormData();
    formData.append('token', token);
    return fetch(serverUrl + 'check_token', {
        'method': 'POST',
        'body': formData,
    });
}

/**
 * Retrieves Omnom settings from storage and calls callback
 * @param {Function} cb - Callback function to receive settings
 */
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

/**
 * Extracts visible text blocks from a DOM element
 * @param {HTMLElement} [el] - Element to extract text from (defaults to document.body)
 * @returns {Array<string>} Array of visible text blocks
 */
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
