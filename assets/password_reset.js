$(document).ready((e) => {
    e.preventDefault();
    console.log("Reset password form submitted"); // Добавьте это сообщение
    let email = $("#password-reset-email").val();
    $.ajax({
        url: `/users/password-reset/`,
        method: "POST",
        headers: {
            'Content-Type': 'application/json'
        },
        data: JSON.stringify({ email: email })
    })
        .then(response => {
            console.log(response);
            return response.json();
        })
        .then(data => {
            console.log(data);
            if (data.message) {
                alert("На вашу почту было отправлено письмо с инструкциями по смене пароля.");
                setCookie("PassReset", email, 1)
                location.href = "/users/password-reset-new-pass/"
            } else {
                console.error("Error sending email:", data);
            }
        })
        .catch(error => console.error('Error:', error));
})
//reset password-href(email, submit)-href(code, pass,pass)