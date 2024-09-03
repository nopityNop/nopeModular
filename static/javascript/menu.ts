document.addEventListener("DOMContentLoaded", () => {
    const accountButton = document.getElementById("accountButton");
    const menu = document.getElementById("account-button-menu");

    accountButton?.addEventListener("click", () => {
        if (menu?.style.display === "block") {
            menu.style.display = "none";
        } else {
            menu.style.display = "block";
        }
    });

    document.addEventListener("click", (event) => {
        if (menu?.style.display === "block" && !accountButton?.contains(event.target as Node) && !menu.contains(event.target as Node)) {
            menu.style.display = "none";
        }
    });
});
