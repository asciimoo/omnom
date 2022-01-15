import { renderProgressBar, destroyProgressBar } from './file-download';
import {
    createSnapshot
} from './snapshot';
import { TagInputController } from './tag-input';
import {
    renderError,
    renderSuccess,
    executeScriptToPromise,
    getSiteUrl,
    isDebug,
    setOmnomSettings,
    getOmnomUrl,
    getOmnomToken,
    checkStatus,
    copyScript
} from './utils';

const is_ff = typeof InstallTrigger !== 'undefined';
const is_chrome = !is_ff;

let tagInput = null;
let templates = new Map();
let boundVars = { onoptions: false, onafterdownload: false, onmain: true };
let contentContainer = null;
let form = null;

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

function displayPopup() {
    setTemplates();
    evaluateTemplates();
    setEventListeners();
    setOmnomSettings().then(fillFormFields, (err) => { optionsHandler(); renderError(err) });
    console.log('Loaded!');
}

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

async function fillFormFields() {
    document.querySelector('form').action = `${getOmnomUrl()}add_bookmark`;
    document.getElementById('token').value = getOmnomToken();

    // fill url input field
    document.getElementById('url').value = getSiteUrl();

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
    tagInput.renderTags();

}

/* ---------------------------------*
    * Event handlers                   *
    * ---------------------------------*/

function backHandler() {
    clearContentContainer();
    updateBoundVar([{ 'onoptions': false }, { 'onafterdownload': false }, { 'onmain': true }]);
    fillFormFields();
}

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

function closeHandler() {
    window.close();
}

/* ---------------------------------*
    * Save bookmarks                   *
    * ---------------------------------*/

async function createBookmark(e) {
    e.preventDefault();
    form = new FormData(document.forms['add']);
    updateBoundVar([{ 'onafterdownload': true }, { 'onmain': false }]);
    renderProgressBar(document.getElementById('omnom_status'));
    createSnapshot();
}
async function saveBookmark(snapshotData) {
    form.set('tags', tagInput.getTags().join(','));

    //console.time('createSnapshot');
    //console.timeEnd('createSnapshot');
    if (isDebug()) {
        debugPopup(snapshotData['dom']);
        return;
    }
    const snapshotBlob = new Blob([snapshotData['dom']], { type: 'text/html' });
    form.append('snapshot', snapshotBlob);
    form.append('snapshot_text', snapshotData['text']);
    form.set('favicon', snapshotData['favicon']);
    console.log(form);
    const requestBody = {
        method: 'POST',
        body: form,
        // headers: {
        //     'Content-type': 'application/json; charset=UTF-8'
        // }
    }
    await fetch(`${getOmnomUrl()}add_bookmark`, requestBody)
        .then((resp) => checkStatus(resp)).then(() => {
            destroyProgressBar();
            renderSuccess('Snapshot successfully saved!');
        }, (err) => {
            destroyProgressBar();
            renderError(`Failed to save bookmark! Error: ${err.statusText}`, err);
        });
}

/* ---------------------------------*
    * Template managemen               *
    * ---------------------------------*/

function setTemplates() {
    const templateElements = document.querySelectorAll('template');
    contentContainer = document.getElementById('omnom-content');
    [...templateElements].forEach(template => templates.set(template.id, template));
}

function evaluateTemplates() {
    [...templates.values()].forEach(template => {
        const templateData = Object.keys(template.dataset);
        const parent = template.parentNode || contentContainer;
        if (templateData.length) {
            const shouldShow = templateData.some(attribute => {
                const attributeValue = (template.dataset[attribute] === 'true');
                return boundVars.hasOwnProperty(attribute) && boundVars[attribute] === attributeValue
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
    })
}

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

function clearContentContainer() {
    const templates = contentContainer.querySelectorAll('template');
    [...contentContainer.children].forEach(child => {
        if (![...templates].includes(child)) {
            contentContainer.removeChild(child);
        }
    });
}


function processMsg(msg) {
    console.log(msg);
}


export {
    displayPopup,
    saveBookmark
}
