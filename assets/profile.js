document.getElementById('updateAvatarButton').addEventListener('click', function() {
    var fileInput = document.getElementById('avatarInput');
    var file = fileInput.files[0];

    if (!file) {
        alert('Пожалуйста, выберите файл для загрузки.');
        return;
    }

    var formData = new FormData();
    formData.append('avatar', file);

    fetch('/user/update-avatar/', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.message === "File uploaded successfully") {
                document.getElementById('profileImage').src = data.avatar_url;
                console.log("Avatar URL: ", data.avatar_url);
            } else {
                alert('Ошибка при обновлении аватара: ' + data.message);
            }
        })
        .catch(error => console.error('Ошибка:', error));
});
