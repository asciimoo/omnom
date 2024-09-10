// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

async function createSnapshot(doc) {
    await doc.transformDom();
    return {
        'dom': doc.getDomAsText(),
        'text': doc.dom.getElementsByTagName("body")[0].innerText,
        'favicon': doc.favicon
    };
}

export { createSnapshot }
