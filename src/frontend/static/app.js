document.addEventListener('DOMContentLoaded', function() {
    const uploadForm = document.getElementById('upload-form');
    const searchInput = document.getElementById('search');
    const resultsDiv = document.getElementById('results');
    const autocompleteResults = document.getElementById('autocomplete-results');
    const recentUploadsBody = document.getElementById('recent-uploads-body');
    const recentUploadsContainer = document.querySelector('.grid-container');
    const searchResultsContainer = document.getElementById('search-results-body');

    async function fetchRecentUploads() {
        try {
            const response = await fetch('/search');
            const images = await response.json();
            
            recentUploadsContainer.innerHTML = images.map(image => `
                <div class="grid-item">
                    <a href="/image/${image.id}">
                        <img src="${image.URL || '#'}" alt="${image.description}" class="thumbnail" onerror="this.src='/static/placeholder.svg'">
                    </a>
                    <div class="filename"><a href="/image/${image.id}">${image.original_filename}</a></div>
                    <div class="description">${image.description}</div>
                    <div class="tags">${image.tags.map(tag => `<a href="/search?q=${tag}" class="tag-link">${tag}</a>`).join(' ')}</div>
                    <div class="upload-date">${new Date(image.created_at).toLocaleDateString()}</div>
                </div>
            `).join('');
        } catch (error) {
            console.error('Error fetching recent uploads:', error);
        }
    }

    fetchRecentUploads();

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
            fetchRecentUploads(); // Refresh the recent uploads table
        } catch (error) {
            alert('Error uploading image: ' + error.message);
        }
    });

    let searchTimeout;
    searchInput.addEventListener('input', function() {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(async () => {
            const query = searchInput.value;
            if (query.length < 2) {
                searchResultsContainer.innerHTML = '';
                return;
            }

            try {
                const response = await fetch(`/search?q=${encodeURIComponent(query)}`);
                const images = await response.json();
                
                searchResultsContainer.innerHTML = images.map(image => `
                    <div class="grid-item">
                        <a href="/image/${image.id}">
                            <img src="${image.URL || '#'}" alt="${image.description}" class="thumbnail" onerror="this.src='/static/placeholder.svg'">
                        </a>
                        <div class="filename"><a href="/image/${image.id}">${image.original_filename}</a></div>
                        <div class="description">${image.description}</div>
                        <div class="tags">${image.tags.map(tag => `<a href="/search?q=${tag}" class="tag-link">${tag}</a>`).join(' ')}</div>
                        <div class="upload-date">${new Date(image.created_at).toLocaleDateString()}</div>
                    </div>
                `).join('');
            } catch (error) {
                console.error('Search error:', error);
            }
        }, 300);
    });
});