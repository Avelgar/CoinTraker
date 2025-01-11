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
});
   

new Vue({
    
    el: '#app',
    data: {
        isSignUpModalOpen: false,
        isLogInModalOpen: false
    },
    methods: {
        closeModal() {
            console.log("Закрытие модального окна");
            this.isSignUpModalOpen = false;
            this.isLogInModalOpen = false;
        },
        openSignUpModal() {
            console.log("Открытие модального окна регистрации");
            this.isSignUpModalOpen = true;
        },
        submitSignUpForm() {
            const login = document.getElementById('SignUpLogin').value;
            const email = document.getElementById('SignUpEmail').value;
            const password = document.getElementById('SignUpPassword').value;
            const password2 = document.getElementById('SignUpPassword2').value;
            
            
            fetch('/register', {
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
            .then(response => response.json())
            .then(data => {
                console.log('Успех:', data);
                this.closeModal();
            })
            .catch((error) => {
                console.error('Ошибка:', error);
            });
        },
        openLogInModal(){
            console.log("Открытие модального окна авторизации");
            this.isLogInModalOpen = true;
        },
        submitLogInForm() {
            console.log("Форма авторизации отправлена");
            this.closeModal();
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
        }
    }
});
