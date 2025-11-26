// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Resource storage and management for downloaded assets.
 * Handles fetching, storing, and hashing of page resources.
 */

import { sha256 } from './utils';
import { fetchURL } from './file-download';

/**
 * Map of file extension aliases
 * @type {Map<string, string>}
 */
const extMap = new Map([
    ["jpeg", "jpg"],
]);

/**
 * Represents a downloadable resource (image, stylesheet, font, etc.)
 * @class
 */
class Resource {
    /**
     * Creates a Resource instance
     * @param {ArrayBuffer} content - The resource content
     * @param {string} mimetype - The resource MIME type
     * @param {string} filename - The resource filename
     */
    constructor(content, mimetype, filename) {
        this.content = content;
        this.mimetype = mimetype;
        this.filename = filename;
        this.extension = 'unknown';
        if (mimetype) {
            this.extension = mimetype.split(" ")[0].split("/").pop().toLowerCase().split("+")[0].split(";")[0];
        }
        if (extMap.has(this.extension)) {
            this.extension = extMap.get(this.extension);
        }
        this.src = '';
    }

    /**
     * Computes SHA-256 hash of the resource content and sets the src path
     * @async
     */
    async sha() {
        this.sha256sum = await sha256(this.content);
        this.src = `../../resources/${this.sha256sum[0]}${this.sha256sum[1]}/${this.sha256sum}.${this.extension}`;
    }

    /**
     * Updates the resource content and recomputes hash
     * @async
     * @param {ArrayBuffer|string} newContent - The new content
     */
    async updateContent(newContent) {
        this.content = newContent;
        await this.sha();
    }
}

/**
 * Storage for managing downloaded resources
 * @class
 */
class ResourceStorage {
    /**
     * Creates a ResourceStorage instance
     */
    constructor() {
        this.resources = new Map([]);
    }

    /**
     * Creates and stores a resource from a URL
     * @async
     * @param {string} url - The URL to fetch
     * @returns {Promise<Resource|undefined>} The created resource or undefined on error
     */
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

    /**
     * Gets all stored resources
     * @returns {Iterator<Resource>} Iterator over all resources
     */
    getAll() {
        return this.resources.values();
    }
}

export {
    Resource,
    ResourceStorage
}
