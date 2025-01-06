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
        isModalVisible: false
    },
    methods: {
        openModal() {
            console.log("Открытие модального окна");
            this.isModalVisible = true;
        },
        closeModal() {
            console.log("Закрытие модального окна");
            this.isModalVisible = false;
        },
        submitForm() {
            console.log("Форма отправлена");
            this.closeModal();
        }
    },
    watch: {
        isModalVisible(newValue) {
            this.$nextTick(() => {
                const modal = document.querySelector('.modal');
                if (modal) {
                    modal.style.visibility = newValue ? 'visible' : 'hidden'; 
                }
            });
        }
    }
});
