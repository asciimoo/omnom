'use strict';

const br = chrome;
const is_ff = typeof InstallTrigger !== 'undefined';
const is_chrome = !is_ff;

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

function saveBookmark(e) {
    e.preventDefault();
    createSnapshot().then(async (result) => {
        if (debug) {
            debugPopup(result);
            return;
        }
        let form = new FormData(document.forms['add']);
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

function displayPopup() {
    document.querySelector("form").addEventListener("submit", saveBookmark);
    document.getElementById("omnom_options").addEventListener("click", function () {
        br.runtime.openOptionsPage(function () {
            window.close();
        });
    });
    setOmnomSettings().then(fillFormFields, renderError);
}

function renderError(errorMessage) {
    console.log(errorMessage);
    document.getElementById("omnom_content").innerHTML = `<h1>${errorMessage}</h1>`;
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

function fillFormFields() {
    document.getElementById("omnom_url").innerHTML = "Server URL: " + omnom_url;
    document.querySelector("form").action = omnom_url + 'add_bookmark';
    document.getElementById("token").value = omnom_token;
    // fill url input field
    br.tabs.query({ active: true, lastFocusedWindow: true }, (tabs) => {
        document.getElementById('url').value = tabs[0].url;
        site_url = tabs[0].url;
        tabId = tabs[0].tabId;
    });
    // fill title input field
    br.tabs.executeScript({
        code: 'document.title;'
    }, (title) => {
        if (title && title[0]) {
            document.getElementById('title').value = title[0];
        }
    });
    // fill notes input field
    br.tabs.executeScript({
        code: "window.getSelection().toString();"
    }, function (selection) {
        if (selection && selection[0]) {
            document.getElementById("notes").value = selection[0];
        }
    });
}

function rewriteAttributes(node) {
    for (let i = 0; i < node.attributes.length; i++) {
        let attr = node.attributes[i];
        if (attr.nodeName === undefined) {
            continue;
        }
        if (attr.nodeName.startsWith("on")) {
            attr.nodeValue = '';
            //} else if(attr.nodeName.startsWith("data-")) {
            //    attr.nodeValue = '';
        } else if (attr.nodeValue.trim().toLowerCase().startsWith("javascript:")) {
            attr.nodeValue = '';
        }
        if (attr.nodeName == "href") {
            attr.nodeValue = fullURL(attr.nodeValue);
        }
    }
}

function getDOMData() {
    function fullURL(url) {
        return new URL(url, window.location.origin).href
    }
    function getCSSText(obj) {
        if (obj.cssText) {
            return obj.cssText;
        }
        let text = '';
        for (let i = 0; i < obj.length; i++) {
            let key = obj.item(i);
            text += key + ':' + obj[key] + ';';
        }
        return text;
    }
    function walkDOM(node, func) {
        func(node);
        let children = node.childNodes;
        for (let i = 0; i < children.length; i++) {
            walkDOM(children[i], func);
        }
    }
    let html = document.getElementsByTagName('html')[0];
    let ret = {
        'html': html.cloneNode(true),
        'attributes': {},
        'title': '',
        'doctype': '',
    };
    for (let k in html.attributes) {
        let a = html.attributes[k];
        ret.attributes[a.nodeName] = a.nodeValue;
    }
    if (document.doctype) {
        ret.doctype = new XMLSerializer().serializeToString(document.doctype);
    }
    if (document.getElementsByTagName('title').length > 0) {
        ret.title = document.getElementsByTagName('title')[0].innerText;
    }
    let nodesToRemove = [];
    walkDOM(html, async function (n) {
        if (n.nodeName == 'SCRIPT') {
            nodesToRemove.push(n);
            return;
        }
    });
    for (i in nodesToRemove) {
        nodesToRemove[i].remove();
    }
    ret.html = ret.html.outerHTML;
    return ret;
}

//
async function getDOM() {
    const data = await br.tabs.executeScript({
        code: `(${getDOMData})()`
    });

    if (data && data[0]) {
        return Promise.resolve(data[0]);
    } else {
        return Promise.reject('meh')
    }
}

function getDomNew() {
    br.scripting.executeScript({
        target: { tabId },
        func: () => getDOMData()
    }, ([data]) => data ? console.log('yay data: ', data) : console.log('meeh data: ', data));
}
async function createSnapshot() {
    let doc = await getDOM();
    let dom = document.createElement('html');
    dom.innerHTML = doc.html;
    for (let k in doc.attributes) {
        dom.setAttribute(k, doc.attributes[k]);
    }
    let nodesToAppend = [];
    let nodesToRemove = [];
    await walkDOM(dom, async function (node) {
        if (node.nodeType !== Node.ELEMENT_NODE) {
            return;
        }
        if (node.nodeName == 'SCRIPT') {
            node.remove();
            return;
        }
        await rewriteAttributes(node);
        if (node.nodeName == 'LINK') {
            if (node.attributes.rel && node.attributes.rel.nodeValue.trim().toLowerCase() == "stylesheet") {
                if (!node.attributes.href) {
                    console.log("no css href found", node);
                    return;
                }
                let cssHref = node.attributes.href.nodeValue;
                let style = document.createElement('style');
                let cssText = await inlineFile(cssHref);
                style.innerHTML = await sanitizeCSS(cssText);
                nodesToAppend.push([style, node.parentNode]);
                nodesToRemove.push(node);
                return;
            }
            if ((node.getAttribute("rel") || '').trim().toLowerCase() == "icon" || (node.getAttribute("rel") || '').trim().toLowerCase() == "shortcut icon") {
                let favicon = await inlineFile(node.href);
                document.getElementById('favicon').value = favicon;
                node.href = favicon;
                return;
            }
        }
        if (node.nodeName == 'STYLE') {
            node.innerText = await sanitizeCSS(node.innerText);
            return;
        }
        if (node.nodeName == 'IMG') {
            node.src = await inlineFile(node.getAttribute("src"));
            return;
        }
    });
    for (let i in nodesToAppend) {
        let elem = nodesToAppend[i][0]
        let parent = nodesToAppend[i][1];
        parent.appendChild(elem);
    }
    for (let i in nodesToRemove) {
        nodesToRemove[i].remove();
    }
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

async function walkDOM(node, func) {
    await func(node);
    let children = node.childNodes;
    for (let i = 0; i < children.length; i++) {
        await walkDOM(children[i], func);
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
    let obj = await fetch(request, options).then(async function (response) {
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

async function sanitizeCSS(rules) {
    if (typeof rules === 'string' || rules instanceof String) {
        rules = parseCSS(rules);
    }
    let sanitizedCSS = '';
    for (let k in rules) {
        let r = rules[k];
        // TODO handle other rule types
        // https://developer.mozilla.org/en-US/docs/Web/API/CSSRule/type

        // CSSStyleRule
        if (r.type == 1) {
            sanitizedCSS += await sanitizeCSSRule(r);
            // CSSimportRule
        } else if (r.type == 3) {
            // TODO handle import loops
            sanitizedCSS += await sanitizeCSS(r.href);
            // r.href = '' need this here ?;
            // CSSMediaRule
        } else if (r.type == 4) {
            sanitizedCSS += "@media " + r.media.mediaText + '{';
            for (let k2 in r.cssRules) {
                let r2 = r.cssRules[k2];
                sanitizedCSS += await sanitizeCSSRule(r2);
            }
            sanitizedCSS += '}';
            // CSSFontFaceRule
        } else if (r.type == 5) {
            let fontRule = await sanitizeCSSFontFace(r);
            if (fontRule) {
                sanitizedCSS += fontRule;
            } else {
                sanitizedCSS += r.cssText;
            }
        } else {
            console.log("MEEEH, unknown css rule type: ", r);
        }
    }
    return sanitizedCSS
}

async function sanitizeCSSRule(r) {
    // huh? how can r be undefined?....
    if (!r || !r.style) {
        return '';
    }
    console.log(r);
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
