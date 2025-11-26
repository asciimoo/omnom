// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Snapshot creation functionality.
 * Transforms documents and creates snapshots for archiving.
 */

/**
 * Creates a snapshot from a Document instance
 * @async
 * @param {Document} doc - The document to snapshot
 * @returns {Promise<Object>} Snapshot object with DOM and favicon
 * @property {string} dom - The transformed DOM as text
 * @property {string} favicon - The page favicon
 */
async function createSnapshot(doc) {
    await doc.transformDom();
    return {
        'dom': doc.getDomAsText(),
        'favicon': doc.favicon,
    };
}

export { createSnapshot }
