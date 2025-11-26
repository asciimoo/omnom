// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview File download utilities for fetching and tracking resource downloads.
 * Provides progress tracking and status updates via RxJS subjects.
 */

import { Subject } from 'rxjs';
import { arrayBufferToBase64, checkStatus, fullURL } from './utils';

/**
 * Download status enumeration
 * @enum {string}
 */
const downloadStatus = {
    DOWNLOADING: 'downloading',
    DOWNLOADED: 'downloaded',
    FAILED: 'failed'
};
const initialBar = `<div class="file-download">
<h3>Snapshotting in progress...</h3>
<progress class="progress is-small file-download__progress" value="15" max="100">0%</progress>
<span id="progress-counter">0 of 0 resources snapshotted</span>
<h3>Failed requests: <span id="failed-counter">0</span></h3>
</div>`;

/**
 * Counter for successfully downloaded files
 * @type {number}
 */
let downloadedCount = 0;

/**
 * Total number of files to download
 * @type {number}
 */
let downloadCount = 0;

/**
 * Counter for failed downloads
 * @type {number}
 */
let failedCount = 0;

/**
 * RxJS Subject for download state updates
 * @type {Subject|null}
 */
let downloadState = null;

/**
 * DOM element where progress bar is rendered
 * @type {HTMLElement|null}
 */
let componentTarget = null;

/**
 * Document fragment for progress bar
 * @type {DocumentFragment|null}
 */
let fragment = null;

/**
 * Downloads a file from the given URL and converts it to base64 data URI
 * @async
 * @param {string} url - The URL of the file to download
 * @returns {Promise<string>} Base64-encoded data URI or empty string on error
 */
async function downloadFile(url) {
    if (!url || (url || '').startsWith('data:')) {
        return url;
    }
    url = fullURL(url);
    console.log('fetching ', url);
    const options = {
        method: 'GET',
        cache: 'default',
    };
    const request = new Request(url, options);
    updateStatus(downloadStatus.DOWNLOADING);
    let hasError = false;
    const responseObj = await fetch(request, options)
        .then(checkStatus).catch(() => {
            updateStatus(downloadStatus.FAILED);
            hasError = true;
        });
    if (hasError) {
        return '';
    }
    const contentType = responseObj.headers.get('Content-Type');
    updateStatus(downloadStatus.DOWNLOADED);
    if (contentType && contentType.toLowerCase().search('text') != -1) {
        // TODO use charset of the response        
        return await responseObj.text();
    }
    const buff = await responseObj.arrayBuffer();
    const base64Flag = `data:${contentType};base64,`;
    return `${base64Flag}${arrayBufferToBase64(buff)}`
}

/**
 * Fetches a URL and returns the response object
 * @async
 * @param {string} url - The URL to fetch
 * @returns {Promise<Response|string>} Response object or empty string on error
 */
async function fetchURL(url) {
    const options = {
        method: 'GET',
        cache: 'default',
    };
    updateStatus(downloadStatus.DOWNLOADING);
    const request = new Request(url, options);
    let hasError = false;
    const responseObj = await fetch(request, options).then(checkStatus).catch(() => {
        hasError = true;
        updateStatus(downloadStatus.FAILED);
    });
    if (!hasError) {
        updateStatus(downloadStatus.DOWNLOADED);
        return responseObj;
    }
    return '';
}

/**
 * Updates the download status counters and notifies subscribers
 * @param {string} status - The download status (DOWNLOADING, DOWNLOADED, or FAILED)
 */
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
    if(downloadState !== null) {
        downloadState.next({ downloadCount, downloadedCount, failedCount });
    }
}

/**
 * Renders the progress bar UI in the specified target element
 * @param {HTMLElement} target - The DOM element to render the progress bar into
 */
function renderProgressBar(target) {
    downloadState = new Subject();
    componentTarget = target;
    fragment = createFragment(initialBar, componentTarget);
    componentTarget.appendChild(fragment);
    downloadState.subscribe(state => {
        const failedCounter = componentTarget.querySelector('#failed-counter');
        failedCounter.innerText = ` ${state.failedCount}`;
        const progressCounter = componentTarget.querySelector('#progress-counter');
        progressCounter.innerText = `${state.downloadedCount} of ${state.downloadCount} resources snapshotted`;
        const progressBar = componentTarget.querySelector('progress');
        const percentage = ((state.downloadedCount + state.failedCount) / state.downloadCount) * 100;
        progressBar.value = percentage;
    });
}

/**
 * Removes the progress bar from the DOM and resets counters
 */
function destroyProgressBar() {
    downloadState.unsubscribe();
    const el = componentTarget.querySelector('.file-download');
    componentTarget.removeChild(el);
    downloadCount = 0;
    downloadedCount = 0;
    failedCount = 0;
}

/**
 * Creates a document fragment from an HTML string
 * @param {string} componentString - HTML string to convert to fragment
 * @returns {DocumentFragment|undefined} Document fragment or undefined
 */
function createFragment(componentString) {
    if (componentTarget) {
        const range = document.createRange();
        range.selectNode(componentTarget);
        return range.createContextualFragment(componentString);
    }
    return
}

export {
    downloadState,
    downloadFile,
    fetchURL,
    renderProgressBar,
    destroyProgressBar
};
