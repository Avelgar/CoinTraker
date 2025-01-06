const themeToggles = document.querySelectorAll('.themeToggle');

function setTheme(theme) {
    document.body.classList.toggle('dark-theme', theme === 'dark');
}

function toggleTheme() {
    const currentTheme = localStorage.getItem('theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    localStorage.setItem('theme', newTheme);
    setTheme(newTheme);
}

themeToggles.forEach(button => {
    button.addEventListener('click', toggleTheme);
});

const savedTheme = localStorage.getItem('theme') || 'light';
setTheme(savedTheme);