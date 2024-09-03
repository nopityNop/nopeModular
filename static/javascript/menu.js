document.addEventListener("DOMContentLoaded", function () {
    var accountButton = document.getElementById("accountButton");
    var menu = document.getElementById("account-button-menu");
    accountButton === null || accountButton === void 0 ? void 0 : accountButton.addEventListener("click", function () {
        if ((menu === null || menu === void 0 ? void 0 : menu.style.display) === "block") {
            menu.style.display = "none";
        }
        else {
            menu.style.display = "block";
        }
    });
    document.addEventListener("click", function (event) {
        if ((menu === null || menu === void 0 ? void 0 : menu.style.display) === "block" && !(accountButton === null || accountButton === void 0 ? void 0 : accountButton.contains(event.target)) && !menu.contains(event.target)) {
            menu.style.display = "none";
        }
    });
});
