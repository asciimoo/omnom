// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

/**
 * @fileoverview Tag input controller for managing bookmark tags.
 * Provides UI for adding, removing, and rendering tags as chips.
 */

/**
 * Adds a tag to the tags array and renders it
 * @param {string} value - The tag text to add
 * @param {HTMLElement} tagChipContainer - Container for tag chips
 * @param {Array<string>} tags - Array of tag strings
 */
function addTag(value, tagChipContainer, tags) {
    value = value.trim();
    if(tags.includes(value) || !value) {
        return;
    }
    renderTag(value, tagChipContainer, tags);
    tags.push(value);
}

/**
 * Checks if a key press should trigger tag addition
 * @param {KeyboardEvent} event - The keyboard event
 * @param {HTMLElement} chipContainer - Container for tag chips
 * @param {Array<string>} tags - Array of tag strings
 * @param {HTMLInputElement} inputElement - The tag input element
 * @returns {boolean} Always returns false
 */
function checkAddTrigger(event, chipContainer, tags, inputElement) {
    console.log(event);
    if (event.key == 'Enter' || event.keyCode == 13 || event.key == ',') {
        addTag(event.target.value, chipContainer, tags);
        inputElement.value = '';
        event.preventDefault();
    }
    return false;
}

/**
 * Renders a single tag chip
 * @param {string} value - The tag text
 * @param {HTMLElement} tagChipContainer - Container for tag chips
 * @param {Array<string>} tags - Array of tag strings
 */
function renderTag(value, tagChipContainer, tags) {
    const tagTemplate = `<div class="control chip-control" aria-label="tag ${value}">
        <span class="tag is-rounded">
            ${value}
            <button type="button" class="delete is-small" aria-label="Delete tag ${value}"><span class="icon"><i class="fas fa-times"></i></span></button>
        </span>
    </div>`
    const template = document.createElement('template');
    template.innerHTML = tagTemplate;
    const deleteButton = template.content.querySelector('button');
    deleteButton.addEventListener('click', deleteTag.bind({}, template.content.firstChild, tagChipContainer, tags));
    tagChipContainer.appendChild(template.content.firstChild);
}

/**
 * Deletes a tag chip
 * @param {HTMLElement} chipElement - The chip element to delete
 * @param {HTMLElement} tagChipContainer - Container for tag chips
 * @param {Array<string>} tags - Array of tag strings
 */
function deleteTag(chipElement, tagChipContainer, tags) {
    tagChipContainer.removeChild(chipElement);
    tags = [...tagChipContainer.children].map(child => child.innerText);
}

/**
 * Renders all tags as chips
 * @param {Array<string>} tags - Array of tag strings
 * @param {HTMLElement} tagChipContainer - Container for tag chips
 */
function renderTags(tags, tagChipContainer) {
    tagChipContainer.innerHTML = '';
    const fragment = document.createDocumentFragment();
    tags.forEach(tag => {
        renderTag(tag, fragment);
    });
    tagChipContainer.appendChild(fragment);
}

/**
 * Controller for managing tag input and display
 * @class
 * @param {HTMLInputElement} inputElement - The tag input field
 * @param {HTMLElement} chipContainer - Container for displaying tag chips
 */
function TagInputController(inputElement, chipContainer) {
    let tags = [];
    this.getTags = () => (tags);
    this.renderTags = () => renderTags(tags, chipContainer);
    this.addTags = (new_tags) => {
        for(let t of new_tags) {
            addTag(t.text ? t.text : t, chipContainer, tags);
        }
    };
    inputElement?.addEventListener('keydown', (event) => { checkAddTrigger(event, chipContainer, tags, inputElement) })
    //inputElement?.addEventListener('input', (event) => { checkAddTrigger(event, chipContainer, tags, inputElement) })
    inputElement?.addEventListener('change', (event) => { addTag(event.target.value, chipContainer, tags); inputElement.value = '' });
}

export { TagInputController };
