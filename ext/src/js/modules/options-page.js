export default function () {
    function saveOptions(e) {
        chrome.storage.local.set({
            omnom_url: document.querySelector('#url').value,
            omnom_token: document.querySelector('#token').value,
            omnom_debug: document.querySelector('#debug').checked,
        });
        e.preventDefault();
        window.close();
    }

    function loadContent() {
        const template = document.getElementById('options-body');
        const container = document.createElement('div');
        container.className = 'omnom-popup__container options-page__container';

        const optionsContent = [...template.content.children];
        optionsContent.forEach(child => container.appendChild(child));
        document.body.appendChild(container);
        restoreOptions();
        document.querySelector('form').addEventListener('submit', saveOptions);
    }

    function restoreOptions() {
        const optionsElement = document.getElementById('omnom-options');
        if (optionsElement) {
            chrome.storage.local.get(['omnom_url', 'omnom_token', 'omnom_debug'], function (data) {
                document.querySelector('#url').value = data.omnom_url || '';
                document.querySelector('#token').value = data.omnom_token || '';
                document.querySelector('#debug').checked = data.omnom_debug;
                isFormValid();
            });
            document.querySelector('form').addEventListener('submit', saveOptions);
            document.querySelector('#url').addEventListener('input', isFormValid);
        } else {
            document.addEventListener('DOMContentLoaded', () => {
                loadContent();
            })
        }
    }

    function isFormValid() {
        const serverUrl = document.querySelector('#url');
        const backButton = document.querySelector('#back');
        console.log(serverUrl.value);
        if (!serverUrl || !backButton) return;
        backButton.disabled = !!serverUrl.value ? false : true;
    }

    (function onLoad() {
        restoreOptions();
    })()
}
