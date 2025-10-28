
const extIds = [
    "nhpakcgbfdhghjnilnbgofmaeecoojei", // chrome
    "{f0bca7ce-0cda-41dc-9ea8-126a50fed280}", // firefox
];

const hasUser = document.getElementById("has_user").getAttribute("value") == "1";
const baseUrl = new URL(document.getElementById("base_url").getAttribute("value"), window.location.href).href;

function sendMsg(msg, responseHandler) {
    for(let extId of extIds) {
        try {
            chrome.runtime.sendMessage(extId, msg, responseHandler);
        } catch {
        }
    }
}

window.onload = () => {
    sendMsg({"action": "ping", "url": baseUrl}, (resp) => {
        if(!resp || !resp.url) {
            return;
        }
        if(hasUser && resp.url != "same") {
            fetch(baseUrl + "profile?format=json", {"method": "POST"}).then((r) => {
                if(r.ok) {
                    return r.json();
                }
            }).then((j) => {
                sendMsg({"action": "set-settings", "token": j.AddonTokens[0].text, "url": baseUrl}, console.log);
            });
        }
        // TODO consider security ->
        //if(!hasUser && resp.url == "same") {
        //    console.log("show option to sign in");
        //}

    });
}

//function createModal(onSuccess) {
//    let temp = document.getElementById("tpl-modal");
//    let n = temp.content.cloneNode(true);
//    let modal = document.getElementById("modal");
//    modal.appendChild(n);
//    modal.classList.add('is-active');
//    modal.getElementsByClassName('tpl-no')[0].addEventListener('click', (e) => {
//        modal.classList.remove('is-active');
//    });
//    modal.getElementsByClassName('tpl-yes')[0].addEventListener('click', onSuccess);
//}
//function closeModal() {
//    let modal = document.getElementById("modal");
//    modal.classList.remove('is-active');
//    modal.innerHTML = '';
//}
