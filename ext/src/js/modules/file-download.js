// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

import { Subject } from 'rxjs';
import { arrayBufferToBase64, checkStatus, fullURL, isDebug } from './utils';

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

let downloadedCount = 0;
let downloadCount = 0;
let failedCount = 0;
let downloadState = null;
let componentTarget = null;
let fragment = null;

async function downloadFile(url) {
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

async function fetchURL(url) {
    const options = {
        method: 'GET',
        cache: isDebug ? 'no-cache' : 'default',
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
    downloadState.next({ downloadCount, downloadedCount, failedCount });
}

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

function destroyProgressBar() {
    downloadState.unsubscribe();
    const el = componentTarget.querySelector('.file-download');
    componentTarget.removeChild(el);
    downloadCount = 0;
    downloadedCount = 0;
    failedCount = 0;
}

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
