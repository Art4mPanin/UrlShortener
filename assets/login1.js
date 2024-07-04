$(document).ready(() => {
    $("#submit").click((e) => {
        e.preventDefault();
        console.log("Form submitted"); // Добавьте это сообщение
        let formData = new FormData(document.getElementById('login-form'));
        let jsonData = {};
        formData.forEach((value, key) => { jsonData[key] = value; });

        fetch('/users/login/', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(jsonData)
        })
            .then(response => {
                console.log(response);
                return response.json();
            })
            .then(data => {
                console.log(data);
                if (data.token) {
                    setCookie("Authorization", data.token, 30);
                    console.log("Token set in cookie:", data.token);
                } else {
                    console.error("Token not found in response:", data);
                }
            })
            .catch(error => console.error('Error:', error));
displayerror()
        let token = getCookie('Authorization');
        let data = parseJwt(token)
        let user_id = data.sub
        location.href = `/users/profile/${user_id}`
    });
});
function displayerror