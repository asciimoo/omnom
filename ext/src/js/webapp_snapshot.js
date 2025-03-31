import { Document } from "./modules/document";
import { createSnapshot } from './modules/snapshot';
import { getDomData } from "./modules/get-dom-data";
import { setSiteUrl } from "./modules/utils";


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
