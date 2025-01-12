package endpoints

import (
	"net/http"
	"strings"
	"time"
)

var loginHTML = `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Login</title>
    <style>
        :root {
            color: #ddd;
            font-family: Arial, sans-serif;
            font-size: 1rem;
            text-align: center;
        }
        error_display {
            position: absolute;
            top: 20px;
            justify-self: center;
            padding: 1rem 2rem;
            background-color: red;
            box-shadow: 1px 1px 3px darkred;
            border-radius: 5px;
            display: none;
            justify-content: center;
            align-items: center;
            transition: all 0.3s ease-in-out;
            z-index: 1000;
        }

        body {
            display: flex;
            width: 100vw;
            height: 100vh;
            justify-content: center;
            align-items: center;
            background-color: #222;
            padding: 0;
            margin: 0;
            top: 0;
            left: 0;
        }

        login_form {
            display: flex;
            border-radius: 10px;
            box-shadow: 1px 1px 7px #111;
            background-color: #252525;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            justify-self: center;
            height: fit-content;
            padding: 20px;
            aspect-ratio: 4.2/1;
        }

        input {
            margin: 10px;
            padding: 10px 50px;
            border-radius: 5px;
            color: #ddd;
            box-shadow: 1px 1px 3px #2f2f2f;
            border: none;
            background-color: #444;
        }
    </style>
    <script>
        const displayError = (error) => {
            const errorDisplay = document.querySelector("error_display");
            errorDisplay.textContent = error;
            errorDisplay.style.display = "flex";
            setTimeout(() => {
                errorDisplay.style.display = "none";
            }, 15000);
        }
        document.addEventListener("DOMContentLoaded", async () => {
            const passwordInput = document.getElementById("password");
            const usernameInput = document.getElementById("username");

            const login = async () => {
                const username = usernameInput.value;
                const password = passwordInput.value;
                usernameInput.style.border = username.length === 0 ? "2px solid red" : "none"
                passwordInput.style.border = password.length === 0 ? "2px solid red" : "none"
                if (username.length === 0 || password.length === 0) return;
                const response = await loginVRChat(username, password).then((r) => {
                    if (typeof r !== "object")
                        throw r;
                    return r;
                }).catch((err) => {
                    displayError(err);
                    return null;
                });
                if (!response) return;
                if (response.error) return displayError(response.error.message);
                console.log(response);
                if (response.requiresTwoFactorAuth) {
                    const code = prompt("Please enter the 2FA code sent to your email or SMS: ");
                    if (!code) return;
                    const auth_cookie = response["Set-Cookie"].split(";")[0].split("=")[1]
                    const res = await loginVRChatWith2FA(auth_cookie, code).then((r) => {
                        if (typeof r !== "object")
                            throw r;
                        return r;
                    }).catch((err) => {
                        displayError(err);
                        return null;
                    });
                    if (!res) return;
                    if (res.error) return displayError(res.error.message);
                    console.log(res);
                    if (!res.verified) return displayError("2FA code is incorrect. Please try again.");
                    setCookie(auth_cookie)
                } else {
                    displayError("Something went wrong. Please try again later.")
                }
            }
            passwordInput.addEventListener("keyup", (event) => {
                if (event.code !== "Enter") return;
                event.preventDefault();
                login();
            });
            usernameInput.addEventListener("keyup", (event) => {
                if (event.code !== "Enter") return;
                if (passwordInput.value.length > 0) {
                    event.preventDefault();
                    return login();
                }
                event.preventDefault();
                passwordInput.focus()
            });
            const loginButton = document.getElementById("login");
            loginButton.addEventListener("click", login);

        });
    </script>
</head>
<body>
    <error_display>
        asdf
    </error_display>
    <login_form>
            <label for="username">Username</label>
            <input type="text" id="username" name="username" required>
            <label for="password">Password</label>
            <input type="password" id="password" name="password" required>
            <button type="submit" style="background-color: #4CAF50; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer;" id="login">Login</button>
    </login_form>
</body>
</html>`

func init() {
	RegisterEndpoint("login", func(w http.ResponseWriter, r *http.Request) {
		//http.ServeFile(w, r, "static/login.html")
		http.ServeContent(w, r, "index.html", time.Now(), strings.NewReader(loginHTML))
	})
}
