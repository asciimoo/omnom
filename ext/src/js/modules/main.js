// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Main popup functionality for the Omnom browser extension.
 * Handles bookmark creation, form management, and snapshot processing.
 */

import { Document } from "./document";
import { renderProgressBar, destroyProgressBar } from './file-download';
import { getDomData } from "./get-dom-data";
import { createSnapshot } from './snapshot';
import { TagInputController } from './tag-input';
import {
    browser as br,
    checkStatus,
    copyScript,
    executeScriptToPromise,
    extractVisibleTextBlocks,
    getOmnomToken,
    getOmnomUrl,
    getSiteUrl,
    isOmnomDefaultPublic,
    renderError,
    renderSuccess,
    setOmnomSettings,
    validateOptions
} from './utils';

/**
 * Map of message handlers for content script communication
 * @type {Map<string, Function>}
 */
const messageHandlers = new Map([
    ['pong', handlePongMessage],
    ['domData', handleDomDataMessage]
]);

/**
 * Browser detection flags
 */
const is_ff = typeof InstallTrigger !== 'undefined';
const is_chrome = !is_ff;

/**
 * Controller for tag input functionality
 * @type {TagInputController|null}
 */
let tagInput = null;

/**
 * Map of template elements for dynamic UI rendering
 * @type {Map<string, HTMLTemplateElement>}
 */
let templates = new Map();

/**
 * State variables for template visibility
 * @type {Object}
 */
let boundVars = { onoptions: false, onafterdownload: false, onmain: true };

/**
 * Main content container element
 * @type {HTMLElement|null}
 */
let contentContainer = null;

/**
 * Form data for bookmark submission
 * @type {FormData|null}
 */
let form = null;

/**
 * Communication channel with content script
 * @type {Object|null}
 */
let commChan = null;

/**
 * Number of pages (including iframes) being processed
 * @type {number}
 */
let numberOfPages = 0;

/**
 * Main document object
 * @type {Document|null}
 */
let doc = null;

/**
 * Array of iframe documents
 * @type {Array<Document>}
 */
let iframes = [];

/**
 * Maximum size for blob uploads (7MB)
 * @type {number}
 */
let blobSizeLimit = 7 * 1024 * 1024; // 7Mb

/* ---------------------------------*
 * Content js messaging             *
 * ---------------------------------*/

/**
 * Sets up communication channel with content scripts
 * @param {Object} [msg] - Optional initial message to send
 */
function setupComms(msg) {
    br.tabs.query({
        active: true,
        currentWindow: true
    }, tabs => {
        let tab = tabs[0];
        commChan = br.tabs.connect(
            tab.id,
            {name: "omnom"}
        );
        commChan.onMessage.addListener((msg, sender) => {
            if(sender.name != "omnom") {
                return false;
            }
            const msgHandler = messageHandlers.get(msg.type);
            if(msgHandler) {
                msgHandler(msg);
            } else {
                console.log("unknown message: ", msg);
            }
            return true;
        });
        if (msg) {
            commChan.postMessage(msg);
            if(chrome.runtime.lastError) {
                console.log(`Failed to deliver ${msg.type} message`, chrome.runtime.lastError);
            }
        }
    });
}

/**
 * Handles pong response from content scripts
 * @async
 * @param {Object} msg - The pong message
 */
async function handlePongMessage(msg) {
    numberOfPages += 1;
}

/**
 * Handles DOM data received from content scripts
 * @async
 * @param {Object} msg - Message containing DOM data
 */
async function handleDomDataMessage(msg) {
    if (!doc) {
        let data = await executeScriptToPromise(getDomData, br);
        if (data) {
            data = data[0].result;
            doc = new Document(data.html, data.text, data.url, data.doctype, data.title, data.attributes);
        } else {
            // TODO display error to user
            console.log("failed to get dom information");
        }
    }
    if (msg.data.url == getSiteUrl()) {
        numberOfPages -= 1;
    } else {
        let d = new Document(msg.data.html, msg.data.text, msg.data.url, msg.data.doctype, msg.data.doctype, msg.data.attributes);
        iframes.push(d);
    }
    if (doc && iframes.length >= numberOfPages) {
        doc.iframes = iframes;
        saveBookmark();
        return;
    }
}

/**
 * Opens a debug window with the given content (Chrome) or replaces current content (Firefox)
 * @param {string} content - HTML content to display for debugging
 */
function debugPopup(content) {
    if (is_chrome) {
        const win = window.open('', 'omnomDebug', 'menubar=yes,location=yes,resizable=yes,scrollbars=yes,status=yes');
        win.document.write(content);
    } else {
        document.getElementsByTagName('html')[0].innerHTML = content;
    }
}

/* ---------------------------------*
 * Diplay extension popup           *
 * ---------------------------------*/

/**
 * Initializes and displays the extension popup
 */
function displayPopup() {
    setTemplates();
    evaluateTemplates();
    setEventListeners();
    setOmnomSettings().then(fillFormFields, (err) => { optionsHandler(); renderError(err) });
    console.log('Omnom popup loaded!');
}

/**
 * Sets up event listeners for popup UI elements
 */
function setEventListeners() {
    const tagsInput = document.getElementById('tags');
    const chipContainer = document.getElementById('tag-chips');
    tagInput = new TagInputController(tagsInput, chipContainer);

    const closeButton = document.getElementById('close');
    closeButton?.addEventListener('click', closeHandler);

    const backButton = document.getElementById('back');
    backButton?.addEventListener('click', () => backHandler());

    const optionsButton = document.getElementById('omnom_options');
    optionsButton?.addEventListener('click', () => optionsHandler());

    const bookmarkForm = document.querySelector('form');
    bookmarkForm?.addEventListener('submit', createBookmark);
}

/**
 * Fills form fields with data from the current tab and extension settings
 * @async
 */
async function fillFormFields() {
    document.querySelector('form').action = `${getOmnomUrl()}add_bookmark`;
    document.getElementById('token').value = getOmnomToken();
    document.getElementById('public').checked = isOmnomDefaultPublic();

    // fill url input field
    document.getElementById('url').value = getSiteUrl();

    // fill title input field
    const title = await executeScriptToPromise(() => document.title);
    if (title && title[0]) {
        document.getElementById('title').value = title[0].result;
    }

    // fill notes input field
    const selection = await executeScriptToPromise(() => window.getSelection().toString());
    if (selection && selection[0]) {
        document.getElementById('notes').value = selection[0].result;
    }

    await fetchPageInfo();

    //fill tags
    tagInput.renderTags();

}

/**
 * Fetches page information (tags, collections) from the Omnom server
 * @async
 */
async function fetchPageInfo() {
    let pageTextData = await executeScriptToPromise(extractVisibleTextBlocks);
    let pageText = '';
    if(pageTextData && pageTextData[0]) {
        pageText = pageTextData[0].result.join(' ');
    }
    const requestOptions = {
        method: 'POST',
        headers: {'Content-Type': 'application/x-www-form-urlencoded'},
        body: `token=`+getOmnomToken()+'&text='+encodeURIComponent(pageText),
    };
    fetch(`${getOmnomUrl()}page_info`, requestOptions).then(async r => {
        let pageInfo = await r.json();
        if(Array.isArray(pageInfo.collections) && pageInfo.collections.length > 0){
            let el = document.querySelector("#collectionfield");
            el.classList.remove("hidden");
            for(let col of pageInfo.collections) {
                let o = document.createElement('option');
                o.innerText = col.name;
                o.setAttribute('value', col.id);
                el.querySelector("select").appendChild(o);
            }
        }
        if(Array.isArray(pageInfo.tags) && pageInfo.tags.length > 0) {
            tagInput.addTags(pageInfo.tags);
        }
    }).catch(error => console.log("error while fetching page info:", error));
}

/* ---------------------------------*
 * Event handlers                   *
 * ---------------------------------*/

/**
 * Handles back button click to return to main form
 */
function backHandler() {
    chrome.storage.local.get(['omnom_url', 'omnom_token'], function (data) {
        validateOptions(data.omnom_url, data.omnom_token).then(response => {
            if(response.ok) {
                clearContentContainer();
                updateBoundVar([{ 'onoptions': false }, { 'onafterdownload': false }, { 'onmain': true }]);
                fillFormFields();
                return
            }
            renderError('Invalid settings! Check out the <a href="https://github.com/asciimoo/omnom/wiki/Browser-extension#setup" target="_blank">documentation</a> for details.');
        });
    });
}

/**
 * Handles options button click to display settings
 * @async
 */
async function optionsHandler() {
    updateBoundVar([{ 'onoptions': true }, { 'onmain': false }]);

    const optionsPageText = await fetch('./options.html').then(stream => stream.text());
    const p = new DOMParser();
    const optionsPageElement = p.parseFromString(optionsPageText, 'text/html');
    const template = optionsPageElement.querySelector('template')?.content.cloneNode(true);
    const script = optionsPageElement.querySelector('script');
    contentContainer.appendChild(template);
    contentContainer.appendChild(copyScript(script));
}

/**
 * Handles close button click to close the popup
 */
function closeHandler() {
    window.close();
}

/* ---------------------------------*
 * Save bookmarks                   *
 * ---------------------------------*/

/**
 * Creates a bookmark from form data and initiates snapshot process
 * @async
 * @param {Event} e - Form submit event
 */
async function createBookmark(e) {
    e.preventDefault();
    form = new FormData(document.forms['add']);
    form.set('tags', tagInput.getTags().join(','));
    updateBoundVar([{ 'onafterdownload': true }, { 'onmain': false }]);
    renderProgressBar(document.getElementById('omnom_status'));
    if (numberOfPages < 2) {
        if (!numberOfPages) {
            console.log("content js isn't running, falling back to the naive snapshotting, which does not include iframes");
        }
        let data = await executeScriptToPromise(getDomData, br);
        if (!data) {
            // TODO display error to user
            console.log("failed to get dom information");
            saveBookmark();
            return;
        }
        data = data[0].result;
        doc = new Document(data.html, data.text, data.url, data.doctype, data.title, data.attributes);
        saveBookmark();
        return;
    } else {
        commChan.postMessage({type: "getDom"});
        if(chrome.runtime.lastError) {
            console.log(`Failed to deliver getDom message`, chrome.runtime.lastError);
        }
    }
}

/**
 * Saves the bookmark and snapshot to the Omnom server
 * @async
 */
async function saveBookmark() {
    console.time('createSnapshot');
    const snapshotData = await createSnapshot(doc);
    console.timeEnd('createSnapshot');
    const snapshotBlob = new Blob([snapshotData['dom']], { type: 'text/html' });
    form.append('snapshot', snapshotBlob);
    form.append('snapshot_text', doc.text);
    form.append('snapshot_title', doc.title);
    form.set('favicon', snapshotData['favicon']);
    const requestBody = {
        method: 'POST',
        body: form,
        // headers: {
        //     'Content-type': 'application/json; charset=UTF-8'
        // }
    };
    await fetch(`${getOmnomUrl()}add_bookmark`, requestBody)
        .then((resp) => checkStatus(resp)).then(async (resp) => {
            destroyProgressBar();
            const msg = await resp.json();
            let blobs = new Array();
            let blobMetas = new Array();
            let blobsSize = 0;
            for (const resource of doc.resources.getAll()) {
                const resourceBlob = new Blob([resource.content], { type: resource.mimetype });
                if (blobsSize && blobsSize + resourceBlob.size > blobSizeLimit) {
                    let rform = new FormData();
                    rform.append('token', getOmnomToken());
                    rform.append('sid', msg.snapshot_key);
                    rform.append('meta', JSON.stringify(blobMetas));
                    for (let i in blobs) {
                        rform.append('resource'+i, blobs[i]);
                    }
                    await fetch(`${getOmnomUrl()}add_resource`, {
                        method: 'POST',
                        body: rform
                    });
                    blobs = new Array();
                    blobMetas = new Array();
                    blobsSize = 0;
                }
                blobsSize += resourceBlob.size;
                blobs.push(resourceBlob);
                blobMetas.push({
                    'filename': resource.filename,
                    'mimetype': resource.mimetype,
                    'extension': resource.extension,
                });
            }
            if (blobs) {
                let rform = new FormData();
                rform.append('token', getOmnomToken());
                rform.append('sid', msg.snapshot_key);
                rform.append('meta', JSON.stringify(blobMetas));
                for (let i in blobs) {
                    rform.append('resource'+i, blobs[i]);
                }
                await fetch(`${getOmnomUrl()}add_resource`, {
                    method: 'POST',
                    body: rform
                });
            }
            renderSuccess('Snapshot successfully saved!', msg);
        }, async function(resp) {
            resp.json().then(j => {
                destroyProgressBar();
                renderError(`Failed to save bookmark! ${j.error}`);
            }).catch(error => {
                destroyProgressBar();
                renderError(`Failed to save bookmark!`);
            });
        });
}

/* ------------------------------------*
 * Template management                 *
 * ------------------------------------*/

/**
 * Extracts and stores templates from the DOM
 */
function setTemplates() {
    const templateElements = document.querySelectorAll('template');
    contentContainer = document.getElementById('omnom-content');
    [...templateElements].forEach(template => templates.set(template.id, template));
}

/**
 * Evaluates templates based on bound variables and renders/hides them accordingly
 */
function evaluateTemplates() {
    [...templates.values()].forEach(template => {
        const templateData = Object.keys(template.dataset);
        const parent = template.parentNode || contentContainer;
        if (templateData.length) {
            const shouldShow = templateData.some(attribute => {
                const attributeValue = (template.dataset[attribute] === 'true');
                return boundVars.hasOwnProperty(attribute) && boundVars[attribute] === attributeValue;
            });
            const templateNode = parent ?
                [...parent.children]?.find(child => child.id === template.content.children[0].id) :
                null;
            if (templateNode)
                if (!shouldShow && templateNode) {
                    parent.removeChild(templateNode);
                }
            if (!templateNode && shouldShow) {
                parent.appendChild(template.content.cloneNode(true));
            }
        }
    });
}

/**
 * Updates bound variables and re-evaluates templates
 * @param {Array<Object>} keys - Array of key-value pairs to update
 */
function updateBoundVar(keys) {
    let changed = null;
    keys.forEach(key => {
        changed = Object.keys(key)[0];
        if (boundVars.hasOwnProperty(changed)) {
            boundVars[changed] = key[changed];
        }
    });
    evaluateTemplates();
    setEventListeners();
}

/**
 * Clears all non-template content from the content container
 */
function clearContentContainer() {
    const templates = contentContainer.querySelectorAll('template');
    [...contentContainer.children].forEach(child => {
        if (![...templates].includes(child)) {
            contentContainer.removeChild(child);
        }
    });
}

export {
    displayPopup,
    saveBookmark
}

setupComms({type: "ping"});
