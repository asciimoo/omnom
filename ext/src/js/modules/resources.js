import { sha256 } from './utils';
import { fetchURL } from './file-download';

const extMap = new Map([
    ["jpeg", "jpg"],
]);

class Resource {
    constructor(content, mimetype, filename) {
        this.content = content;
        this.mimetype = mimetype;
        this.filename = filename;
        this.filetype = mimetype.split(" ")[0].split("/").pop().toLowerCase().split("+")[0];
        if (extMap.has(this.filetype)) {
            this.filetype = extMap.get(this.filetype);
        }
        this.src = '';
    }

    async sha() {
        this.sha256sum = await sha256(this.content);
        this.src = `../../resources/${this.sha256sum[0]}${this.sha256sum[1]}/${this.sha256sum}.${this.filetype}`;
    }

    async updateContent(newContent) {
        this.content = newContent;
        await this.sha();
    }
}

class ResourceStorage {
    constructor() {
        this.resources = new Map([]);
    }

    async create(url) {
        if (this.resources.has(url)) {
            return this.resources.get(url);
        }
        let resp = await fetchURL(url);
        if (!resp) {
            return;
        }
        const content = await resp.arrayBuffer();
        if (!content) {
            return;
        }
        const contentType = resp.headers.get('Content-Type');
        const parsedURL = new URL(url);
        const fname = parsedURL.pathname.split('/').pop();
        let res = new Resource(content, contentType, fname);
        await res.sha();
        this.resources.set(url, res);
        return res;
    }

    getAll() {
        return this.resources.values();
    }
}

let resources = new ResourceStorage();


export {
    Resource,
    resources
}
