// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+
import {
    renderError,
    renderSuccess
} from './utils';

export default function () {
    function saveOptions(e) {
        const serverUrlErrMsg = `Invalid server URL. Use <code>http[s]://youromnom.tld/</code> format.`;
        let serverUrl = document.querySelector('#url').value;
        if (!serverUrl.endsWith('/')) {
            serverUrl += '/';
        }
        let formData = new FormData();
        formData.append('token', document.querySelector('#token').value);
        fetch(serverUrl + 'check_token', {
            'method': 'POST',
            'body': formData,
        }).then(response => {
			if (response.ok) {
                persistSettings(serverUrl);
				return
			}
            if(response.status == 403) {
                response.json().then(j => {
                    renderError(`Invalid settings! ${j.message}`);
                }).catch(error => {
                    renderError(serverUrlErrMsg);
                });
            } else {
                renderError(serverUrlErrMsg);
            }
            return
		}).catch(error => {
			renderError(serverUrlErrMsg);
			return
		});
        e.preventDefault();
    }
    function persistSettings(serverUrl) {
        chrome.storage.local.set({
            omnom_url: serverUrl,
            omnom_token: document.querySelector('#token').value,
            omnom_public: document.querySelector('#public').checked,
        });
        renderSuccess('Settings successfully saved!');
        setTimeout(window.close, 2000);
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
            chrome.storage.local.get(['omnom_url', 'omnom_token', 'omnom_public'], function (data) {
                document.querySelector('#url').value = data.omnom_url || '';
                document.querySelector('#token').value = data.omnom_token || '';
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
