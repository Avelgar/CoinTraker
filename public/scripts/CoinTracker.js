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
            console.log("Форма регистрации отправлена");
            this.closeModal();
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
