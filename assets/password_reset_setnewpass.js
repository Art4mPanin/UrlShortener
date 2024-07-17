$(document).ready((e) => {
    e.preventDefault();
    console.log("Reset password form submitted")
    let email = getCookie("PassReset");
    let code = $("#code").val();
    let password = $("#new-password").val();
    let repeat_password = $("#repeat-password").val();
    $.ajax({
        url: `/users/password-reset-new-pass/`,
        method: "POST",
        headers: {
            'Content-Type': 'application/json'
        },
        data: JSON.stringify({
            email: email,
            code: code,
            newpass: password,
            newpassrep: repeat_password,

        })
    })
        .then(response => {
            console.log(response);
            return response.json();
        })
        .then(data => {
            console.log(data);
            if (data.message) {
                alert("Data was changed");
                location.href = "/users/login/"
            } else {
                console.error("Error sending email:", data);
            }
        })
        .catch(error => console.error('Error:', error));
})