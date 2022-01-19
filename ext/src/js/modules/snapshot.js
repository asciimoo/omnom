async function createSnapshot(doc) {
    await doc.transformDom();
    return {
        'dom': doc.getDomAsText(),
        'text': doc.dom.getElementsByTagName("body")[0].innerText,
        'favicon': doc.favicon
    };
}

export { createSnapshot }
