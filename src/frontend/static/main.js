class TagSearchManager {
    constructor(searchContainer) {
        if (!searchContainer) {
            throw new Error('Search container element is required');
        }
        this.searchContainer = searchContainer;
        const searchInput = searchContainer.querySelector('#tag-search');
        if (!searchInput) {
            throw new Error('Search input element with id "tag-search" not found');
        }
        this.searchInput = searchInput;
        this.queryDebug = document.getElementById('query-debug');
        if (!this.queryDebug) {
            console.warn('Query debug element not found');
        }
        this.tagContainer = document.createElement('div');
        this.tagContainer.className = 'tag-container';
        const tagSearch = this.searchContainer.querySelector('.tag-search');
        if (!tagSearch) {
            throw new Error('Tag search container element not found');
        }
        tagSearch.insertBefore(this.tagContainer, this.searchInput);
        this.autocompleteResults = document.getElementById('autocomplete-results');
        if (!this.autocompleteResults) {
            throw new Error('Autocomplete results element not found');
        }
        this.selectedTags = new Set();
        this.setupEventListeners();
    }

    setupEventListeners() {
        this.searchInput.addEventListener('input', this.handleInput.bind(this));
        this.searchInput.addEventListener('keydown', this.handleKeydown.bind(this));
        this.searchButton = this.searchContainer.querySelector('#search-button');
        this.searchButton.addEventListener('click', () => {
            this.triggerSearch();
        });
        this.autocompleteResults.addEventListener('click', this.handleAutocompleteClick.bind(this));
    }

    async handleInput(event) {
        const query = event.target.value.trim();
        if (query.length < 1) {
            this.hideAutocomplete();
            return;
        }

        try {
            const response = await fetch(`/api/tags/suggest?q=${encodeURIComponent(query)}`);
            const tags = await response.json();
            this.showAutocomplete(tags, query);
        } catch (error) {
            console.error('Error fetching tag suggestions:', error);
        }
    }

    handleAutocompleteClick(event) {
        const tagElement = event.target.closest('.tag-suggestion');
        if (tagElement) {
            this.addTag(tagElement.textContent);
            this.triggerSearch();
        }
    }

    handleKeydown(event) {
        if (event.key === 'Enter') {
            const value = this.searchInput.value.trim();
            if (value) {
                if (this.autocompleteResults.children.length > 0) {
                    this.addTag(value);
                    this.triggerSearch();
                } else {
                    this.triggerSearch();
                }
            } else {
                this.triggerSearch();
            }
            event.preventDefault();
        } else if (event.key === 'Backspace' && !this.searchInput.value) {
            this.removeLastTag();
            this.triggerSearch();
        }
    }

    showAutocomplete(tags, query) {
        if (!tags || !Array.isArray(tags)) {
            console.warn('Invalid tags data received');
            this.hideAutocomplete();
            return;
        }

        if (tags.length === 0) {
            this.hideAutocomplete();
            return;
        }

        this.autocompleteResults.innerHTML = tags
            .filter(tag => !this.selectedTags.has(tag))
            .map(tag => `<div class="tag-suggestion">${tag}</div>`)
            .join('');
        this.autocompleteResults.style.display = 'block';
    }

    hideAutocomplete() {
        this.autocompleteResults.style.display = 'none';
        this.autocompleteResults.innerHTML = '';
    }

    addTag(tag) {
        if (this.selectedTags.has(tag)) return;

        const tagElement = document.createElement('div');
        tagElement.className = 'selected-tag';
        const removeButton = document.createElement('span');
        removeButton.className = 'remove-tag';
        removeButton.textContent = '×';
        removeButton.addEventListener('click', () => {
            tagElement.remove();
            this.removeTag(tag);
        });

        tagElement.textContent = tag;
        tagElement.appendChild(removeButton);
        this.tagContainer.appendChild(tagElement);
        this.selectedTags.add(tag);
        this.searchInput.value = '';
        this.hideAutocomplete();
        this.triggerSearch();
    }

    removeTag(tag) {
        this.selectedTags.delete(tag);
        this.triggerSearch();
    }

    removeLastTag() {
        const tags = this.tagContainer.getElementsByClassName('selected-tag');
        if (tags.length > 0) {
            const lastTag = tags[tags.length - 1];
            const tagText = lastTag.textContent.slice(0, -1).trim();
            this.selectedTags.delete(tagText);
            lastTag.remove();
            this.triggerSearch();
        }
    }

    triggerSearch() {
        const tags = Array.from(this.selectedTags);
        const query = this.searchInput.value.trim();
        
        this.queryDebug.textContent = `Current query: ${query ? `text="${query}"` : ''} ${tags.length ? `tags=[${tags.join(', ')}]` : ''}`;

        const searchEvent = new CustomEvent('tagsearch', {
            detail: {
                tags: tags,
                query: query
            }
        });
        this.searchContainer.dispatchEvent(searchEvent);
    }

    getSearchParams() {
        return {
            tags: Array.from(this.selectedTags),
            query: this.searchInput.value.trim()
        };
    }
}

let updateImageGrid;
let handleTagClick;
let showImageModal;

document.addEventListener('DOMContentLoaded', function() {
    const uploadForm = document.getElementById('upload-form');
    const gridContainer = document.querySelector('.grid-container');
    const modal = document.getElementById('imageModal');
    const closeBtn = document.querySelector('.close');
    let activeTags = [];

    updateImageGrid = async function(query = '') {
        try {
            const tagsParam = activeTags.length > 0 ? activeTags.join(',') : '';
            const url = '/search' + (query ? `?q=${encodeURIComponent(query)}` : '') + 
                        (tagsParam ? `${query ? '&' : '?'}tags=${encodeURIComponent(tagsParam)}` : '');
            const response = await fetch(url);
            const images = await response.json();
            
            gridContainer.innerHTML = images.map(image => `
                <div class="grid-item">
                    <div class="image-link" onclick="showImageModal('${image.id}')">
                        <img src="${image.URL || '/static/placeholder.svg'}" alt="${image.description}" class="thumbnail" onerror="this.src='/static/placeholder.svg'; console.error('Failed to load image:', image.URL);">
                    </div>
                    <div class="filename"><span class="image-link" onclick="showImageModal('${image.id}')">${image.original_filename}</span></div>
                    <div class="description">${image.description}</div>
                    <div class="tags">${image.tags.map(tag => 
                        `<a href="#" class="tag-link ${activeTags.includes(tag) ? 'active' : ''}" onclick="event.preventDefault(); handleTagClick('${tag}');">${tag}</a>`
                    ).join(' ')}</div>
                    <div class="upload-date">${new Date(image.created_at).toLocaleDateString()}</div>
                </div>
            `).join('');
        } catch (error) {
            console.error('Error fetching images:', error);
        }
    };

    handleTagClick = function(tag) {
        const tagIndex = activeTags.indexOf(tag);
        if (tagIndex === -1) {
            activeTags.push(tag);
        } else {
            activeTags.splice(tagIndex, 1);
        }
        updateImageGrid(searchInput.value);
    };

    showImageModal = async function(imageId) {
        try {
            const response = await fetch(`/api/images/${imageId}`);
            const image = await response.json();
            
            document.getElementById('modalImage').src = image.URL;
            document.getElementById('modalDescription').textContent = image.description;
            document.getElementById('modalFilename').textContent = image.original_filename;
            document.getElementById('modalTags').innerHTML = image.tags.map(tag => 
                `<a href="#" class="tag-link" onclick="event.preventDefault(); handleTagClick('${tag}');">${tag}</a>`
            ).join(' ');
            document.getElementById('modalUploadDate').textContent = new Date(image.created_at).toLocaleString();
            document.getElementById('modalViews').textContent = image.view_count;
            
            modal.style.display = 'block';
            incrementViewCount(imageId);
        } catch (error) {
            console.error('Error fetching image details:', error);
        }
    };

    closeBtn.onclick = function() {
        modal.style.display = 'none';
    }

    window.onclick = function(event) {
        if (event.target == modal) {
            modal.style.display = 'none';
        }
    }

    async function updateImageGrid(query = '') {
        try {
            const tagsParam = activeTags.length > 0 ? activeTags.join(',') : '';
            const url = '/search' + (query ? `?q=${encodeURIComponent(query)}` : '') + 
                        (tagsParam ? `${query ? '&' : '?'}tags=${encodeURIComponent(tagsParam)}` : '');
            const response = await fetch(url);
            const images = await response.json();
            
            gridContainer.innerHTML = images.map(image => `
                <div class="grid-item">
                    <div class="image-link" onclick="showImageModal('${image.id}')">
                        <img src="${image.URL || '/static/placeholder.svg'}" alt="${image.description}" class="thumbnail" onerror="this.src='/static/placeholder.svg'; console.error('Failed to load image:', image.URL);">
                    </div>
                    <div class="filename"><span class="image-link" onclick="showImageModal('${image.id}')">${image.original_filename}</span></div>
                    <div class="description">${image.description}</div>
                    <div class="tags">${image.tags.map(tag => 
                        `<a href="#" class="tag-link ${activeTags.includes(tag) ? 'active' : ''}" onclick="event.preventDefault(); handleTagClick('${tag}');">${tag}</a>`
                    ).join(' ')}</div>
                    <div class="upload-date">${new Date(image.created_at).toLocaleDateString()}</div>
                </div>
            `).join('');
        } catch (error) {
            console.error('Error fetching images:', error);
        }
    }

    function handleTagClick(tag) {
        const tagIndex = activeTags.indexOf(tag);
        if (tagIndex === -1) {
            activeTags.push(tag);
        } else {
            activeTags.splice(tagIndex, 1);
        }
        updateImageGrid(searchInput.value);
    }

    window.handleTagClick = handleTagClick;
    window.showImageModal = showImageModal;

    uploadForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        const formData = new FormData();
        formData.append('image', uploadForm.querySelector('input[type="file"]').files[0]);
        formData.append('description', uploadForm.querySelector('textarea').value);
        formData.append('tags', uploadForm.querySelector('input[type="text"]').value);

        try {
            const response = await fetch('/upload', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error('Upload failed');
            }

            const result = await response.json();
            alert('Image uploaded successfully!');
            uploadForm.reset();
            updateImageGrid();
        } catch (error) {
            alert('Error uploading image: ' + error.message);
        }
    });

    try {
        const searchContainer = document.querySelector('.search-container');
        if (!searchContainer || !gridContainer) {
            throw new Error('Required elements not found');
        }
        window.tagSearchManager = new TagSearchManager(searchContainer);
        searchContainer.addEventListener('tagsearch', function(event) {
            activeTags = event.detail.tags || [];
            updateImageGrid(event.detail.query);
        });
        updateImageGrid();
    } catch (error) {
        console.error('Error initializing application:', error);
        const statusMessages = document.getElementById('status-messages');
        if (statusMessages) {
            statusMessages.innerHTML = `<div class="error">Error initializing application: ${error.message}</div>`;
        }
    }
});