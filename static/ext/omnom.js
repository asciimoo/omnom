'use strict';

const br = chrome;
const is_ff = typeof InstallTrigger !== 'undefined';
const is_chrome = !is_ff;
const nodeTransformFunctons = new Map([
    ['SCRIPT', (node) => node.remove()],
    ['LINK', transformLink],
    ['STYLE', transformStyle],
    ['IMG', transfromImg]
]);

const cssSanitizeFunctions = new Map([
    ['CSSStyleRule', sanitizeStyleRule],
    ['CSSImportRule', sanitizeImportRule],
    ['CSSMediaRule', sanitizeMediaRule],
    ['CSSFontFaceRule', sanitizeFontFaceRule],
    ['CSSPageRule', sanitizePageRule],
    ['CSSKeyframesRule', sanitizeKeyframesRule],
    ['CSSKeyframeRule', sanitizeKeyframeRule],
    ['CSSNamespaceRule', unknownRule], // XML only
    ['CSSCounterStyleRule', sanitizeCounterStyleRule],
    ['CSSSupportsRule', sanitizeSupportsRule],
    ['CSSDocumentRule', unknownRule], // FF only
    ['CSSFontFeatureValuesRule', unknownRule], // FF only
    ['CSSViewportRule', unknownRule], // IE only
])
const downloadStatus = {
    DOWNLOADING: 'downloading',
    DOWNLOADED: 'downloaded',
    FAILED: 'failed'
}

const styleNodes = new Map();

let downloadedCount = 0;
let downloadCount = 0;
let failedCount = 0;
let debug = false;
let site_url = '';
let omnom_token = '';
let omnom_url = '';
let styleIndex = 0;
let tags = [];
let templates = new Map();
let boundVars = { onoptions: false };

function debugPopup(content) {
    if (is_chrome) {
        const win = window.open('', 'omnomDebug', 'menubar=yes,location=yes,resizable=yes,scrollbars=yes,status=yes');
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
    setTemplates();
    evaluateTemplates();
    setEventListeners();
    setOmnomSettings().then(fillFormFields, renderError);
}

function setTemplates() {
    const templateElements = document.querySelectorAll('template');
    [...templateElements].forEach(template => templates.set(template.id, template));
}

function evaluateTemplates() {
    [...templates.values()].forEach(template => {
        const templateData = Object.keys(template.dataset);
        if (templateData.length) {
            const shouldShow = templateData.some(attribute => {
                const attributeValue = (template.dataset[attribute] === 'true');
                return boundVars.hasOwnProperty(attribute) && boundVars[attribute] === attributeValue
            });
            if (!shouldShow) {
                const nodeToRemove = [...template.parentNode.children].find(child => child.id === template.content.children[0].id);
                if (nodeToRemove) {
                    template.parentNode.removeChild(nodeToRemove);
                }
                return;
            }
        }
        template.parentNode.appendChild(template.content.cloneNode(true));
    })
}

function updateBoundVar(key, value) {
    if (boundVars.hasOwnProperty(key)) {
        boundVars[key] = value;
    }
    evaluateTemplates();
    setEventListeners();
}

function setEventListeners() {
    const tagsInput = document.getElementById('tags');
    tagsInput?.addEventListener('change', (event) => { addTag(event); tagsInput.value = '' });

    const closeButton = document.getElementById('close');
    closeButton?.addEventListener('click', closeHandler);

    const backButton = document.getElementById('back');
    backButton?.addEventListener('click', () => backHandler());

    const optionsButton = document.getElementById('omnom_options');
    optionsButton?.addEventListener('click', () => optionsHandler());

    const bookmarkForm = document.querySelector('form');
    bookmarkForm?.addEventListener('submit', saveBookmark);
}

function backHandler() {
    clearContentContainer();
    updateBoundVar('onoptions', false);
    fillFormFields();
}

function clearContentContainer() {
    const popupContent = document.getElementById('omnom-content');
    const templates = popupContent.querySelectorAll('template');
    [...popupContent.children].forEach(child => {
        if (![...templates].includes(child)) {
            popupContent.removeChild(child);
        }
    });
    return popupContent;
}

async function optionsHandler() {
    updateBoundVar('onoptions', true);

    const popupContent = document.getElementById('omnom-content');
    const optionsPageText = await fetch('./options.html').then(stream => stream.text());
    const p = new DOMParser();
    const optionsPageElement = p.parseFromString(optionsPageText, 'text/html');
    const template = optionsPageElement.querySelector('template')?.content.cloneNode(true);
    const script = optionsPageElement.querySelector('script');
    popupContent.appendChild(template);
    popupContent.appendChild(copyScript(script));
}

function closeHandler() {
    window.close();
}

function copyScript(script) {
    const newScript = document.createElement('script');
    newScript.src = script.src;
    return newScript;
}

function addTag(event) {
    const value = event.target.value;
    const tagChipContainer = document.getElementById('tag-chips');
    renderTag(value, tagChipContainer);
    tags.push(value);
}

function renderTag(value, parent) {
    const newChip = document.createElement('div');
    newChip.className = 'control chip-control';

    const chipContainer = document.createElement('span');
    chipContainer.className = 'tag is-rounded';
    chipContainer.innerText = value;

    const chipDelete = document.createElement('button');
    chipDelete.className = 'delete is-small';
    chipDelete.type = 'button';
    chipDelete.addEventListener('click', deleteTag.bind({}, newChip));

    chipContainer.appendChild(chipDelete);
    newChip.appendChild(chipContainer);
    parent.appendChild(newChip);
}

function deleteTag(chipElement) {
    const tagChipContainer = document.getElementById('tag-chips');
    tagChipContainer.removeChild(chipElement);
    tags = [...tagChipContainer.children].map(child => child.innerText);
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

function getOmnomDataFromLocal() {
    return new Promise((resolve, reject) => {
        br.storage.local.get(['omnom_url', 'omnom_token', 'omnom_debug'], (data) => {
            data ? resolve(data) : reject('Could not get Data');
        });
    });
}

async function fillFormFields() {
    document.querySelector('form').action = `${omnom_url}add_bookmark`;
    document.getElementById('token').value = omnom_token;

    // fill url input field
    const tab = await queryTabsToPromise();
    if (tab) {
        document.getElementById('url').value = tab.url;
        site_url = tab.url;
    }

    // fill title input field
    const title = await executeScriptToPromise(() => document.title);
    if (title && title[0]) {
        document.getElementById('title').value = title[0];
    }

    // fill notes input field
    const selection = await executeScriptToPromise(() => window.getSelection().toString());
    if (selection && selection[0]) {
        document.getElementById('notes').value = selection[0];
    }

    //fill tags
    const fragment = document.createDocumentFragment();
    const tagChips = document.getElementById('tag-chips');
    const tagChipsContainer = document.getElementById('tag-chips-container');
    fragment.appendChild(tagChips);
    tags.forEach(tag => {
        renderTag(tag, tagChips);
    });
    tagChipsContainer.appendChild(fragment);

}

/* ---------------------------------*
 * Save bookmarks                   *
 * ---------------------------------*/

async function saveBookmark(e) {
    e.preventDefault();
    console.time('createSnapshot');
    const snapshotData = await createSnapshot();
    console.timeEnd('createSnapshot');
    console.log(tags);
    if (debug) {
        debugPopup(snapshotData['dom']);
        return;
    }
    const form = new FormData(document.forms['add']);
    form.append('snapshot', snapshotData['dom']);
    form.append('snapshot_text', snapshotData['text']);
    form.set('tags', tags.join(','));
    await fetch(`${omnom_url}add_bookmark`, {
        method: 'POST',
        body: form,
        // headers: {
        //     'Content-type': 'application/json; charset=UTF-8'
        // }
    }).then(checkStatus).catch(err => renderError(`Failed to save bookmark:${err}`));
    document.body.innerHTML = '<h1>Bookmark saved</h1>';
    setTimeout(window.close, 2000);
}

/* ---------------------------------*
 * Create Snapshot                  *
 * ---------------------------------*/

async function createSnapshot() {
    const doc = await getDOM();
    const dom = document.createElement('html');
    dom.innerHTML = doc.html;
    for (const k in doc.attributes) {
        dom.setAttribute(k, doc.attributes[k]);
    }
    await walkDOM(dom, transformNode);
    setStyleNodes(dom);
    if (!document.getElementById('favicon').value) {
        const favicon = await inlineFile(fullURL('/favicon.ico'));
        if (favicon) {
            document.getElementById('favicon').value = favicon;
            const faviconElement = document.createElement('style');
            faviconElement.setAttribute('rel', 'icon');
            faviconElement.setAttribute('href', favicon);
            document.getElementsByTagName('head')[0].appendChild(faviconElement);
        }
    }
    return {
        'dom': `${doc.doctype}${dom.outerHTML}`,
        'text': dom.getElementsByTagName("body")[0].innerText,
    };
}

function setStyleNodes(dom) {
    const sortedStyles = new Map([...styleNodes.entries()].sort((e1, e2) => e1[0] - e2[0]));
    let parent;
    if (dom.getElementsByTagName("head")) {
        parent = dom.getElementsByTagName("head")[0];
    } else {
        parent = dom.documentElement;
    }
    sortedStyles.forEach(style => {
        parent.appendChild(style);
    });
}

async function getDOM() {
    const data = await executeScriptToPromise(getDOMData);
    if (data && data[0]) {
        return Promise.resolve(data[0]);
    } else {
        return Promise.reject('meh')
    }
}

function getDOMData() {
    const html = document.getElementsByTagName('html')[0];
    const ret = {
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
    [...html.attributes].forEach(attr => ret.attributes[attr.nodeName] = attr.nodeValue);
    let canvases = document.getElementsByTagName('canvas');
    if (canvases) {
        let canvasImages = [];
        for (let canvas of canvases) {
            let el = document.createElement("img");
            el.src = canvas.toDataURL();
            canvasImages.push(el);
        }
        let snapshotCanvases = ret.html.getElementsByTagName('canvas');
        for (let i in canvasImages) {
            let canvas = snapshotCanvases[i];
            canvas.parentNode.replaceChild(canvasImages[i], canvas);

        }
    }
    ret.html = ret.html.outerHTML;
    return ret;
}

async function transformNode(node) {
    if (node.nodeType !== Node.ELEMENT_NODE) {
        return;
    }
    const transformFunction = nodeTransformFunctons.get(node.nodeName);
    await rewriteAttributes(node);
    if (transformFunction) {
        await transformFunction(node);
        return;
    }
    return;
}

async function transformLink(node) {
    if (node.attributes.rel && node.attributes.rel.nodeValue.trim().toLowerCase() == 'stylesheet') {
        if (!node.attributes.href) {
            return;
        }
        const index = styleIndex++;
        const cssHref = fullURL(node.attributes.href.nodeValue);
        const style = document.createElement('style');
        const cssText = await inlineFile(cssHref);
        style.innerHTML = await sanitizeCSS(cssText, cssHref);
        styleNodes.set(index, style);
        node.remove();
    }
    if ((node.getAttribute('rel') || '').trim().toLowerCase() == 'icon' || (node.getAttribute('rel') || '').trim().toLowerCase() == 'shortcut icon') {
        const favicon = await inlineFile(node.href);
        document.getElementById('favicon').value = favicon;
        node.href = favicon;
    }
}

async function transformStyle(node) {
    const innerText = await sanitizeCSS(node.innerText, site_url);
    node.innerText = innerText;
}

async function transfromImg(node) {
    const src = await inlineFile(node.getAttribute('src'));
    node.src = src;
}

async function rewriteAttributes(node) {
    const nodeAttributeArray = [...node.attributes];
    return Promise.allSettled(nodeAttributeArray.map(async (attr) => {
        if (attr.nodeName.startsWith('on') || attr.nodeValue.startsWith('javascript:')) {
            attr.nodeValue = '';
        }
        if (attr.nodeName == 'href') {
            attr.nodeValue = fullURL(attr.nodeValue);
        }
        if (attr.nodeName == 'style') {
            const sanitizedValue = await sanitizeCSS(`a{${attr.nodeValue}}`, site_url);
            attr.nodeValue = sanitizedValue.substr(4, sanitizedValue.length - 6);
        }
        if (attr.nodeName == 'srcset') {
            let newParts = [];
            for (let s of attr.nodeValue.split(',')) {
                let srcParts = s.trim().split(' ');
                srcParts[0] = await inlineFile(srcParts[0]);
                newParts.push(srcParts.join(' '));
            }
            attr.nodeValue = newParts.join(', ');
        }
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
        cache: debug ? 'no-cache' : 'default',
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

async function sanitizeCSS(rules, baseURL) {
    if (typeof rules === 'string' || rules instanceof String) {
        rules = parseCSS(rules);
    }
    const cssMap = new Map();
    const rulesArray = [...rules];
    await Promise.allSettled(rulesArray.map(async (r, index) => {
        const sanitizeFunction = cssSanitizeFunctions.get(r.constructor.name);
        if (sanitizeFunction) {
            const css = await sanitizeFunction(r, baseURL).catch(err => console.log(err));
            cssMap.set(index, css);
        } else {
            unknownRule(r, baseURL);
        }
    }));
    const sortedCss = new Map([...cssMap.entries()].sort((e1, e2) => e1[0] - e2[0]));
    const result = [...sortedCss.values()].join('');
    return result;
}

/* ---------------------------------*
 * Sanitize css                     *
 * ---------------------------------*/

async function sanitizeStyleRule(rule, baseURL) {
    return await sanitizeCSSRule(rule, baseURL);
}

async function sanitizeImportRule(rule, baseURL) {
    // TODO handle import loops
    let href = absoluteURL(baseURL, rule.href);
    return await sanitizeCSS(await inlineFile(href), href);
}

async function sanitizeMediaRule(rule, baseURL) {
    const cssRuleArray = [...rule.cssRules];
    let cssResult = '';
    await Promise.allSettled(cssRuleArray.map(async (r, index) => {
        const css = await sanitizeCSSRule(r, baseURL);
        cssResult += css;
    }));
    return `@media ${rule.media.mediaText}{${cssResult}}`;
}

async function sanitizeFontFaceRule(rule, baseURL) {
    const fontRule = await sanitizeCSSFontFace(rule, baseURL);
    return fontRule ? fontRule : rule.cssText;
}

async function sanitizePageRule(rule, baseURL) {
    return rule.cssText;
}

async function sanitizeKeyframesRule(rule, baseURL) {
    let cssResult = await sanitizeCSS(rule.cssRules, baseURL);
    return `@keyframes ${rule.name}{${cssResult}}`;
}

async function sanitizeKeyframeRule(rule, baseURL) {
    return await sanitizeStyleRule(rule);
}

async function sanitizeSupportsRule(rule, baseURL) {
    let cssResult = await sanitizeCSS(rule.cssRules, baseURL);
    return `@supports ${rule.conditionText}{${cssResult}}`;
}

async function sanitizeCounterStyleRule(rule, baseURL) {
    return rule.cssText;
}

async function unknownRule(rule) {
    console.log('MEEEH, unknown css rule type: ', rule);
    return Promise.reject('MEEEH, unknown css rule type: ', rule);
}

async function sanitizeCSSRule(r, baseURL) {
    // huh? how can r be undefined?....
    if (!r || !r.style) {
        return '';
    }
    // TODO handle ::xy { content: }
    await sanitizeCSSBgImage(r, baseURL);
    await sanitizeCSSListStyleImage(r, baseURL);
    return r.cssText;
}

async function sanitizeCSSBgImage(r, baseURL) {
    const bgi = r.style.backgroundImage;
    if (bgi && bgi.startsWith('url("') && bgi.endsWith('")')) {
        const bgURL = absoluteURL(baseURL, bgi.substring(5, bgi.length - 2));
        if (!bgURL.startsWith('data:')) {
            const inlineImg = await inlineFile(bgURL);
            if (inlineImg) {
                try {
                    r.style.backgroundImage = `url('${inlineImg}')`;
                } catch (error) {
                    console.log('failed to set background image: ', error);
                    r.style.backgroundImage = '';
                }
            } else {
                r.style.backgroundImage = '';
            }
        }
    }
}

async function sanitizeCSSListStyleImage(r, baseURL) {
    const lsi = r.style.listStyleImage;
    if (lsi && lsi.startsWith('url("') && lsi.endsWith('")')) {
        const iURL = absoluteURL(baseURL, lsi.substring(5, lsi.length - 2));
        if (!iURL.startsWith('data:')) {
            const inlineImg = await inlineFile(iURL);
            if (inlineImg) {
                try {
                    r.style.listStyleImage = `url('${inlineImg}')`;
                } catch (error) {
                    console.log('failed to set list-style-image:', error);
                    r.style.listStyleImage = '';
                }
            } else {
                r.style.listStyleImage = '';
            }
        }
    }
}

async function sanitizeCSSFontFace(r, baseURL) {
    const src = r.style.getPropertyValue('src');
    const srcParts = src.split(/\s+/);
    let changed = false;
    for (const i in srcParts) {
        const part = srcParts[i];
        if (part && part.startsWith('url("') && part.endsWith('")')) {
            const iURL = absoluteURL(baseURL, part.substring(5, part.length - 2));
            if (!iURL.startsWith('data:')) {
                const inlineImg = await inlineFile(absoluteURL(baseURL, iURL));
                srcParts[i] = `url('${inlineImg}')`;
                changed = true;
            }
        }
    }
    if (changed) {
        try {
            return `@font-face {${r.style.cssText.replace(src, srcParts.join(' '))}}`;
        } catch (error) {
            console.log('failed to set font-src:', error);
            r.style.src = '';
        }
    }
}

/* ---------------------------------*
 * Utility functions                *
 * ---------------------------------*/

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
        br.tabs.executeScript({
            code: `(${functionToExecute})()`
        },
            (data) => {
                resolve(data);
            });
    });
}

function fullURL(url) {
    return new URL(url, site_url).href
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
        br.tabs.query({ active: true, currentWindow: true }, ([tab]) => resolve(tab));
    });
}

function renderError(errorMessage) {
    console.log(errorMessage);
    document.getElementById('omnom-content').innerHTML = `<h1>${errorMessage}</h1>`;
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

async function walkDOM(node, func) {
    await func(node);
    const children = [...node.childNodes];
    return Promise.allSettled(children.map(async (node) => {
        await walkDOM(node, func)
    }));
}

document.addEventListener('DOMContentLoaded', displayPopup);

/* ---------------------------------*
 * End of omnom code                *
 * ---------------------------------*/
