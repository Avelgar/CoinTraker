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

document.querySelector('.arrow').addEventListener('click', function(e) {
    e.preventDefault();
    document.querySelector('#next-section').scrollIntoView({ 
        behavior: 'smooth' 
    });
});

document.querySelector('.learn-more-button').addEventListener('click', function(event) {
    event.preventDefault();
    document.querySelector('#advantages').scrollIntoView({
        behavior: 'smooth'
    });
});

