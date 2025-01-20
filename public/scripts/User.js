document.addEventListener('DOMContentLoaded', function() {
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

    function checkAuthenticationAndLoadCoins() {
        fetch('/api/authenticate')
            .then(response => {
                if (!response.ok) {
                    window.location.href = '/public/CoinTracker.html';
                    return;
                }
                return response.json();
            })
            .then(data => {
                if (data && data.success) {
                    displayCoins(data.coins);
                    if (data.is_admin) {
                        document.getElementById('adminPanelLink').style.display = 'contents';
                    }
                } else {
                    // Если success: false, но пользователь уже аутентифицирован, не перенаправляем
                    console.warn("Аутентификация не удалась, но пользователь может быть уже в системе.");
                }
            })
            .catch(error => {
                console.error("Ошибка при проверке аутентификации:", error);
                window.location.href = '/public/CoinTracker.html';
            });
    }    

    function displayCoins(coins) {
        const coinsListElement = document.getElementById('coins'); // Предполагаем, что у вас есть элемент с id 'coins'
        coinsListElement.innerHTML = ''; // Очищаем предыдущий список

        coins.forEach(coin => {
            const listItem = document.createElement('li');
            listItem.textContent = coin; // Добавляем название монеты
            coinsListElement.appendChild(listItem);
        });
    }
});

const channel = new BroadcastChannel('auth_channel');

channel.onmessage = (event) => {
    if (event.data === 'logout') {
        window.location.href = '/public/CoinTracker.html';
    }
};

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
            fetch('/api/logout', {
                method: 'POST',
            })
            .then(response => {
                if (response.ok) {
                    channel.postMessage('logout');
                    window.location.href = '/public/CoinTracker.html'; // Перенаправление на главную страницу
                } else {
                    console.error("Ошибка при выходе:", response.statusText);
                }
            })
            .catch(error => {
                console.error("Ошибка при выходе:", error);
            });
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


