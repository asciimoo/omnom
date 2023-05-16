// SPDX-FileCopyrightText: 2021-2022 Adam Tauber, <asciimoo@gmail.com> et al.
//
// SPDX-License-Identifier: AGPL-3.0-only

async function createSnapshot(doc) {
    await doc.transformDom();
    return {
        'dom': doc.getDomAsText(),
        'text': doc.dom.getElementsByTagName("body")[0].innerText,
        'favicon': doc.favicon
    };
}

export { createSnapshot }
