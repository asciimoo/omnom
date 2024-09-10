// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

function addTag(value, tagChipContainer, tags) {
    renderTag(value, tagChipContainer, tags);
    tags.push(value);
}

function checkAddTrigger(event, chipContainer, tags, inputElement) {
    if (event.target.value.includes(',')) {
        addTag(event.target.value.slice(0, -1), chipContainer, tags); inputElement.value = '';
    }
}

function renderTag(value, tagChipContainer, tags) {
    const tagTemplate = `<div class="control chip-control">
        <span class="tag is-rounded">
            ${value}
            <button type="button" class="delete is-small"></button>
        </span>
    </div>`
    const template = document.createElement('template');
    template.innerHTML = tagTemplate;
    const deleteButton = template.content.querySelector('button');
    deleteButton.addEventListener('click', deleteTag.bind({}, template.content.firstChild, tagChipContainer, tags));
    tagChipContainer.appendChild(template.content.firstChild);
}

function deleteTag(chipElement, tagChipContainer, tags) {
    tagChipContainer.removeChild(chipElement);
    tags = [...tagChipContainer.children].map(child => child.innerText);
}

function renderTags(tags, tagChipContainer) {
    const fragment = document.createDocumentFragment();
    tags.forEach(tag => {
        renderTag(tag, fragment);
    });
    tagChipContainer.appendChild(fragment);
}

function TagInputController(inputElement, chipContainer) {
    let tags = [];
    this.getTags = () => (tags);
    this.renderTags = () => renderTags(tags, chipContainer);
    inputElement?.addEventListener('input', (event) => { checkAddTrigger(event, chipContainer, tags, inputElement) })
    inputElement?.addEventListener('change', (event) => { addTag(event.target.value, chipContainer, tags); inputElement.value = '' });
}

export { TagInputController };