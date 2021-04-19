function saveOptions(e) {
  chrome.storage.local.set({
    omnom_url: document.querySelector("#url").value,
    omnom_token: document.querySelector("#token").value
  });
  e.preventDefault();
  window.close();
}

function restoreOptions() {
  chrome.storage.local.get(['omnom_url', 'omnom_token'], function(data) {
    document.querySelector("#url").value = data.omnom_url || '';
    document.querySelector("#token").value = data.omnom_token || '';
  });
}

document.addEventListener('DOMContentLoaded', restoreOptions);
document.querySelector("form").addEventListener("submit", saveOptions);
