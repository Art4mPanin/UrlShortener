$(document).ready(() => {
    $("#submit-button").click((e) => {
        e.preventDefault();

        let username = $("#username").val();
        let password = $("#password").val();
        let email = $("#email").val();
        let repeat_password = $("#repeat_password").val();
        let birth_date = $("#birth_date").val();

        $.ajax({
            url: `/users/signup/`,
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            data: JSON.stringify({
                username: username,
                password: password,
                repeat_password: repeat_password,
                email: email,
                birth_date: birth_date
            }),
            success: (data) => {
                alert("Registration successful!");
            },
            error: (error) => {
                alert("Error: " + error.responseText);
            }
        });
    });
});
