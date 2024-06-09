document.getElementById('signup-form').addEventListener('submit', function() {
    var formData = new FormData(document.getElementById('signup-form'));
    var jsonData = {};
    var birthDate = new Date(formData.get('birth_date')).toISOString();
    formData.set('birth_date', birthDate);

    formData.forEach((value, key) => { jsonData[key] = value; });

    fetch('/users/signup/', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(jsonData)
    })
        .then(response => response.json())
        .then(data => console.log(data))
        .catch(error => console.error('Error:', error));
});
