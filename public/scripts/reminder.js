function getRegToken() {
    const url = new URL(window.location.href);
    const Regtoken = url.searchParams.get('token');
    if (!Regtoken) {
        return;
    }
    else{
        return Regtoken;
    }
}
document.addEventListener('DOMContentLoaded', function () {
    document.querySelector('.scroll-up').onclick = scrollToTop;

    const Regtoken = getRegToken()
    if (Regtoken){
        fetch('/submitEmail', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ Regtoken })
        })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    alert('Ссылка на подтверждение по который вы перешли ранее не работает');
                    window.location.href = '/reminder.html';
                    throw new Error('submit failed');
                });
            }
            return response.text();
        })
        .then(response => {
                window.location.href = '/reminder.html';
        })
        .catch(error => {
            console.error(error);
        });
    }
    var signinModal = document.getElementById("Signin");
    var loginModal = document.getElementById("Login");
    var recoveryModal = document.getElementById("Recovery");
    var captchaModal = document.getElementById("Captcha");
    var openSigninBtns = document.getElementsByClassName("openSigninForm");
    var openLoginBtns = document.getElementsByClassName("openLoginForm");
    var openCaptchaLinks = document.getElementsByClassName("openCaptchaForm");
    var closeBtns = document.getElementsByClassName("close");
    const recoveryMessage = document.getElementById("recoveryMessage");
    const captchaErrorMessage = document.getElementById("captchaErrorMessage");
    

    Array.from(openSigninBtns).forEach(button => {
        button.onclick = function () {
            signinModal.style.display = "block";
        }
    });

    Array.from(openLoginBtns).forEach(button => {
        button.onclick = function () {
            loginModal.style.display = "block";
        }
    });

    Array.from(openCaptchaLinks).forEach(link => {
        link.onclick = function () {
            captchaModal.style.display = "block";
        }
    });

    Array.from(closeBtns).forEach(button => {
        button.onclick = function () {
            this.parentElement.parentElement.style.display = "none";
            captchaErrorMessage.style.display = "none";
        }
    });

    window.onclick = function (event) {
        if (event.target == signinModal || event.target == loginModal || event.target == recoveryModal || event.target == captchaModal) {
            event.target.style.display = "none";
            captchaErrorMessage.style.display = "none";
        }
    }

    document.getElementById("LoginForm").onsubmit = function (event) {
        event.preventDefault();
        const login = document.getElementById("loginEmail").value;
        const password = document.getElementById("loginPassword").value;
    
        document.getElementById("loginErrorMessage").style.display = "none";
        document.getElementById("bannedMessage").style.display = "none";
        document.getElementById("nonExistentMessage").style.display = "none";
        document.getElementById("LockedOutMessage").style.display = "none";
        document.getElementById("Notsubmited").style.display = "none";
        document.getElementById("Onrecovery").style.display = "none";
    
        fetch('/reminder.html/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ login, password })
        })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    if (text.includes('User not found')) {
                        document.getElementById("nonExistentMessage").style.display = "block";
                    } else if (text.includes('Invalid password')) {
                        document.getElementById("loginErrorMessage").style.display = "block";
                    } else if (text.includes('User is banned')) {
                        document.getElementById("bannedMessage").style.display = "block";
                    } else if (text.includes('Too many failed login attempts. Please try again later.')) {
                        document.getElementById("LockedOutMessage").style.display = "block";
                    } else if(text.includes('Not submited')){
                        document.getElementById("Notsubmited").style.display = "block";
                    } else if(text.includes('On recovery')){
                        document.getElementById("Onrecovery").style.display = "block";
                    }
                    throw new Error('Login failed');
                });
            }
            return response.json();
        })
        .then(data => {
            if (data.message === 'Login successful') {
                const userId = data.user_id;
                window.location.href = `/user.html`;
            } else if (data.message === 'Login as admin') {
                const userId = data.user_id;
                window.location.href = `/admin.html`;
            } else {
                console.log('Unexpected response:', data);
            }
        })
        .catch(error => {
            console.error(error);
        });        
    };

    
    document.getElementById("SignInForm").onsubmit = function (event) {
        event.preventDefault();
        const email = document.getElementById("email").value;
        const password = document.getElementById("password").value;
        const confirmPassword = document.getElementById("password2").value;
        const login = document.getElementById("login").value;
    
        document.getElementById("registerSuccessMessage").style.display = "none";
        document.getElementById("userExistsMessage").style.display = "none";
        document.getElementById("passwordMismatchMessage").style.display = "none";
        document.getElementById("loginExistsMessage").style.display = "none";
        document.getElementById("PasswordIsTooWeak").style.display = "none";
        document.getElementById("recoveryMessageAlreadyreg").style.display = "none";

        if (password !== confirmPassword) {
            document.getElementById("passwordMismatchMessage").style.display = "block";
            return;
        }
        fetch('/reminder.html/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password, login })
        })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    if (text.includes('User already exists')) {
                        document.getElementById("userExistsMessage").style.display = "block";
                    } 
                    else if (text.includes('Login exists')) {
                        document.getElementById("loginExistsMessage").style.display = "block";
                    }
                    else if(text.includes('Password is too weak reg')) {
                        document.getElementById("PasswordIsTooWeak").style.display = "block";
                    }
                    else if (text.includes('Is already reg pending')) {
                        document.getElementById("recoveryMessageAlreadyreg").style.display = "block";
                    }
                    throw new Error('Registration failed');
                });
            }
            return response.text();
        })
        .then(message=>{
            if (message === 'User registered successfully') {
                document.getElementById("registerSuccessMessage").style.display = "block";
            }
        }
        )
        .catch(error => {
            console.error(error);
        });
    };
    

    document.getElementById("RecoveryForm").onsubmit = function (event) {
        event.preventDefault();
        recoveryMessage.style.display = "none";
        document.getElementById("recoveryMessageNoUser").style.display = "none";
        document.getElementById("recoveryMessageAlready").style.display = "none";
        document.getElementById("NotsubmitedRec").style.display = "none";

        const email = document.getElementById("recoveryEmail").value;
        fetch('/reminder.html/recovery', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email })
        })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    if (text.includes('User not found')) {
                        document.getElementById("recoveryMessageNoUser").style.display = "block";
                    }
                    else if (text.includes('Is already pending')) {
                        document.getElementById("recoveryMessageAlready").style.display = "block";
                    } else if(text.includes('On reg recovery')) {
                        document.getElementById("NotsubmitedRec").style.display = "block";
                    }
                    throw new Error('Recovery failed');
                });
            }
            return response.text();
        })
        .then(message=>{
            if (message === 'Password recovery email sent. Please check your inbox.') {
                recoveryMessage.style.display = "block";
            }
        }
        )
    };

    document.getElementById("CaptchaForm").onsubmit = function (event) {
        event.preventDefault();
        if (grecaptcha.getResponse() == ""){
            captchaErrorMessage.style.display = "block";
            grecaptcha.reset();
        } else {
            captchaModal.style.display = "none";
            recoveryModal.style.display = "block";
            grecaptcha.reset();
        }
    };
});