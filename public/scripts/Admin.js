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

    function checkAuthenticationAndLoadCoins() {
        fetch('/api/authenticate')
            .then(response => {
                if (!response.ok) {
                    // Если пользователь не аутентифицирован, перенаправляем на главную страницу
                    window.location.href = '/public/CoinTracker.html';
                    return;
                }
                return response.json();
            })
            .then(data => {
                if (data && data.success) {
                    displayCoins(data.coins);
                }
            })
            .catch(error => {
                console.error("Ошибка при проверке аутентификации:", error);
                // В случае ошибки также перенаправляем на главную страницу
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

