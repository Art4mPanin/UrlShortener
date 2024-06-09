document.getElementById('login-form').addEventListener('submit', function() {
    var formData = new FormData(document.getElementById('login-form'));
    var jsonData = {};
    formData.forEach((value, key) => { jsonData[key] = value; });

    fetch('/users/login/', {
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