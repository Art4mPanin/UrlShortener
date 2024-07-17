$(document).ready(() => {
    const userID = window.location.pathname.split('/').pop();

    let token = getCookie('Authorization');
    let data = parseJwt(token)
    console.log(data)
    let user_id = data.sub
    console.log(user_id)
    $.ajax({
        url: `/api/users/profile/${userID}`,
        method: 'GET',
        dataType: 'json',
        success: (data) => {
            if (data.avatar_url) {
                $('#profileImage').attr('src', data.avatar_url);
            }
            $('#displayName').val(data.displayed_name);
            $('#profileTitle').val(data.profile_title);
            $('#bio').val(data.bio);
            $('#email').val(data.email);
            $('#').val(data.google_id);
        },
        error: (jqXHR, textStatus, errorThrown) => {
            console.error('Error:', textStatus, errorThrown);
            alert('Error loading profile data or invalid user id');
        }
    });

    $("#updateAvatarButton").click(function () {
        let file_inp = $("#avatarInput")[0].files[0];
        if (file_inp && (file_inp.name.endsWith("png") || file_inp.name.endsWith("jpg") || file_inp.name.endsWith("jpeg"))) {
            let formData = new FormData();
            formData.append("avatar", file_inp);
            $.ajax({
                method: 'PUT',
                url: `/users/update-avatar/${user_id}`,
                data: formData,
                processData: false,
                contentType: false,
                success: (response) => {
                    if (response.avatar_url) {
                        $('#profileImage').attr('src', response.avatar_url);
                        alert("Avatar has been changed");
                    } else {
                        alert("Ошибка в ответе сервера: не найден URL аватара.");
                    }
                },
                error: (response) => {
                    alert("Ошибка при загрузке аватара: " + response.responseText);
                }
            });
        } else {
            alert('Файл должен быть формата png, jpg или jpeg.');
        }
    });
    $("#changedata").click((e) => {
        e.preventDefault();

        let displayedname = $("#displayName").val();
        let proftitle = $("#profileTitle").val();
        let bio = $("#bio").val();
        let email = $("#email").val();
        $.ajax({
            url: `/users/update-data/${user_id}`,
            method: "PUT",
            headers: {
                "Content-Type": "application/json"
            },
            data: JSON.stringify({
                displayed_name: displayedname,
                profile_title: proftitle,
                bio: bio,
                Email: email,
            }),
            success: (data) => {
                alert("Data was changed successfully!");
            },
            error: (error) => {
                alert("Error: " + error.responseText);
            }
        });
    });
    $("#ConfirmNewPass").click((e) => {
        e.preventDefault();
        let oldpass = $("#oldPassword").val();
        let newpass = $("#newPassword").val();
        let newpassconfirm = $("#confirmPassword").val();
        $.ajax({
            url: `/users/update-password/${user_id}`,
            method: "PUT",
            headers: {
                "Content-Type": "application/json"
            },
            data: JSON.stringify({
                oldpass: oldpass,
                newpass: newpass,
                newpassconfirm: newpassconfirm,
            }),
            success: (data) => {
                alert("Pass was changed successfully!");
            },
            error: (error) => {
                alert("Error: " + error.responseText);
            }
        });
    })
    $("#LogOut").click((e) => {
        deleteCookie("Authorization")
        location.href = "/users/login/"
    })
    $("#unlink_google").click((e) =>
    {
        e.preventDefault();
        $.ajax({
            url: `/users/unlink_google/${user_id}`,
            method: "PUT",
            headers: {
                "Content-Type": "application/json"
            },
            success: (data) => {
                alert("Google аккаунт успешно отвязан!");
                // $("#unlink_google").hide();
            },
            error: (error) => {
                alert("Error: " + error.responseText);
            }
        });
    }
)
});