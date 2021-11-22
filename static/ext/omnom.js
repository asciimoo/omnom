'use strict';

const br = chrome;
const is_ff = typeof InstallTrigger !== 'undefined';
const is_chrome = !is_ff;
const nodeTransformationFunctons = new Map([
    ['SCRIPT', (node) => node.remove()],
    ['LINK', transformLink],
    ['STYLE', transformStyle],
    ['IMG', transfromImg]
]);
let downloadedCount = 0;
let downloadCount = 0;
let debug = false;
let site_url = '';
let omnom_token = '';
let omnom_url = '';
let tabId = '';

function debugPopup(content) {
    if (is_chrome) {
        let win = window.open("", "omnomDebug", "menubar=yes,location=yes,resizable=yes,scrollbars=yes,status=yes");
        win.document.write(content);
    } else {
        document.getElementsByTagName('html')[0].innerHTML = content;
    }
    console.log(content);
}

/* ---------------------------------*
 * Diplay extension popup           *
 * ---------------------------------*/

function displayPopup() {
    document.querySelector("form").addEventListener("submit", saveBookmark);
    document.getElementById("omnom_options").addEventListener("click", function () {
        br.runtime.openOptionsPage(function () {
            window.close();
        });
    });
    setOmnomSettings().then(fillFormFields, renderError);
}

function getOmnomDataFromLocal() {
    return new Promise((resolve, reject) => {
        br.storage.local.get(['omnom_url', 'omnom_token', 'omnom_debug'], (data) => {
            data ? resolve(data) : reject('Could not get Data');
        });
    });
}

async function setOmnomSettings() {
    const omnomData = await getOmnomDataFromLocal().catch(renderError);
    omnom_url = omnomData.omnom_url || '';
    omnom_token = omnomData.omnom_token || '';
    debug = omnomData.omnom_debug || false;
    if (omnom_token == '') {
        return Promise.reject('Token not found. Specify it in the extension\'s options');
    }
    if (omnom_url == '') {
        return Promise.reject('Server URL not found. Specify it in the extension\'s option');
    }
    return Promise.resolve();
}

async function fillFormFields() {
    document.getElementById("omnom_url").innerHTML = "Server URL: " + omnom_url;
    document.querySelector("form").action = omnom_url + 'add_bookmark';
    document.getElementById("token").value = omnom_token;

    // fill url input field
    const tab = await queryTabsToPromise();
    if (tab) {
        document.getElementById('url').value = tab.url;
        site_url = tab.url;
        tabId = tab.tabId;
    }

    // fill title input field
    const title = await executeScriptToPromise(() => document.title);
    if (title && title[0]) {
        document.getElementById('title').value = title[0];
    }

    // fill notes input field
    const selection = await executeScriptToPromise(() => window.getSelection().toString());
    if (selection && selection[0]) {
        document.getElementById("notes").value = selection[0];
    }
}

/* ---------------------------------*
 * Save bookmarks                   *
 * ---------------------------------*/

function saveBookmark(e) {
    e.preventDefault();
    console.time('createSnapshot');
    createSnapshot().then(async (result) => {
        console.timeEnd('createSnapshot');
        if (debug) {
            debugPopup(result);
            return;
        }
        const form = new FormData(document.forms['add']);
        form.append("snapshot", result);
        fetch(omnom_url + 'add_bookmark', {
            method: 'POST',
            body: form,
            //headers: {
            //    'Content-type': 'application/json; charset=UTF-8'
            //}
        })
            .then(resp => {
                if (resp.status !== 200) {
                    throw Error(resp.statusText);
                }
                document.body.innerHTML = '<h1>Bookmark saved</h1>';
                // setTimeout(window.close, 2000);
            })
            .catch(err => {
                document.body.innerHTML = '<h1>Failed to save bookmark: ' + err + '</h1>';
            });
    });
}

/* ---------------------------------*
 * Create Snapshot                  *
 * ---------------------------------*/

async function createSnapshot() {
    const doc = await getDOM();
    let dom = document.createElement('html');
    dom.innerHTML = doc.html;
    for (let k in doc.attributes) {
        dom.setAttribute(k, doc.attributes[k]);
    }
    await walkDOM(dom, transformNode);
    if (!document.getElementById("favicon").value) {
        let favicon = await inlineFile(fullURL('/favicon.ico'));
        if (favicon) {
            document.getElementById('favicon').value = favicon;
            let faviconElement = document.createElement("style");
            faviconElement.setAttribute("rel", "icon");
            faviconElement.setAttribute("href", favicon);
            document.getElementsByTagName("head")[0].appendChild(faviconElement);
        }
    }
    return doc.doctype + dom.outerHTML;
}

async function transformNode(node) {
    if (node.nodeType !== Node.ELEMENT_NODE) {
        return;
    }
    const transformationFunction = nodeTransformationFunctons.get(node.nodeName);
    await rewriteAttributes(node);
    if (transformationFunction) {
        await transformationFunction(node);
        return;
    }
    return;
}

async function transformLink(node) {
    if (node.attributes.rel && node.attributes.rel.nodeValue.trim().toLowerCase() == "stylesheet") {
        if (!node.attributes.href) {
            return;
        }
        let cssHref = node.attributes.href.nodeValue;
        let style = document.createElement('style');
        return inlineFile(cssHref).then(async (cssText) => {
            style.innerHTML = await sanitizeCSS(cssText);
            node.parentNode.appendChild(style);
            node.remove();
        });
    }
    if ((node.getAttribute("rel") || '').trim().toLowerCase() == "icon" || (node.getAttribute("rel") || '').trim().toLowerCase() == "shortcut icon") {
        return inlineFile(node.href).then(async (favicon) => {
            document.getElementById('favicon').value = favicon;
            node.href = favicon;
        });
    }
}

async function transformStyle(node) {
    node.innerText = await sanitizeCSS(node.innerText);
    return;
}

async function transfromImg(node) {
    return inlineFile(node.getAttribute("src")).then(async (src) => {
        node.src = src;
    });
}

function getDOMData() {
    let html = document.getElementsByTagName('html')[0];
    let ret = {
        'html': html.cloneNode(true),
        'attributes': {},
        'title': '',
        'doctype': '',
    };
    if (document.doctype) {
        ret.doctype = new XMLSerializer().serializeToString(document.doctype);
    }
    if (document.getElementsByTagName('title').length > 0) {
        ret.title = document.getElementsByTagName('title')[0].innerText;
    }
    ret.html = ret.html.outerHTML;
    return ret;
}

async function getDOM() {
    const data = await executeScriptToPromise(getDOMData);
    if (data && data[0]) {
        return Promise.resolve(data[0]);
    } else {
        return Promise.reject('meh')
    }
}

async function rewriteAttributes(node) {
    for (let i = 0; i < node.attributes.length; i++) {
        let attr = node.attributes[i];
        if (attr.nodeName === undefined) {
            continue;
        }
        if (attr.nodeName.startsWith("on")) {
            attr.nodeValue = '';
        }
        if (attr.nodeValue.startsWith("javascript:")) {
            attr.nodeValue = '';
        }
        if (attr.nodeName == "href") {
            attr.nodeValue = fullURL(attr.nodeValue);
        }
        if (attr.nodeName == "style") {
            let sanitizedValue = await sanitizeCSS('a{' + attr.nodeValue + '}');
            attr.nodeValue = sanitizedValue.substr(4, sanitizedValue.length - 6);
        }
    }
}

async function inlineFile(url) {
    if (!url || (url || '').startsWith('data:')) {
        return url;
    }
    url = fullURL(url);
    console.log("fetching " + url);
    let options = {
        method: 'GET',
        cache: 'default',
    };
    if (debug) {
        options.cache = 'no-cache';
    }
    let request = new Request(url, options);
    downloadCount++;
    updateStatus();
    let obj = fetch(request, options).then(async function (response) {
        let contentType = response.headers.get("Content-Type");
        if (contentType.toLowerCase().search("text") != -1) {
            // TODO use charset of the response
            let body = await response.text();
            return body;
        }
        let buff = await response.arrayBuffer()
        let base64Flag = 'data:' + contentType + ';base64,';
        return base64Flag + arrayBufferToBase64(buff);
    }).catch(function (error) {
        console.log("MEH, network error", error)
    });
    downloadedCount++;
    updateStatus();
    return obj;
}

/* ---------------------------------*
 * Utility functions                *
 * ---------------------------------*/

function arrayBufferToBase64(buffer) {
    let binary = '';
    let bytes = [].slice.call(new Uint8Array(buffer));
    bytes.forEach((b) => binary += String.fromCharCode(b));

    return btoa(binary);
}

function fullURL(url) {
    return new URL(url, site_url).href
}

function parseCSS(styleContent) {
    let doc = document.implementation.createHTMLDocument("");
    let styleElement = document.createElement("style");

    styleElement.textContent = styleContent;
    // the style will only be parsed once it is added to a document
    doc.body.appendChild(styleElement);
    return styleElement.sheet.cssRules;
}

function executeScriptToPromise(functionToExecute) {
    return new Promise(resolve => {
        br.tabs.executeScript({
            code: `(${functionToExecute})()`
        },
            (data) => {
                resolve(data);
            });
    });
}

function renderError(errorMessage) {
    console.log(errorMessage);
    document.getElementById("omnom_content").innerHTML = `<h1>${errorMessage}</h1>`;
}

function queryTabsToPromise() {
    return new Promise(resolve => {
        br.tabs.query({ active: true, lastFocusedWindow: true }, ([tab]) => resolve(tab));
    });
}

async function walkDOM(node, func) {
    await func(node);
    let children = Array.from(node.childNodes);
    return Promise.allSettled(children.map(async (node, index) => {
        await walkDOM(node, func)
    }));
}

async function sanitizeCSS(rules) {
    if (typeof rules === 'string' || rules instanceof String) {
        rules = parseCSS(rules);
    }
    let cssMap = new Map();
    const rulesArray = Array.from(rules);
    await Promise.allSettled(rulesArray.map(async (r, index) => {
        // TODO handle other rule types
        // https://developer.mozilla.org/en-US/docs/Web/API/CSSRule/type

        // CSSStyleRule        
        if (r.type == 1) {
            const css = await sanitizeCSSRule(r);
            cssMap.set(index, css);
            // CSSimportRule
        } else if (r.type == 3) {
            // TODO handle import loops
            const sanitizedCSS = await sanitizeCSS(r.href);
            cssMap.set(index, sanitizedCSS);
            // r.href = '' need this here ?;
            // CSSMediaRule
        } else if (r.type == 4) {
            let sanitizedCSS = "@media " + r.media.mediaText + '{';
            for (let k2 in r.cssRules) {
                let r2 = r.cssRules[k2];
                sanitizedCSS += await sanitizeCSSRule(r2);
            }
            sanitizedCSS += '}';
            cssMap.set(index, sanitizedCSS);
            // CSSFontFaceRule
        } else if (r.type == 5) {
            let fontRule = await sanitizeCSSFontFace(r);
            if (fontRule) {
                cssMap.set(index, fontRule);
            } else {
                cssMap.set(index, r.cssText);
            }
        } else {
            console.log("MEEEH, unknown css rule type: ", r);
            return Promise.reject("MEEEH, unknown css rule type: ", r);
        }
    }));
    const sanitizedCSS = new Map([...cssMap.entries()].sort());
    const result = [...sanitizedCSS.values()].join('');
    return result;
}

async function sanitizeCSSRule(r) {
    // huh? how can r be undefined?....
    if (!r || !r.style) {
        return '';
    }
    // TODO handle ::xy { content: }
    await sanitizeCSSBgImage(r);
    await sanitizeCSSListStyleImage(r);
    return r.cssText;
}

async function sanitizeCSSBgImage(r) {
    let bgi = r.style.backgroundImage;
    if (bgi && bgi.startsWith('url("') && bgi.endsWith('")')) {
        let bgURL = fullURL(bgi.substring(5, bgi.length - 2));
        if (!bgURL.startsWith("data:")) {
            let inlineImg = await inlineFile(bgURL);
            if (inlineImg) {
                try {
                    r.style.backgroundImage = 'url("' + inlineImg + '")';
                } catch (error) {
                    console.log("failed to set background image: ", error);
                    r.style.backgroundImage = '';
                }
            } else {
                r.style.backgroundImage = '';
            }
        }
    }
}

async function sanitizeCSSListStyleImage(r) {
    let lsi = r.style.listStyleImage;
    if (lsi && lsi.startsWith('url("') && lsi.endsWith('")')) {
        let iURL = fullURL(lsi.substring(5, lsi.length - 2));
        if (!iURL.startsWith("data:")) {
            let inlineImg = await inlineFile(iURL);
            if (inlineImg) {
                try {
                    r.style.listStyleImage = 'url("' + inlineImg + '")';
                } catch (error) {
                    console.log("failed to set list-style-image:", error);
                    r.style.listStyleImage = '';
                }
            } else {
                r.style.listStyleImage = '';
            }
        }
    }
}

async function sanitizeCSSFontFace(r) {
    let src = r.style.getPropertyValue("src");
    let srcParts = src.split(/\s+/);
    let inlineImg;
    let changed = false;
    for (let i in srcParts) {
        let part = srcParts[i];
        if (part && part.startsWith('url("') && part.endsWith('")')) {
            let iURL = fullURL(part.substring(5, part.length - 2));
            if (!iURL.startsWith("data:")) {
                let inlineImg = await inlineFile(iURL);
                srcParts[i] = 'url("' + inlineImg + '")';
                let changed = true;
            }
        }
    }
    if (changed) {
        try {
            return '@font-face {' + r.style.cssText.replace(src, srcParts.join(" ")) + '}';
        } catch (error) {
            console.log("failed to set font-src:", error);
            r.style.src = '';
        }
    }
}

function updateStatus() {
    document.getElementById("omnom_status").innerHTML = '<h3>Downloading resources (' + downloadCount + '/' + downloadedCount + ')</h3>';
}

document.addEventListener('DOMContentLoaded', displayPopup);

/* ---------------------------------*
 * End of omnom code                *
 * ---------------------------------*/
