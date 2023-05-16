// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

export default function () {
    function saveOptions(e) {
        let serverUrl = document.querySelector('#url').value;
        if (!serverUrl.endsWith('/')) {
            serverUrl += '/';
        }
        chrome.storage.local.set({
            omnom_url: serverUrl,
            omnom_token: document.querySelector('#token').value,
            omnom_debug: document.querySelector('#debug').checked,
            omnom_public: document.querySelector('#public').checked,
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
            chrome.storage.local.get(['omnom_url', 'omnom_token', 'omnom_debug', 'omnom_public'], function (data) {
                document.querySelector('#url').value = data.omnom_url || '';
                document.querySelector('#token').value = data.omnom_token || '';
                document.querySelector('#debug').checked = data.omnom_debug;
                document.querySelector('#public').checked = data.omnom_public;
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
        if (!serverUrl || !backButton) return;
        backButton.disabled = !!serverUrl.value ? false : true;
    }

    (function onLoad() {
        restoreOptions();
    })()
}
