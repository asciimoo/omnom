import { Subject } from 'rxjs';
import { arrayBufferToBase64, checkStatus, fullURL, isDebug } from './utils';

const downloadStatus = {
    DOWNLOADING: 'downloading',
    DOWNLOADED: 'downloaded',
    FAILED: 'failed'
};
const initialBar = `<h3>Downloading resources <span id="progress-counter">(0/0)</span></h3>
<h3>Failed requests: <span id="failed-counter">0</span></h3>`;

let downloadedCount = 0;
let downloadCount = 0;
let failedCount = 0;
let downloadState = new Subject();
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
    if (contentType.toLowerCase().search('text') != -1) {
        // TODO use charset of the response        
        return await responseObj.text();
    }
    const buff = await responseObj.arrayBuffer();
    const base64Flag = `data:${contentType};base64,`;
    return `${base64Flag}${arrayBufferToBase64(buff)}`
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
    componentTarget = target;
    fragment = createFragment(initialBar, componentTarget);
    componentTarget.appendChild(fragment);
    downloadState.subscribe(state => {
        const progressCounter = componentTarget.querySelector('#progress-counter');
        progressCounter.innerText = `(${state.downloadCount}/${state.downloadedCount})`;
        const failedCounter = componentTarget.querySelector('#failed-counter');
        failedCounter.innerText = ` ${state.failedCount}`;
    });
}

function createFragment(componentString) {
    if (componentTarget) {
        const range = document.createRange();
        range.selectNode(componentTarget);
        return range.createContextualFragment(componentString);
    }
    return
}

export { downloadState, downloadFile, renderProgressBar };
