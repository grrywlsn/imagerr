document.addEventListener('DOMContentLoaded', function() {
    const uploadForm = document.getElementById('upload-form');
    const searchInput = document.getElementById('search');
    const autocompleteResults = document.getElementById('autocomplete-results');
    const gridContainer = document.querySelector('.grid-container');
    const modal = document.getElementById('imageModal');
    const closeBtn = document.querySelector('.close');
    let activeTags = [];

    async function showImageModal(imageId) {
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
    }

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

    // Load initial images
    updateImageGrid();

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
            updateImageGrid(); // Refresh the image grid
        } catch (error) {
            alert('Error uploading image: ' + error.message);
        }
    });

    let searchTimeout;
    searchInput.addEventListener('input', function() {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
            const query = searchInput.value;
            if (query.length < 2) {
                updateImageGrid(); // Show recent uploads when search is cleared
                return;
            }
            updateImageGrid(query);
        }, 300);
    });
});