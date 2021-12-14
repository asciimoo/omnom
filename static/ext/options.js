function saveOptions(e) {
    chrome.storage.local.set({
        omnom_url: document.querySelector('#url').value,
        omnom_token: document.querySelector('#token').value,
        omnom_debug: document.querySelector('#debug').checked,
    });
    e.preventDefault();
    window.close();
}

function restoreOptions() {
    console.log('restoring');
    chrome.storage.local.get(['omnom_url', 'omnom_token', 'omnom_debug'], function (data) {
        document.querySelector('#url').value = data.omnom_url || '';
        document.querySelector('#token').value = data.omnom_token || '';
        document.querySelector('#debug').checked = data.omnom_debug;
    });
}

document.addEventListener('DOMContentLoaded', restoreOptions);
document.querySelector('form').addEventListener('submit', saveOptions);
(function onLoad() {
    restoreOptions();
})()
