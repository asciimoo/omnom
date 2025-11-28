/**
 * @fileoverview Web application snapshot functionality.
 * Creates snapshots of web pages for the Omnom extension.
 */

import { Document } from "./modules/document";
import { createSnapshot } from './modules/snapshot';
import { getDomData } from "./modules/get-dom-data";
import { setSiteUrl } from "./modules/utils";

/**
 * Creates a complete snapshot of the current web page including DOM, resources, and metadata
 * @async
 * @returns {Promise<Object>} Snapshot object containing DOM, favicon, resources, text, and title
 * @property {string} dom - The processed DOM content
 * @property {string} favicon - The page favicon
 * @property {Array<Object>} resources - Array of page resources (images, stylesheets, etc.)
 * @property {string} text - The extracted text content
 * @property {string} title - The page title
 */
async function createOmnomSnapshot() {
    const data = getDomData();
    const doc = new Document(data.html, data.text, data.url, data.doctype, data.title, data.attributes);
    await setSiteUrl(data.url);
    const s = await createSnapshot(doc);
    const ret = {
        'dom': s.dom,
        'favicon': s.favicon,
        'resources':  [],
        'text': data.text,
        'title': data.title,
        'multimedia_count': doc.multimediaCount,
    };
    for(let res of doc.resources.getAll()) {
        res.content = Array.from(new Uint8Array(res.content));
        ret.resources.push(res);
    }
    return ret;
}

export {
    createOmnomSnapshot
};
