document.addEventListener('DOMContentLoaded', function() {
    const uploadForm = document.getElementById('upload-form');
    const searchInput = document.getElementById('search');
    const resultsDiv = document.getElementById('results');
    const autocompleteResults = document.getElementById('autocomplete-results');

    uploadForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        const formData = new FormData();
        formData.append('image', uploadForm.querySelector('input[type="file"]').files[0]);
        formData.append('description', uploadForm.querySelector('textarea').value);
        formData.append('tags', uploadForm.querySelector('input[type="text"]').value);

        try {
            const response = await fetch('/api/images', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error('Upload failed');
            }

            const result = await response.json();
            alert('Image uploaded successfully!');
            uploadForm.reset();
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
                autocompleteResults.innerHTML = '';
                return;
            }

            try {
                const response = await fetch(`/api/images/search?q=${encodeURIComponent(query)}`);
                const images = await response.json();
                
                resultsDiv.innerHTML = images.map(image => `
                    <div class="image-card">
                        <img src="${image.URL}" alt="${image.description}">
                        <p>${image.description}</p>
                        <p class="tags">${image.tags.join(', ')}</p>
                    </div>
                `).join('');
            } catch (error) {
                console.error('Search error:', error);
            }
        }, 300);
    });
});