document.addEventListener('DOMContentLoaded', function() {
    document.getElementById('menuToggle').addEventListener('click', function() {
        const sidebar = document.getElementById('sidebar');
        sidebar.classList.toggle('active');
    });

    document.addEventListener('click', function(event) {
        const sidebar = document.getElementById('sidebar');
        const menuToggle = document.getElementById('menuToggle');

        if (!sidebar.contains(event.target) && !menuToggle.contains(event.target)) {
            sidebar.classList.remove('active');
        }
    });

    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');

    if (token) {
        fetch('/api/checkToken', {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ token: token })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error("Ошибка при проверке токена");
            }
            return response.json();
        })
        .then(data => {
            if (data.success) {
            } else {
                alert("Токен просрочен или недействителен.");
                window.location.href = "/public/CoinTracker.html"; 
            }
        })
        .catch(error => {
            console.error("Ошибка:", error);
            alert("Произошла ошибка при проверке токена.");
            window.location.href = "/public/CoinTracker.html";
        });
    }
    else{
            fetch('/api/checkCookie')
            .then(response => {
                return response.json();
            })
            .then(data => {
                if (data.success) {
                    window.location.href = "/public/User.html";
                }
            })
            .catch(error => {
                console.error('There was a problem with the fetch operation:', error);
            });
    }
});


new Vue({
    el: '#app',
    data: {
        isSignUpModalOpen: false,
        isLogInModalOpen: false,
        isCapchaModalOpen: false,
        isRecoveryModalOpen: false
    },
    methods: {
        showNotification(message, type) {
            const notification = document.createElement('div');
            notification.className = `notification ${type}`;
            notification.innerText = message;

            document.getElementById('notifications').appendChild(notification);
            notification.style.display = 'block';

            setTimeout(() => {
                notification.style.display = 'none';
                notification.remove();
            }, 3000);
        },
        openSignUpModal() {
            this.isSignUpModalOpen = true;
        },
        submitSignUpForm() {
            const login = document.getElementById('SignUpLogin').value;
            const email = document.getElementById('SignUpEmail').value;
            const password = document.getElementById('SignUpPassword').value;
            const password2 = document.getElementById('SignUpPassword2').value;
        
            if (password !== password2) {
                this.showNotification('Пароли не совпадают!', 'error');
                return;
            }
        
            fetch('/SignUp', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    login: login,
                    email: email,
                    password: password,
                }),
            })
            .then(response => {
                if (!response.ok) {
                    return response.text().then(text => {
                        if (text.includes('PasswordIsTooWeak')) {
                            return this.showNotification('Пароль слишком слабый.', 'error');
                        } else if (text.includes('UserAlreadyExistsWithEmailAndNoToken')) {
                            return this.showNotification('Пользователь с таким email уже существует.', 'error');
                        } else if (text.includes('UserAlreadyExistsWithEmailAndHasToken')) {
                            return this.showNotification('На эту почту уже отправлена ссылка на поддтверждение.', 'error');
                        } else if (text.includes('UserAlreadyExistsWithLogin')) {
                            return this.showNotification('Это имя пользователя уже занято', 'error');
                        } else if (text.includes('Bad request')) {
                            return this.showNotification('Ошибка базы данных. Попробуйте позже.', 'error');
                        } else if (text.includes('InternalServerError')) {
                            return this.showNotification('Ошибка сервера. Попробуйте позже.', 'error');
                        } else {
                            return this.showNotification('Неизвестная ошибка. Попробуйте снова.', 'error');
                        }
                    });
                }
                else{
                    this.showNotification('Подтвердите аккаунта в своем почтовом ящике!', 'success');
                    this.closeSignUpModal();
                }
            })
            .catch((error) => {
                console.error('Ошибка:', error);
                this.showNotification('Ошибка регистрации. Попробуйте еще раз.', 'error');
            });
        },
        closeSignUpModal() {
            this.isSignUpModalOpen = false;
        },
        openLogInModal() {
            this.closeSignUpModal(); // Закрываем окно регистрации
            this.isLogInModalOpen = true; // Открываем окно авторизации
            const signUpLogin = document.getElementById('SignUpLogin');
            const signUpEmail = document.getElementById('SignUpEmail');
            const signUpPassword = document.getElementById('SignUpPassword');
            const signUpPassword2 = document.getElementById('SignUpPassword2');
        
            if (signUpLogin) signUpLogin.value = '';
            if (signUpEmail) signUpEmail.value = '';
            if (signUpPassword) signUpPassword.value = '';
            if (signUpPassword2) signUpPassword2.value = '';
        },
        submitLogInForm() {
            const login = document.getElementById('LogInEmail').value;
            const password = document.getElementById('LogInPassword').value;
            const rememberMe = document.getElementById('rememberMe').checked;
        
            fetch('/LogIn', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    login: login,
                    password: password,
                    rememberMe: rememberMe,
                }),
            })
            .then(response => {
                if (!response.ok) {
                    return response.text().then(text => {
                        if (text.includes('UserNotFound')) {
                            return this.showNotification('Пользователь не найден.', 'error');
                        } else if (text.includes('UserIsBanned')) {
                            return this.showNotification('Пользователь забанен.', 'error');
                        } else if (text.includes('InvalidCredentials')) {
                            return this.showNotification('Неверный логин или пароль.', 'error');
                        } else if (text.includes('UserHasToken')) {
                            return this.showNotification('Аккаунт на подтверждении, проверьте почту.', 'error');
                        } else if (text.includes('UserHasRecoveryToken')) {
                            return this.showNotification('Аккаунт на восстановлении, проверьте почту.', 'error');
                        } else if (text.includes('InternalServerError')) {
                            return this.showNotification('Ошибка сервера!', 'error');
                        } else if (text.includes('Bad request')) {
                            return this.showNotification('Плохое соединение!', 'error');
                        } else {
                            return this.showNotification('Неизвестная ошибка. Попробуйте снова.', 'error');
                        }
                    });
                } else {
                    this.showNotification('Вход выполнен успешно!', 'success');
                    this.closeLogInModal();
                            window.location.href = '/public/User.html'; // Перенаправляем на страницу пользователя
                    }
            })
            .catch((error) => {
                console.error('Ошибка:', error);
                this.showNotification('Ошибка входа. Попробуйте еще раз.', 'error');
            });
        },
        closeLogInModal(){
            this.isLogInModalOpen = false;
        },
        openCapchaModal() {
            this.isCapchaModalOpen = true; // Открываем окно капчи
        },
        submitCapchaForm() {
            // if (grecaptcha.getResponse() == ""){
            //     this.showNotification('Капча не пройдена. Попробуйте еще раз.', 'error');
            //     grecaptcha.reset();
            // } else {
            //     this.closeCapchaModal();
            //     grecaptcha.reset();
            // }
            this.openRecoveryModal();
        },
        closeCapchaModal() {
            this.isCapchaModalOpen = false;
        },
        openRecoveryModal(){
            this.closeCapchaModal();
            this.isRecoveryModalOpen = true;
        },
        submitRecoveryForm(){
            const email = document.getElementById('RecoveryEmail').value;
            fetch('/Recovery', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    email: email,
                }),
            })
            .then(response => {
                if (!response.ok) {
                    return response.text().then(text => {
                        if (text.includes('UserNotFound')) {
                            return this.showNotification('Пользователь не найден.', 'error');
                        } else if (text.includes('UserIsBanned')) {
                            return this.showNotification('Пользователь забанен.', 'error');
                        } else if (text.includes('UserHasToken')) {
                            return this.showNotification('Аккаунт на подтверждении, проверьте почту.', 'error');
                        } else if (text.includes('UserHasRecoveryToken')) {
                            return this.showNotification('Аккаунт на восстановлении, проверьте почту.', 'error');
                        } else if (text.includes('InternalServerError')) {
                            return this.showNotification('Ошибка сервера!', 'error');
                        } else if (text.includes('Bad request')) {
                            return this.showNotification('Плохое соединение!', 'error');
                        } else {
                            return this.showNotification('Неизвестная ошибка. Попробуйте снова.', 'error');
                        }
                    });
                } else {
                    this.showNotification('Письмо с ссылкой на восстановление пароля отправлено!.', 'success');
                    this.closeRecoveryModal();
                }
            })
            .catch((error) => {
                console.error('Ошибка:', error);
                this.showNotification('Ошибка входа. Попробуйте еще раз.', 'error');
            });
        },
        closeRecoveryModal(){
            this.isRecoveryModalOpen = false;
        }
    },
    watch: {
        isSignUpModalOpen(newValue) {
            this.$nextTick(() => {
                const modal = document.querySelector('.modal');
                if (modal) {
                    modal.style.visibility = newValue ? 'visible' : 'hidden'; 
                }
            });
        },
        isLogInModalOpen(newValue) {
            this.$nextTick(() => {
                const modal = document.querySelector('.modal');
                if (modal) {
                    modal.style.visibility = newValue ? 'visible' : 'hidden'; 
                }
            });
        },
        isCapchaModalOpen(newValue) {
            this.$nextTick(() => {
                const modal = document.querySelector('.modal-capcha');
                if (modal) {
                    modal.style.visibility = newValue ? 'visible' : 'hidden'; 
                }
            });
        },
        isRecoveryModalOpen(newValue) {
            this.$nextTick(() => {
                const modal = document.querySelector('.modal-capcha');
                if (modal) {
                    modal.style.visibility = newValue ? 'visible' : 'hidden'; 
                }
            });
        }
    }
});

