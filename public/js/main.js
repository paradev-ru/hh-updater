function GetMe() {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/me', true);
    xhr.onload = function() {
        if (xhr.status != 200) {
            return
        }
        location = '/logged.html';
    }
    xhr.send();
};

function Logout() {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/logout', true);
    xhr.onload = function() {
        if (xhr.status != 200) {
            return
        }
        location = '/';
    }
    xhr.send();
};

function Delete() {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/delete', true);
    xhr.onload = function() {
        if (xhr.status != 200) {
            return
        }
        location = '/';
    }
    xhr.send();
};
