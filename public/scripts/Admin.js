document.addEventListener('DOMContentLoaded', function() {
    // Переключение бокового меню
    document.getElementById('menuToggle').addEventListener('click', function() {
        const sidebar = document.getElementById('sidebar');
        sidebar.classList.toggle('active');
    });

    checkAuthenticationAndLoadCoins();
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
        isExitModalOpen: false
    },
    methods: {
        openExitModal() {
            this.isExitModalOpen = true;
        },
        submitExitForm() {
            this.closeExitModal();
        },
        closeExitModal() {
            this.isExitModalOpen = false;
        }
    },
    watch: {
        isExitModalOpen(newValue) {
            this.$nextTick(() => {
                const modal = document.querySelector('.modal');
                if (modal) {
                    modal.style.visibility = newValue ? 'visible' : 'hidden'; 
                }
            });
        },
    }
});
