$(document).ready(() => {
    let email = getCookie("Verification");
    console.log("Email from cookie:", email);  // Debugging line

    $("#verify-button").click((e) => {
        e.preventDefault();
        let code = $("#verification_code").val();
        console.log("Code from input:", code);  // Debugging line

        $.ajax({
            url: '/users/verification/',
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            data: JSON.stringify({
                code: code,
                email: email
            }),
            success: (response) => {
                console.log("AJAX success response:", response);  // Debugging line
                if (response.success) {
                    alert("Аккаунт успешно подтвержден.");
                    location.href = "/users/login/";
                } else {
                    alert(response.message || "Неверный код подтверждения.");
                }
            },
            error: (response) => {
                alert("Ошибка при подтверждении аккаунта: " + response.responseText);
                console.error("AJAX error response:", response);  // Debugging line
            }
        });
    });
});