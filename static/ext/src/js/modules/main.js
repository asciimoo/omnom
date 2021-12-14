import {
    createSnapshot
} from './snapshot';
import {
    renderError,
    renderSuccess,
    executeScriptToPromise,
    getSiteUrl,
    isDebug,
    setOmnomSettings,
    getOmnomUrl,
    getOmnomToken,
    checkStatus
} from './utils';

export default function () {
    const is_ff = typeof InstallTrigger !== 'undefined';
    const is_chrome = !is_ff;

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
                    const nodeToRemove = template.parentNode ?
                        [...template.parentNode.children]?.find(child => child.id === template.content.children[0].id) :
                        null;
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
        if (isDebug()) {
            debugPopup(snapshotData['dom']);
            return;
        }
        const form = new FormData(document.forms['add']);
        form.append('snapshot', snapshotData['dom']);
        form.append('snapshot_text', snapshotData['text']);
        form.set('tags', tags.join(','));
        console.log(form);
        const requestBody = {
            method: 'POST',
            body: form,
            // headers: {
            //     'Content-type': 'application/json; charset=UTF-8'
            // }
        }
        await fetch(`${getOmnomUrl()}add_bookmark`, requestBody).then(
            checkStatus).then(renderSuccess('Bookmark saved!'), renderError(`Failed to save bookmark:${err}, ${requestBody}`));
    }

    /* ---------------------------------*
     * Create Snapshot                  *
     * ---------------------------------*/


    /* ---------------------------------*
     * Sanitize css                     *
     * ---------------------------------*/

    /* ---------------------------------*
     * Utility functions                *
     * ---------------------------------*/

    document.addEventListener('DOMContentLoaded', displayPopup);

    /* ---------------------------------*
     * End of omnom code                *
     * ---------------------------------*/
}