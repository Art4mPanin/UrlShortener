document.addEventListener('DOMContentLoaded', function () {
    console.log('DOM fully loaded and parsed');

    var updateAvatarButton = document.getElementById('updateAvatarButton');
    if (updateAvatarButton) {
        console.log('Button found');
        updateAvatarButton.addEventListener('click', function () {
            console.log('Button clicked');
            var fileInput = document.getElementById('avatarInput');
            var file = fileInput.files[0];

            if (!file) {
                alert('Пожалуйста, выберите файл для загрузки.');
                console.log('No file selected');
                return;
            }

            var formData = new FormData();
            formData.append('avatar', file);

            console.log('Preparing to send fetch request');
            fetch('/user/update-avatar/', {
                method: 'POST',
                body: formData
            })
                .then(response => {
                    console.log('Response:', response);
                    return response.json();
                })
                .then(data => {
                    console.log('Data:', data);
                    if (data.message === "File uploaded successfully") {
                        document.getElementById('profileImage').src = data.avatar_url;
                    } else {
                        alert('Ошибка при обновлении аватара: ' + data.message);
                    }
                })
                .catch(error => console.error('Ошибка:', error));
        });
    } else {
        console.log('Button not found');
    }
});
